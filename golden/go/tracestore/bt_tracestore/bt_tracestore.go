// Package bt_tracestore implements a tracestore backed by BigTable
// See BIGTABLE.md for an overview of the schema and design.
package bt_tracestore

import (
	"context"
	"encoding/binary"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/bigtable"
	"go.skia.org/infra/go/bt"
	"go.skia.org/infra/go/metrics2"
	"go.skia.org/infra/go/paramtools"
	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/go/sklog"
	"go.skia.org/infra/go/tiling"
	"go.skia.org/infra/go/util"
	"go.skia.org/infra/go/vcsinfo"
	"go.skia.org/infra/golden/go/tracestore"
	"go.skia.org/infra/golden/go/types"
	"golang.org/x/sync/errgroup"
)

// InitBT initializes the BT instance for the given configuration. It uses the default way
// to get auth information from the environment and must be called with an account that has
// admin rights.
func InitBT(conf BTConfig) error {
	return bt.InitBigtable(conf.ProjectID, conf.InstanceID, conf.TableID, btColumnFamilies)
}

// BTConfig contains the configuration information for the BigTable-based implementation of
// TraceStore.
type BTConfig struct {
	ProjectID  string
	InstanceID string
	TableID    string
	VCS        vcsinfo.VCS
}

// BTTraceStore implements the TraceStore interface.
type BTTraceStore struct {
	vcs    vcsinfo.VCS
	client *bigtable.Client
	table  *bigtable.Table

	tileSize int32
	shards   int32

	// if cacheOps is true, then cache the OrderedParamSets between calls
	// where possible.
	cacheOps bool
	// maps rowName (string) -> *OpsCacheEntry
	opsCache sync.Map

	availIDsMutex sync.Mutex
	availIDs      []digestID
}

// New implements the TraceStore interface backed by BigTable. If cache is true,
// the OrderedParamSets will be cached based on the row name.
func New(ctx context.Context, conf BTConfig, cache bool) (*BTTraceStore, error) {
	client, err := bigtable.NewClient(ctx, conf.ProjectID, conf.InstanceID)
	if err != nil {
		return nil, skerr.Fmt("could not instantiate client: %s", err)
	}

	ret := &BTTraceStore{
		vcs:      conf.VCS,
		client:   client,
		tileSize: DefaultTileSize,
		shards:   DefaultShards,
		table:    client.Open(conf.TableID),
		cacheOps: cache,
		availIDs: []digestID{},
	}
	return ret, nil
}

// Put implements the TraceStore interface.
func (b *BTTraceStore) Put(ctx context.Context, commitHash string, entries []*tracestore.Entry, ts time.Time) error {
	defer metrics2.FuncTimer().Stop()
	// if there are no entries this becomes a no-op.
	if len(entries) == 0 {
		return nil
	}

	// Accumulate all parameters into a paramset and collect all the digests.
	paramSet := make(paramtools.ParamSet, len(entries[0].Params))
	digestSet := make(types.DigestSet, len(entries))
	for _, entry := range entries {
		paramSet.AddParams(entry.Params)
		digestSet[entry.Digest] = true
	}

	repoIndex, err := b.vcs.IndexOf(ctx, commitHash)
	if err != nil {
		return skerr.Fmt("could not look up commit %s: %s", commitHash, err)
	}

	// Find out what tile we need to fetch and what index into that tile we need.
	// Reminder that tileKeys start at 2^32-1 and decrease in value.
	tileKey, commitIndex := b.getTileKey(repoIndex)

	// If these entries have any params we haven't seen before, we need to store those in BigTable.
	ops, err := b.updateOrderedParamSet(ctx, tileKey, paramSet)
	if err != nil {
		sklog.Warningf("Bad paramset: %#v", paramSet)
		return skerr.Fmt("cannot update paramset: %s", err)
	}

	// Similarly, if we have some new digests (almost certainly), we need to update
	// the digestMap with them in there. Of note, we store this
	// map of string (types.Digest) -> int64(DigestId) in big table, then refer to
	// the DigestID elsewhere in the table. DigestIds are essentially a monotonically
	// increasing arbitrary number.
	digestMap, err := b.updateDigestMap(ctx, digestSet)
	if err != nil {
		sklog.Warningf("Bad digestSet: %#v", digestSet)
		return skerr.Fmt("cannot update digest map: %s", err)
	}

	metrics2.GetInt64Metric("gold_digest_map_size").Update(int64(digestMap.Len()))

	if len(digestMap.Delta(digestSet)) != 0 {
		// Should never happen
		return skerr.Fmt("delta should be empty at this point: %v", digestMap.Delta(digestSet))
	}

	// These are two parallel arrays. mutations[i] should be applied to rowNames[i] for all i.
	rowNames, mutations, err := b.createPutMutations(entries, ts, tileKey, commitIndex, ops, digestMap)
	if err != nil {
		return skerr.Fmt("could not create mutations to put data: %s", err)
	}

	// Write the trace data. We pick a batchsize based on the assumption
	// that the whole batch should be 2MB large and each entry is ~200 Bytes of data.
	// 2MB / 200B = 10000. This is extremely conservative but should not be a problem
	// since the batches are written in parallel.
	return b.applyBulkBatched(ctx, rowNames, mutations, 10000)
}

// createPutMutations is a helper function that returns two parallel arrays of
// the rows that need updating and the mutations to apply to those rows.
// Specifically, the mutations will add the given entries to BT, clearing out
// anything that was there previously.
func (b *BTTraceStore) createPutMutations(entries []*tracestore.Entry, ts time.Time, tk tileKey, commitIndex int, ops *paramtools.OrderedParamSet, dm *digestMap) ([]string, []*bigtable.Mutation, error) {
	// These mutations...
	mutations := make([]*bigtable.Mutation, 0, len(entries))
	// .. should be applied to these rows.
	rowNames := make([]string, 0, len(entries))
	btTS := bigtable.Time(ts)
	before := bigtable.Time(ts.Add(-1 * time.Millisecond))

	for _, entry := range entries {
		// To save space, traceID isn't the long form tiling.TraceId
		// (e.g. ,foo=bar,baz=gm,), it's a string of key-value numbers
		// that refer to the params.(e.g. ,0=3,2=18,)
		// See params.paramsEncoder
		sTrace, err := ops.EncodeParamsAsString(entry.Params)
		if err != nil {
			return nil, nil, skerr.Fmt("invalid params: %s", err)
		}
		traceID := encodedTraceID(sTrace)

		rowName := b.calcShardedRowName(tk, typeTrace, string(traceID))
		rowNames = append(rowNames, rowName)

		dID, err := dm.ID(entry.Digest)
		if err != nil {
			// this should never happen, the digest map should know about every digest already.
			return nil, nil, skerr.Fmt("could not fetch id for digest %s: %s", entry.Digest, err)
		}

		// Create a mutation that puts the given digest at the given row
		// (i.e. the trace combined with the tile), at the given column
		// (i.e. the commit offset into this tile).
		mut := bigtable.NewMutation()
		column := strconv.Itoa(commitIndex)
		dBytes, err := dID.MarshalBinary()
		if err != nil {
			// this should never happen, we are just marshalling an int to binary
			return nil, nil, skerr.Fmt("could not encode digest id %d to bytes: %s", dID, err)
		}
		mut.Set(traceFamily, column, btTS, dBytes)
		// Delete anything that existed at this cell before now.
		mut.DeleteTimestampRange(traceFamily, column, 0, before)
		mutations = append(mutations, mut)
	}
	return rowNames, mutations, nil
}

// GetTile implements the TraceStore interface.
// Of note, due to this request possibly spanning over multiple tiles, the ParamsSet may have a
// set of params that does not actually correspond to a trace (this shouldn't be a problem, but is
// worth calling out). For example, suppose a trace with param " device=alpha" abruptly ends on
// tile 4, commit 7 (where the device was removed from testing). If we are on tile 5 and need to
// query both tile 4 starting at commit 10 and tile 5 (the whole thing), we'll just merge the
// paramsets from both tiles, which includes the "device=alpha" params, but they don't exist in
// any traces seen in the tile (since it ended prior to our cutoff point).
func (b *BTTraceStore) GetTile(ctx context.Context, nCommits int) (*tiling.Tile, []*tiling.Commit, error) {
	defer metrics2.FuncTimer().Stop()
	// Look up the commits we need to query from BT
	idxCommits := b.vcs.LastNIndex(nCommits)
	if len(idxCommits) == 0 {
		return nil, nil, skerr.Fmt("No commits found.")
	}

	// These commits could span across multiple tiles, so derive the tiles we need to query.
	c := idxCommits[0]
	startTileKey, startCommitIndex := b.getTileKey(c.Index)

	c = idxCommits[len(idxCommits)-1]
	endTileKey, endCommitIndex := b.getTileKey(c.Index)

	var egroup errgroup.Group

	var commits []*tiling.Commit
	egroup.Go(func() error {
		hashes := make([]string, 0, len(idxCommits))
		for _, ic := range idxCommits {
			hashes = append(hashes, ic.Hash)
		}
		var err error
		commits, err = b.makeTileCommits(ctx, hashes)
		if err != nil {
			return skerr.Fmt("could not load tile commits: %s", err)
		}
		return nil
	})

	var traces traceMap
	var params paramtools.ParamSet
	egroup.Go(func() error {
		var err error
		traces, params, err = b.getTracesInRange(ctx, startTileKey, endTileKey, startCommitIndex, endCommitIndex)

		if err != nil {
			return skerr.Fmt("could not load tile commits: %s", err)
		}
		return nil
	})

	if err := egroup.Wait(); err != nil {
		return nil, nil, skerr.Fmt("could not load last %d commits into tile: %s", nCommits, err)
	}

	ret := &tiling.Tile{
		Traces:   traces,
		ParamSet: params,
		Commits:  commits,
		Scale:    0,
	}

	return ret, commits, nil
}

// getTracesInRange returns a traceMap with data from the given start and stop points (tile and index).
// It also includes the ParamSet for that range.
func (b *BTTraceStore) getTracesInRange(ctx context.Context, startTileKey, endTileKey tileKey, startCommitIndex, endCommitIndex int) (traceMap, paramtools.ParamSet, error) {
	// Query those tiles.
	nTiles := int(startTileKey - endTileKey + 1)
	nCommits := int(startTileKey-endTileKey)*int(b.tileSize) + (endCommitIndex - startCommitIndex) + 1
	encTiles := make([]*encTile, nTiles)
	var egroup errgroup.Group
	tk := startTileKey
	for idx := 0; idx < nTiles; idx++ {
		func(idx int, tk tileKey) {
			egroup.Go(func() error {
				var err error
				encTiles[idx], err = b.loadTile(ctx, tk)
				if err != nil {
					return skerr.Fmt("could not load tile with key %d to index %d: %s", tk, idx, err)
				}
				return nil
			})
		}(idx, tk)
		tk--
	}

	var digestMap *digestMap
	egroup.Go(func() error {
		var err error
		digestMap, err = b.getDigestMap(ctx)
		if err != nil {
			return skerr.Fmt("could not load digestMap: %s", err)
		}
		return nil
	})

	if err := egroup.Wait(); err != nil {
		return nil, nil, skerr.Fmt("could not load %d tiles: %s", nTiles, err)
	}

	// This is the full tile we are going to return.
	tileTraces := make(traceMap, len(encTiles[0].traces))
	paramSet := paramtools.ParamSet{}

	commitIDX := 0
	for idx, encTile := range encTiles {
		// Determine the offset within the tile that we should consider.
		endOffset := int(b.tileSize - 1)
		if idx == (len(encTiles) - 1) {
			// If we are on the last tile, stop early (that is, at endCommitIndex)
			endOffset = endCommitIndex
		}
		segLen := endOffset - startCommitIndex + 1

		for encodedKey, encValues := range encTile.traces {
			// at this point, the encodedKey looks like ,0=1,1=3,3=0,
			// See params.paramsEncoder
			params, err := encTile.ops.DecodeParamsFromString(string(encodedKey))
			if err != nil {
				sklog.Warningf("Incomplete OPS: %#v\n", encTile.ops)
				return nil, nil, skerr.Fmt("corrupted trace key - could not decode %s: %s", encodedKey, err)
			}

			// Turn the params into the tiling.TraceId we expect elsewhere.
			traceKey := tracestore.TraceIDFromParams(params)
			if _, ok := tileTraces[traceKey]; !ok {
				tileTraces[traceKey] = types.NewGoldenTraceN(nCommits)
			}
			gt := tileTraces[traceKey].(*types.GoldenTrace)
			gt.Keys = params
			// Build up the total set of params
			paramSet.AddParams(params)

			// Convert the digests from integer IDs to strings.
			digestIDs := encValues[startCommitIndex : startCommitIndex+segLen]
			digests, err := digestMap.DecodeIDs(digestIDs)
			if err != nil {
				return nil, nil, skerr.Fmt("corrupted digest id - could not decode: %s", err)
			}
			copy(gt.Digests[commitIDX:commitIDX+segLen], digests)
		}

		// After the first tile we always start at the first entry and advance the
		// overall commit index by the segment length.
		commitIDX += segLen
		startCommitIndex = 0
	}

	// Sort the params for determinism.
	paramSet.Normalize()

	return tileTraces, paramSet, nil
}

// GetDenseTile implements the TraceStore interface. It fetches the most recent tile and sees if
// there is enough non-empty data, then queries the next oldest tile until it has nCommits
// non-empty commits.
func (b *BTTraceStore) GetDenseTile(ctx context.Context, nCommits int) (*tiling.Tile, []*tiling.Commit, error) {
	defer metrics2.FuncTimer().Stop()
	// Figure out what index we are on.
	idxCommits := b.vcs.LastNIndex(1)
	if len(idxCommits) == 0 {
		return nil, nil, skerr.Fmt("No commits found.")
	}

	c := idxCommits[0]
	tKey, endIdx := b.getTileKey(c.Index)
	tileStartCommitIdx := c.Index - endIdx

	// commitsWithData is a slice of indexes of commits that have data. These indexes are
	// relative to the repo itself, with index 0 being the first (oldest) commit in the repo.
	commitsWithData := make([]int, 0, nCommits)
	paramSet := paramtools.ParamSet{}
	allTraces := traceMap{}

	// Start at the most recent tile and step backwards until we have enough commits with data.
	for {
		traces, params, err := b.getTracesInRange(ctx, tKey, tKey, 0, endIdx)

		if err != nil {
			return nil, nil, skerr.Fmt("could not load commits from tile %d: %s", tKey, err)
		}

		paramSet.AddParamSet(params)
		// filledCommits are the indexes in the traces that have data.
		// That is, they are the indexes of commits in this tile.
		// It will be sorted from low indexes to high indexes
		filledCommits := traces.CommitIndicesWithData()

		if len(filledCommits)+len(commitsWithData) > nCommits {
			targetLength := nCommits - len(commitsWithData)
			// trim filledCommits so we get to exactly nCommits
			filledCommits = filledCommits[len(filledCommits)-targetLength:]
		}

		for _, tileIdx := range filledCommits {
			commitsWithData = append(commitsWithData, tileStartCommitIdx+tileIdx)
		}
		cTraces := traces.MakeFromCommitIndexes(filledCommits)
		allTraces.PrependTraces(cTraces)

		if len(commitsWithData) >= nCommits || tKey == tileKeyFromIndex(0) {
			break
		}

		tKey++                       // go backwards in time one tile
		endIdx = int(b.tileSize - 1) // fetch the whole previous tile
		tileStartCommitIdx -= int(b.tileSize)
	}

	if len(commitsWithData) == 0 {
		return &tiling.Tile{}, nil, nil
	}
	// put them in oldest to newest order
	sort.Ints(commitsWithData)

	oldestIdx := commitsWithData[0]
	oldestCommit, err := b.vcs.ByIndex(ctx, oldestIdx)
	if err != nil {
		return nil, nil, skerr.Fmt("invalid oldest index %d: %s", oldestIdx, err)
	}
	hashes := b.vcs.From(oldestCommit.Timestamp.Add(-1 * time.Millisecond))

	// There's no guarantee that hashes[0] == oldestCommit[0] (e.g. two commits at same timestamp)
	// So we trim hashes down if necessary
	for i := 0; i < len(hashes); i++ {
		if hashes[i] == oldestCommit.Hash {
			hashes = hashes[i:]
			break
		}
	}

	allCommits, err := b.makeTileCommits(ctx, hashes)
	if err != nil {
		return nil, nil, skerr.Fmt("could not make tile commits: %s", err)
	}

	denseCommits := make([]*tiling.Commit, len(commitsWithData))
	for i, idx := range commitsWithData {
		denseCommits[i] = allCommits[idx-oldestIdx]
	}

	ret := &tiling.Tile{
		Traces:   allTraces,
		ParamSet: paramSet,
		Commits:  denseCommits,
		Scale:    0,
	}
	return ret, allCommits, nil
}

// getTileKey retrieves the tile key and the index of the commit in the given tile (commitIndex)
// given the index of a commit in the repo (repoIndex).
// commitIndex starts at 0 for the oldest commit in the tile.
func (b *BTTraceStore) getTileKey(repoIndex int) (tileKey, int) {
	tileIndex := int32(repoIndex) / b.tileSize
	commitIndex := repoIndex % int(b.tileSize)
	return tileKeyFromIndex(tileIndex), commitIndex
}

// loadTile returns an *encTile corresponding to the tileKey.
func (b *BTTraceStore) loadTile(ctx context.Context, tileKey tileKey) (*encTile, error) {
	defer metrics2.FuncTimer().Stop()
	var egroup errgroup.Group

	// Load the OrderedParamSet so the caller can decode the data from the tile.
	var ops *paramtools.OrderedParamSet
	egroup.Go(func() error {
		opsEntry, _, err := b.getOPS(ctx, tileKey)
		if err != nil {
			return skerr.Fmt("could not load OPS: %s", err)
		}
		ops = opsEntry.ops
		return nil
	})

	var traces map[encodedTraceID][]digestID
	egroup.Go(func() error {
		var err error
		traces, err = b.loadEncodedTraces(ctx, tileKey)
		if err != nil {
			return skerr.Fmt("could not load traces: %s", err)
		}
		return nil
	})

	if err := egroup.Wait(); err != nil {
		return nil, err
	}

	return &encTile{
		ops:    ops,
		traces: traces,
	}, nil
}

// loadEncodedTraces returns all traces belonging to the given tileKey.
// As outlined in BIGTABLE.md, the trace ids and the digest ids they
// map to are in an encoded form and will need to be expanded prior to use.
func (b *BTTraceStore) loadEncodedTraces(ctx context.Context, tileKey tileKey) (map[encodedTraceID][]digestID, error) {
	defer metrics2.FuncTimer().Stop()
	var egroup errgroup.Group
	shardResults := make([]map[encodedTraceID][]digestID, b.shards)
	traceCount := int64(0)

	// Query all shards in parallel.
	for shard := int32(0); shard < b.shards; shard++ {
		func(shard int32) {
			egroup.Go(func() error {
				// This prefix will match all traces belonging to the
				// current shard in the current tile.
				prefixRange := bigtable.PrefixRange(shardedRowName(shard, typeTrace, tileKey, ""))
				target := map[encodedTraceID][]digestID{}
				shardResults[shard] = target
				var parseErr error
				err := b.table.ReadRows(ctx, prefixRange, func(row bigtable.Row) bool {
					// The encoded trace id is the "subkey" part of the row name.
					traceKey := encodedTraceID(extractSubkey(row.Key()))
					// If this is the first time we've seen the trace, initialize the
					// slice of digest ids for it.
					if _, ok := target[traceKey]; !ok {
						target[traceKey] = make([]digestID, b.tileSize)
						atomic.AddInt64(&traceCount, 1)
					}

					for _, col := range row[traceFamily] {
						// The columns are something like T:35 where the part
						// after the colon is the commitIndex i.e. the index
						// of this commit in the current tile.
						idx, err := strconv.Atoi(strings.TrimPrefix(col.Column, traceFamilyPrefix))
						if err != nil {
							// Should never happen
							parseErr = err
							return false
						}
						var dID digestID
						if err := dID.UnmarshalBinary(col.Value); err != nil {
							// This should never happen
							parseErr = err
							return false
						}
						if idx < 0 || idx >= int(b.tileSize) {
							// This would happen if the tile size changed from a past
							// value. It shouldn't be changed, even if the Gold tile size
							// (n_commits) changes.
							parseErr = skerr.Fmt("got index %d that is outside of the target slice of length %d", idx, len(target))
							return false
						}
						target[traceKey][idx] = dID
					}
					return true
				}, bigtable.RowFilter(bigtable.LatestNFilter(1)))
				if err != nil {
					return skerr.Fmt("could not read rows: %s", err)
				}
				return parseErr
			})
		}(shard)
	}

	if err := egroup.Wait(); err != nil {
		return nil, err
	}

	// Merge all the results together
	ret := make(map[encodedTraceID][]digestID, traceCount)
	for _, r := range shardResults {
		for traceKey, digestIDs := range r {
			// different shards should never share results for a tracekey
			// since a trace always maps to the same shard.
			ret[traceKey] = digestIDs
		}
	}

	return ret, nil
}

// applyBulkBatched writes the given rowNames/mutation pairs to BigTable in batches that are
// maximally of size 'batchSize'. The batches are written in parallel.
func (b *BTTraceStore) applyBulkBatched(ctx context.Context, rowNames []string, mutations []*bigtable.Mutation, batchSize int) error {

	var egroup errgroup.Group
	err := util.ChunkIter(len(rowNames), batchSize, func(chunkStart, chunkEnd int) error {
		egroup.Go(func() error {
			tctx, cancel := context.WithTimeout(ctx, writeTimeout)
			defer cancel()
			rowNames := rowNames[chunkStart:chunkEnd]
			mutations := mutations[chunkStart:chunkEnd]
			errs, err := b.table.ApplyBulk(tctx, rowNames, mutations)
			if err != nil {
				return skerr.Fmt("error writing batch [%d:%d]: %s", chunkStart, chunkEnd, err)
			}
			if errs != nil {
				return skerr.Fmt("error writing some portions of batch [%d:%d]: %s", chunkStart, chunkEnd, errs)
			}
			return nil
		})
		return nil
	})
	if err != nil {
		return skerr.Fmt("error running ChunkIter: %s", err)
	}
	return egroup.Wait()
}

// calcShardedRowName deterministically assigns a shard for the given subkey (e.g. traceID)
// Once this is done, the shard, rowtype, tileKey and the subkey are combined into a
// single string to be used as a row name in BT.
func (b *BTTraceStore) calcShardedRowName(tileKey tileKey, rowType, subkey string) string {
	shard := int32(crc32.ChecksumIEEE([]byte(subkey)) % uint32(b.shards))
	return shardedRowName(shard, rowType, tileKey, subkey)
}

// To avoid having one monolithic row, we take the first three characters of the digest
// and use it as a subkey in the row. Then, what remains is used as the column name.
// In practice this means our digests will be split using three hexadecimal characters, so
// we will have 16^3 = 4096 rows for our digest map.
func (b *BTTraceStore) rowAndColNameFromDigest(digest types.Digest) (string, string) {
	subkey := string(digest[:3])
	colName := string(digest[3:])
	return b.calcShardedRowName(digestMapTile, typeDigestMap, subkey), colName
}

// getDigestMap gets the global (i.e. same for all tiles) digestMap.
func (b *BTTraceStore) getDigestMap(ctx context.Context) (*digestMap, error) {
	defer metrics2.FuncTimer().Stop()
	// Query all shards in parallel.
	var egroup errgroup.Group
	shardResults := make([]map[types.Digest]digestID, b.shards)
	total := int64(0)
	for shard := int32(0); shard < b.shards; shard++ {
		func(shard int32) {
			egroup.Go(func() error {
				prefRange := bigtable.PrefixRange(shardedRowName(shard, typeDigestMap, digestMapTile, ""))
				var idx int64
				var parseErr error = nil
				ret := map[types.Digest]digestID{}
				err := b.table.ReadRows(ctx, prefRange, func(row bigtable.Row) bool {
					digestPrefix := extractSubkey(row.Key())
					for _, col := range row[digestMapFamily] {
						idx, parseErr = strconv.ParseInt(string(col.Value), 10, 64)
						if parseErr != nil {
							// Should never happen
							return false
						}
						digest := types.Digest(digestPrefix + strings.TrimPrefix(col.Column, digestMapFamilyPrefix))
						ret[digest] = digestID(idx)
					}
					return true
				}, bigtable.RowFilter(bigtable.LatestNFilter(1)))

				if err != nil {
					return skerr.Fmt("problem fetching shard %d of digestmap: %s", shard, err)
				}
				if parseErr != nil {
					return parseErr
				}

				shardResults[shard] = ret
				atomic.AddInt64(&total, int64(len(ret)))
				return nil
			})
		}(shard)
	}
	if err := egroup.Wait(); err != nil {
		return nil, skerr.Fmt("problem fetching digestmap: %s", err)
	}

	ret := newDigestMap(int(total))
	for _, dm := range shardResults {
		if err := ret.Add(dm); err != nil {
			// put the digest map latter in case it gets truncated
			return nil, skerr.Fmt("could not build DigestMap: %s \nresults %#v", err, dm)
		}
	}
	return ret, nil
}

// getIDs returns a []DigestID of length n where each of the
// digestIDs are unique (even between processes).
func (b *BTTraceStore) getIDs(ctx context.Context, n int) ([]digestID, error) {
	defer metrics2.FuncTimer().Stop()
	// Extract up to n ids from those we have already cached.
	b.availIDsMutex.Lock()
	defer b.availIDsMutex.Unlock()
	toExtract := util.MinInt(len(b.availIDs), n)

	ids := make([]digestID, 0, n)
	ids = append(ids, b.availIDs[:toExtract]...)
	b.availIDs = b.availIDs[toExtract:]

	// missing is how many ids we are short
	missing := int64(n - len(ids))
	if missing == 0 {
		return ids, nil
	}
	// For performance reasons, make a few big requests for ids instead of many small ones.
	// That is, always request numReservedIds extra.
	toRequest := missing + numReservedIds
	// Reserve new IDs via the ID counter
	rmw := bigtable.NewReadModifyWrite()
	rmw.Increment(idCounterFamily, idCounterColumn, toRequest)
	row, err := b.table.ApplyReadModifyWrite(ctx, idCounterRow, rmw)
	if err != nil {
		return nil, skerr.Fmt("could not fetch counter from BT: %s", err)
	}

	// ri are the cells in Row of the given counter family
	// This should be 1 cell belonging to 1 column.
	ri, ok := row[idCounterFamily]
	if !ok {
		// should never happen
		return nil, skerr.Fmt("malformed response - no id counter family: %#v", ri)
	}
	if len(ri) != 1 {
		// should never happen
		return nil, skerr.Fmt("malformed response - expected 1 cell: %#v", ri)
	}

	maxID := digestID(binary.BigEndian.Uint64(ri[0].Value))

	lastID := maxID - digestID(toRequest)
	// ID of 0 is a special case - it's already assigned to MISSING_DIGEST, so skip it.
	if lastID == missingDigestID {
		lastID++
	}
	for i := lastID; i < maxID; i++ {
		// Give the first ids to the current allocation request...
		if missing > 0 {
			ids = append(ids, i)
		} else {
			// ... and put the remainder in the store for later.
			b.availIDs = append(b.availIDs, i)
		}
		missing--
	}

	return ids, nil
}

// returnIDs can be called with a []DigestID of ids that were not actually
// assigned to digests. This allows them to be used by future requests to
// getIDs.
func (b *BTTraceStore) returnIDs(unusedIDs []digestID) {
	b.availIDsMutex.Lock()
	defer b.availIDsMutex.Unlock()
	b.availIDs = append(b.availIDs, unusedIDs...)
}

// getOrAddDigests fills the given digestMap with the given digests
// assigned to a DigestID if they don't already have an assignment.
// This is a helper function for updateDigestMap
// TODO(kjlubick): This currently makes a lot of requests to BT -
// Should there be some caching done here to prevent that?
func (b *BTTraceStore) getOrAddDigests(ctx context.Context, digests []types.Digest, digestMap *digestMap) (*digestMap, error) {
	defer metrics2.FuncTimer().Stop()
	availIDs, err := b.getIDs(ctx, len(digests))
	if err != nil {
		return nil, err
	}

	now := bigtable.Time(time.Now())
	newIDMapping := make(map[types.Digest]digestID, len(digests))
	unusedIDs := make([]digestID, 0, len(availIDs))
	for idx, digest := range digests {
		idVal := availIDs[idx]
		if _, err := digestMap.ID(digest); err == nil {
			// digestMap already has a mapping for this digest, no need to check
			// if BT has seen it yet (because it has).
			// Should never happen because we we've already done this check in updateDigestMap.
			unusedIDs = append(unusedIDs, idVal)
			continue
		}
		rowName, colName := b.rowAndColNameFromDigest(digest)
		// This mutation says "Add an entry to the map for digest -> idVal iff
		// the digest doesn't already have a mapping".
		addMut := bigtable.NewMutation()
		addMut.Set(digestMapFamily, colName, now, []byte(strconv.FormatInt(int64(idVal), 10)))
		filter := bigtable.ColumnFilter(colName)
		// Note that we only add the value if filter is false, i.e. the column does not
		// already exist.
		condMut := bigtable.NewCondMutation(filter, nil, addMut)
		var digestAlreadyHadId bool
		if err := b.table.Apply(ctx, rowName, condMut, bigtable.GetCondMutationResult(&digestAlreadyHadId)); err != nil {
			return nil, skerr.Fmt("could not check if row %s col %s already had a DigestID: %s", rowName, colName, err)
		}

		// We didn't need this ID so let's re-use it later.
		if digestAlreadyHadId {
			unusedIDs = append(unusedIDs, idVal)
		} else {
			newIDMapping[digest] = idVal
		}
	}

	// If all ids were added to BT, then we know our newIDMapping can simply be added
	// to what we already have, since there were no collisions between digests and what
	// was in the table already.
	if len(unusedIDs) == 0 {
		if err := digestMap.Add(newIDMapping); err != nil {
			return nil, err
		}
		return digestMap, nil
	}
	// At this point, some of the digests already had ids, so we should reload
	// the entire digestMap to make sure we have the full picture.
	// TODO(kjlubick): Can we not just add what new ones we saw to what we already have?

	// Return the unused IDs for later use.
	b.returnIDs(unusedIDs)
	return b.getDigestMap(ctx)
}

// updateDigestMap returns the current global DigestMap after making sure the given
// digests are a part of it.
func (b *BTTraceStore) updateDigestMap(ctx context.Context, digests types.DigestSet) (*digestMap, error) {
	defer metrics2.FuncTimer().Stop()
	// Load the digest map from BT.
	// TODO(kjlubick): should we cache this map and first check to see if the digests
	// are all in there?
	digestMap, err := b.getDigestMap(ctx)
	if err != nil {
		return nil, err
	}

	delta := digestMap.Delta(digests)
	if len(delta) == 0 {
		return digestMap, nil
	}

	return b.getOrAddDigests(ctx, delta, digestMap)
}

// Copied from btts.go in infra/perf

// UpdateOrderedParamSet will add all params from 'p' to the OrderedParamSet
// for 'tileKey' and write it back to BigTable.
func (b *BTTraceStore) updateOrderedParamSet(ctx context.Context, tileKey tileKey, p paramtools.ParamSet) (*paramtools.OrderedParamSet, error) {
	defer metrics2.FuncTimer().Stop()

	tctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	var newEntry *opsCacheEntry
	for {
		// Get OPS.
		entry, existsInBT, err := b.getOPS(ctx, tileKey)
		if err != nil {
			return nil, skerr.Fmt("failed to get OPS: %s", err)
		}

		// If the OPS contains our paramset then we're done.
		if delta := entry.ops.Delta(p); len(delta) == 0 {
			return entry.ops, nil
		}

		// Create a new updated ops.
		ops := entry.ops.Copy()
		ops.Update(p)
		newEntry, err = opsCacheEntryFromOPS(ops)
		if err != nil {
			return nil, skerr.Fmt("failed to create cache entry: %s", err)
		}
		encodedOps, err := newEntry.ops.Encode()
		if err != nil {
			return nil, skerr.Fmt("failed to encode new ops: %s", err)
		}

		now := bigtable.Time(time.Now())
		condTrue := false
		if existsInBT {
			// Create an update that avoids the lost update problem.
			cond := bigtable.ChainFilters(
				bigtable.LatestNFilter(1),
				bigtable.FamilyFilter(opsFamily),
				bigtable.ColumnFilter(opsHashColumn),
				bigtable.ValueFilter(string(entry.hash)),
			)
			updateMutation := bigtable.NewMutation()
			updateMutation.Set(opsFamily, opsHashColumn, now, []byte(newEntry.hash))
			updateMutation.Set(opsFamily, opsOpsColumn, now, encodedOps)

			// Add a mutation that cleans up old versions.
			before := bigtable.Time(now.Time().Add(-1 * time.Second))
			updateMutation.DeleteTimestampRange(opsFamily, opsHashColumn, 0, before)
			updateMutation.DeleteTimestampRange(opsFamily, opsOpsColumn, 0, before)
			condUpdate := bigtable.NewCondMutation(cond, updateMutation, nil)

			if err := b.table.Apply(tctx, tileKey.OpsRowName(), condUpdate, bigtable.GetCondMutationResult(&condTrue)); err != nil {
				sklog.Warningf("Failed to apply: %s", err)
				return nil, err
			}

			// If !condTrue then we need to try again,
			// and clear our local cache.
			if !condTrue {
				sklog.Warningf("Exists !condTrue - clearing cache and trying again.")
				b.opsCache.Delete(tileKey.OpsRowName())
				continue
			}
		} else {
			// Create an update that only works if the ops entry doesn't exist yet.
			// I.e. only apply the mutation if the HASH column doesn't exist for this row.
			cond := bigtable.ChainFilters(
				bigtable.FamilyFilter(opsFamily),
				bigtable.ColumnFilter(opsHashColumn),
			)
			updateMutation := bigtable.NewMutation()
			updateMutation.Set(opsFamily, opsHashColumn, now, []byte(newEntry.hash))
			updateMutation.Set(opsFamily, opsOpsColumn, now, encodedOps)

			condUpdate := bigtable.NewCondMutation(cond, nil, updateMutation)
			if err := b.table.Apply(tctx, tileKey.OpsRowName(), condUpdate, bigtable.GetCondMutationResult(&condTrue)); err != nil {
				sklog.Warningf("Failed to apply: %s", err)
				// clear cache and try again
				b.opsCache.Delete(tileKey.OpsRowName())
				continue
			}

			// If condTrue then we need to try again,
			// and clear our local cache.
			if condTrue {
				sklog.Warningf("First Write condTrue - clearing cache and trying again.")
				b.opsCache.Delete(tileKey.OpsRowName())
				continue
			}
		}

		// Successfully wrote OPS, so update the cache.
		if b.cacheOps {
			b.opsCache.Store(tileKey.OpsRowName(), newEntry)
		}
		break
	}
	return newEntry.ops, nil
}

// getOps returns the OpsCacheEntry for a given tile.
//
// Note that it will create a new OpsCacheEntry if none exists.
//
// getOps returns false if the OPS in BT was empty, true otherwise (even if cached).
func (b *BTTraceStore) getOPS(ctx context.Context, tileKey tileKey) (*opsCacheEntry, bool, error) {
	defer metrics2.FuncTimer().Stop()
	if b.cacheOps {
		entry, ok := b.opsCache.Load(tileKey.OpsRowName())
		if ok {
			return entry.(*opsCacheEntry), true, nil
		}
	}
	tctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()
	row, err := b.table.ReadRow(tctx, tileKey.OpsRowName(), bigtable.RowFilter(bigtable.LatestNFilter(1)))
	if err != nil {
		return nil, false, skerr.Fmt("failed to read OPS from BigTable for %s: %s", tileKey.OpsRowName(), err)
	}
	// If there is no entry in BigTable then return an empty OPS.
	if len(row) == 0 {
		sklog.Warningf("Failed to read OPS from BT for %s.", tileKey.OpsRowName())
		entry, err := newOpsCacheEntry()
		return entry, false, err
	}
	entry, err := newOpsCacheEntryFromRow(row)
	if err == nil && b.cacheOps {
		b.opsCache.Store(tileKey.OpsRowName(), entry)
	}
	return entry, true, err
}

// makeTileCommits creates a slice of tiling.Commit from the given git hashes.
// Specifically, we need to look up the details to get the author information.
func (b *BTTraceStore) makeTileCommits(ctx context.Context, hashes []string) ([]*tiling.Commit, error) {
	longCommits, err := b.vcs.DetailsMulti(ctx, hashes, false)
	if err != nil {
		// put hashes second in case they get truncated for being quite long.
		return nil, skerr.Fmt("could not fetch commit data for commits %s (hashes: %q)", err, hashes)
	}

	commits := make([]*tiling.Commit, len(hashes))
	for i, lc := range longCommits {
		if lc == nil {
			return nil, skerr.Fmt("commit %s not found from VCS", hashes[i])
		}
		commits[i] = &tiling.Commit{
			Hash:       lc.Hash,
			Author:     lc.Author,
			CommitTime: lc.Timestamp.Unix(),
		}
	}
	return commits, nil
}

// Make sure BTTraceStore fulfills the TraceStore Interface
var _ tracestore.TraceStore = (*BTTraceStore)(nil)