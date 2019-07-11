package ds_tryjobstore

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync/atomic"

	"cloud.google.com/go/datastore"
	"go.skia.org/infra/go/ds"
	"go.skia.org/infra/go/eventbus"
	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/go/sklog"
	"go.skia.org/infra/go/util"
	"go.skia.org/infra/golden/go/tryjobstore"
	"go.skia.org/infra/golden/go/types"
	"golang.org/x/sync/errgroup"
)

const (
	// batchsize is the maximal size of entities processed in a single batch. This
	// is to stay below the limit of 500 and empirically keeps transactions small
	// enough to stay below 10MB limit.
	batchSize = 300
)

// DSTryjobStore implements the TryjobStore interface on top of cloud datastore.
type DSTryjobStore struct {
	client   *datastore.Client
	eventBus eventbus.EventBus
}

// New creates a new instance of TryjobStore based on Datastore.
func New(client *datastore.Client, eventBus eventbus.EventBus) (*DSTryjobStore, error) {
	if client == nil {
		return nil, skerr.Fmt("received nil for datastore client")
	}

	if eventBus == nil {
		return nil, skerr.Fmt("received nil for eventbus")
	}

	return &DSTryjobStore{
		client:   client,
		eventBus: eventBus,
	}, nil
}

// ListIssues implements the TryjobStore interface.
func (c *DSTryjobStore) ListIssues(offset, size int) ([]*tryjobstore.Issue, int, error) {
	ctx := context.TODO()
	query := ds.NewQuery(ds.ISSUE).KeysOnly().Order("-Updated")

	keys, err := c.client.GetAll(ctx, query, nil)
	if err != nil {
		return nil, 0, err
	}

	total := len(keys)
	start := util.MinInt(total, offset)
	end := util.MinInt(start+size, total)
	targetKeys := keys[start:end]

	if len(targetKeys) == 0 {
		return []*tryjobstore.Issue{}, total, nil
	}

	// Fetch the entities.
	ret := make([]*tryjobstore.Issue, len(targetKeys))
	if err := c.client.GetMulti(ctx, targetKeys, ret); err != nil {
		return nil, 0, err
	}
	return ret, total, nil
}

// GetIssue implements the TryjobStore interface.
func (c *DSTryjobStore) GetIssue(issueID int64, loadTryjobs bool) (*tryjobstore.Issue, error) {
	target := &tryjobstore.Issue{}
	key := c.getIssueKey(issueID)
	ok, err := c.getEntity(key, target, nil)
	if err != nil {
		return nil, sklog.FmtErrorf("Error in getEntity for %s: %s", key, err)
	}

	if !ok {
		return nil, nil
	}

	if loadTryjobs {
		_, tryjobsMap, err := c.getTryjobsForIssue(issueID, nil, false, true)
		if err != nil {
			return nil, sklog.FmtErrorf("Error getting tryjobs for issue: %s", err)
		}

		for patchsetID, tryjobs := range tryjobsMap {
			ps := target.FindPatchset(patchsetID)
			if ps == nil {
				return nil, sklog.FmtErrorf("Unable to find patchset %d in issue %d:", patchsetID, target.ID)
			}
			ps.Tryjobs = tryjobs
		}
	}

	return target, nil
}

// UpdateIssue implements the TryjobStore interface.
func (c *DSTryjobStore) UpdateIssue(details *tryjobstore.Issue, updateFn tryjobstore.NewValueFn) error {
	_, err := c.updateEntity(c.getIssueKey(details.ID), details, nil, false, updateFn)
	return err
}

// CommitIssueExp implements the TryjobStore interface.
func (c *DSTryjobStore) CommitIssueExp(issueID int64, commitFn func() error) error {
	// setCommittedFn is executed in a transaction below
	setCommittedFn := func(tx *datastore.Transaction) error {
		issue := &tryjobstore.Issue{}
		key := c.getIssueKey(issueID)
		ok, err := c.getEntity(key, issue, nil)
		if err != nil {
			return skerr.Fmt("Error in getEntity for %s: %s", key, err)
		}

		if !ok {
			return skerr.Fmt("Unable to find issue %d.", issueID)
		}

		// If this is already committed then we are done.
		if issue.Committed {
			return nil
		}

		// Execute the commit function to commit the actual expectations.
		if err := commitFn(); err != nil {
			sklog.Warningf("Ran into error committing the expectations: %s", err)
			return err
		}

		issue.Committed = true
		_, err = c.updateEntity(key, issue, tx, true, nil)
		return err
	}

	_, err := c.client.RunInTransaction(context.TODO(), setCommittedFn)
	return err
}

// DeleteIssue implements the TryjobStore interface.
func (c *DSTryjobStore) DeleteIssue(issueID int64) error {
	ctx := context.TODO()

	// Delete any tryjobs that are still there.
	if err := c.deleteTryjobsForIssue(issueID); err != nil {
		return skerr.Fmt("could not delete tryjobs for issue %d: %s", issueID, err)
	}

	key := c.getIssueKey(issueID)
	// Delete the entity.
	return c.client.Delete(ctx, key)
}

// GetTryjob implements the TryjobStore interface.
func (c *DSTryjobStore) GetTryjob(issueID, buildBucketID int64) (*tryjobstore.Tryjob, error) {
	ret := &tryjobstore.Tryjob{}
	if err := c.client.Get(context.TODO(), c.getTryjobKey(buildBucketID), ret); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, nil
		}
		return nil, err
	}

	return ret, nil
}

// UpdateTryjob implements the TryjobStore interface.
func (c *DSTryjobStore) UpdateTryjob(buildBucketID int64, tryjob *tryjobstore.Tryjob, newValFn tryjobstore.NewValueFn) error {
	// If this is an update that needs to call the newVal function, then set the parameters for updateEntity right.
	if tryjob == nil {
		// make sure we have the necessary information if there are no data to be written directly.
		if (buildBucketID == 0) || (newValFn == nil) {
			return skerr.Fmt("ID and newValFn cannot be nil when no tryjob is provided. Update not possible")
		}
		tryjob = &tryjobstore.Tryjob{}
	} else {
		// signal to updateEntity that this not in a transaction. Just in case.
		newValFn = nil
		buildBucketID = tryjob.BuildBucketID
	}

	newTryjob, err := c.updateEntity(c.getTryjobKey(buildBucketID), tryjob, nil, false, newValFn)
	if err != nil {
		return err
	}
	c.eventBus.Publish(tryjobstore.EV_TRYJOB_UPDATED, copyTryjob(newTryjob.(*tryjobstore.Tryjob)), true)
	return nil
}

func copyTryjob(t *tryjobstore.Tryjob) *tryjobstore.Tryjob {
	ret := &tryjobstore.Tryjob{}
	*ret = *t
	return ret
}

// GetTryjobs implements the TryjobStore interface.
func (c *DSTryjobStore) GetTryjobs(issueID int64, patchsetIDs []int64, filterDup bool, loadResults bool) ([]*tryjobstore.Tryjob, [][]*tryjobstore.TryjobResult, error) {
	flatTryjobKeys, tryjobsMap, err := c.getTryjobsForIssue(issueID, patchsetIDs, false, filterDup)
	if err != nil {
		return nil, nil, skerr.Fmt("could not locate tryjobs for issue %d: %s", issueID, err)
	}

	// Flatten the Tryjobs map and make sure the element in keys matches.
	tryjobs := make([]*tryjobstore.Tryjob, 0, len(flatTryjobKeys))
	for _, tjs := range tryjobsMap {
		tryjobs = append(tryjobs, tjs...)
	}

	sort.Slice(tryjobs, func(i, j int) bool {
		return tryjobs[i].Builder < tryjobs[j].Builder
	})

	var results [][]*tryjobstore.TryjobResult
	if loadResults {
		var err error
		results, err = c.GetTryjobResults(tryjobs)
		if err != nil {
			return nil, nil, skerr.Fmt("could not get tryjob results for issue %d: %s", issueID, err)
		}
	}

	return tryjobs, results, nil
}

// RunningTryjobs implements the TryjobStore interface.
func (c *DSTryjobStore) RunningTryjobs() ([]*tryjobstore.Tryjob, error) {
	query := ds.NewQuery(ds.TRYJOB).
		Filter("Status <", int(tryjobstore.TRYJOB_COMPLETE))

	tryjobs := []*tryjobstore.Tryjob{}
	_, err := c.client.GetAll(context.TODO(), query, &tryjobs)
	if err != nil {
		return nil, sklog.FmtErrorf("Error making GetAll call: %s", err)
	}
	return tryjobs, nil
}

// GetTryjobResults implements the TryjobStore interface.
func (c *DSTryjobStore) GetTryjobResults(tryjobs []*tryjobstore.Tryjob) ([][]*tryjobstore.TryjobResult, error) {
	tryjobKeys := make([]*datastore.Key, 0, len(tryjobs))
	for _, tryjob := range tryjobs {
		tryjobKeys = append(tryjobKeys, tryjob.Key)
	}

	_, tryjobResults, err := c.getResultsForTryjobs(tryjobKeys, false)
	if err != nil {
		return nil, err
	}

	return tryjobResults, nil
}

// UpdateTryjobResult implements the TryjobStore interface.
func (c *DSTryjobStore) UpdateTryjobResult(results []*tryjobstore.TryjobResult) error {
	keys := make([]*datastore.Key, 0, len(results))
	uniqueEntries := util.StringSet{}
	for _, result := range results {
		// Make sure that tests are not bunched together.
		if len(result.Params[types.PRIMARY_KEY_FIELD]) != 1 {
			return fmt.Errorf("Parameter value for primary key field '%s' must exactly contain one value. Found: %v", types.PRIMARY_KEY_FIELD, result.Params[types.PRIMARY_KEY_FIELD])
		}
		keys = append(keys, c.getTryjobResultKey())
		uniqueEntries[string(result.TestName)+string(result.Digest)] = true
	}

	if len(uniqueEntries) != len(keys) {
		return fmt.Errorf("All (test,digest) pairs must be unique when adding tryjob results.")
	}

	for i := 0; i < len(keys); i += batchSize {
		endIdx := util.MinInt(i+batchSize, len(keys))
		if _, err := c.client.PutMulti(context.TODO(), keys[i:endIdx], results[i:endIdx]); err != nil {
			return err
		}
	}
	return nil
}

// deleteTryjobsForIssue deletes all tryjob information for the given issue.
func (c *DSTryjobStore) deleteTryjobsForIssue(issueID int64) error {
	// Get all the tryjob keys.
	tryjobKeys, _, err := c.getTryjobsForIssue(issueID, nil, true, false)
	if err != nil {
		return fmt.Errorf("Error retrieving tryjob keys: %s", err)
	}

	ctx := context.TODO()
	// Delete all results of the tryjobs.
	tryjobResultKeys, _, err := c.getResultsForTryjobs(tryjobKeys, true)
	if err != nil {
		return err
	}

	for _, keys := range tryjobResultKeys {
		// Break the keys down in batches.
		for i := 0; i < len(keys); i += batchSize {
			currBatch := keys[i:util.MinInt(i+batchSize, len(keys))]
			if err := c.client.DeleteMulti(ctx, currBatch); err != nil {
				return fmt.Errorf("Error deleting tryjob results: %s", err)
			}
		}
	}

	// Delete the tryjobs themselves.
	if err := c.client.DeleteMulti(ctx, tryjobKeys); err != nil {
		return fmt.Errorf("Error deleting %d tryjobs for issue %d: %s", len(tryjobKeys), issueID, err)
	}
	return nil
}

// getResultsForTryjobs returns the test results for the given tryjobs.
func (c *DSTryjobStore) getResultsForTryjobs(tryjobKeys []*datastore.Key, keysOnly bool) ([][]*datastore.Key, [][]*tryjobstore.TryjobResult, error) {
	// Collect all results across tryjobs.
	n := len(tryjobKeys)
	tryjobResultKeys := make([][]*datastore.Key, n)
	var tryjobResults [][]*tryjobstore.TryjobResult = nil
	if !keysOnly {
		tryjobResults = make([][]*tryjobstore.TryjobResult, n)
	}

	// Get there keys and results.
	ctx := context.TODO()
	var egroup errgroup.Group
	for idx, key := range tryjobKeys {
		func(idx int, key *datastore.Key) {
			egroup.Go(func() error {
				query := ds.NewQuery(ds.TRYJOB_RESULT).Filter("BuildBucketID =", key.ID)
				if keysOnly {
					query = query.KeysOnly()
				}

				queryResult := []*tryjobstore.TryjobResult{}
				var err error
				if tryjobResultKeys[idx], err = c.client.GetAll(ctx, query, &queryResult); err != nil {
					return err
				}

				if !keysOnly {
					tryjobResults[idx] = queryResult
				}
				return nil
			})
		}(idx, key)
	}

	if err := egroup.Wait(); err != nil {
		return nil, nil, fmt.Errorf("Error getting tryjob results: %s", err)
	}

	return tryjobResultKeys, tryjobResults, nil
}

// getTryjobsForIssue is a utility function that retrieves the Tryjobs for a given
// issue and list of patchsets. If keysOnly is true only the keys of the Tryjobs will
// be returned. If filterDup is true duplicate Tryjobs will be filtered out for
// each patchset, only keeping the newest.
// The first return value is the unordered slice of keys of the Tryjobs.
// The second return value groups the tryjobs for each patchset as a map[patch_set_id][]*tryjobstore.Tryjob.
func (c *DSTryjobStore) getTryjobsForIssue(issueID int64, patchsetIDs []int64, keysOnly bool, filterDup bool) ([]*datastore.Key, map[int64][]*tryjobstore.Tryjob, error) {
	if keysOnly && filterDup {
		return nil, nil, sklog.FmtErrorf("filterDup cannot be true when keysOnly is true, since Tryjob is necessary for filtering")
	}

	if len(patchsetIDs) == 0 {
		patchsetIDs = []int64{-1}
	} else {
		sort.Slice(patchsetIDs, func(i, j int) bool { return patchsetIDs[i] < patchsetIDs[j] })
	}

	n := len(patchsetIDs)
	keysArr := make([][]*datastore.Key, n)
	valsArr := make([][]*tryjobstore.Tryjob, n)
	resultSize := int32(0)
	var egroup errgroup.Group
	for idx, patchsetID := range patchsetIDs {
		func(idx int, patchsetID int64) {
			egroup.Go(func() error {
				query := ds.NewQuery(ds.TRYJOB).
					Filter("IssueID =", issueID)
				if patchsetID > 0 {
					query = query.Filter("PatchsetID =", patchsetID)
				}

				var tryjobs []*tryjobstore.Tryjob = nil
				if keysOnly {
					query = query.KeysOnly()
				}

				keys, err := c.client.GetAll(context.TODO(), query, &tryjobs)
				if err != nil {
					return fmt.Errorf("Error making GetAll call: %s", err)
				}
				keysArr[idx] = keys
				valsArr[idx] = tryjobs
				atomic.AddInt32(&resultSize, int32(len(keys)))
				return nil
			})
		}(idx, patchsetID)
	}

	if err := egroup.Wait(); err != nil {
		return nil, nil, err
	}

	// Assemble the flat array of keys.
	retKeys := make([]*datastore.Key, 0, resultSize)
	for _, keys := range keysArr {
		retKeys = append(retKeys, keys...)
	}

	// If all we want are the keys we are done.
	if keysOnly {
		return retKeys, nil, nil
	}

	// Group the tryjobs by their patchsets
	tryjobsMap := make(map[int64][]*tryjobstore.Tryjob, resultSize)
	for _, tryjobs := range valsArr {
		for _, tj := range tryjobs {
			tryjobsMap[tj.PatchsetID] = append(tryjobsMap[tj.PatchsetID], tj)
		}
	}

	// Go through the patchsets and dedupe tryjobs for each.
	if filterDup {
		for patchsetID, tryjobs := range tryjobsMap {
			// sort when they were last updated
			sort.Slice(tryjobs, func(i, j int) bool { return tryjobs[i].Updated.Before(tryjobs[j].Updated) })

			// Iterate the builders in reverse order and filter out duplicate builders.
			builders := util.StringSet{}
			for i := len(tryjobs) - 1; i >= 0; i-- {
				if builders[tryjobs[i].Builder] {
					copy(tryjobs[i:], tryjobs[i+1:])
					tryjobs[len(tryjobs)-1] = nil
					tryjobs = tryjobs[:len(tryjobs)-1]
				} else {
					builders[tryjobs[i].Builder] = true
				}
			}

			// NOTE: Store the new slice back into the tryjobs map since we might have removed some values.
			tryjobsMap[patchsetID] = tryjobs
		}
	}

	return retKeys, tryjobsMap, nil
}

// updateEntity writes the given entity to the datastore. If the
// newValFn is not nil an error will be returned if the entity does not exist.
// The non-nil return value of newValFn will be written to the data store.
// If newValFn returns nil nothing will be written to the datastore.
// If newValFn is nil the entity will be written to the datastore if either does
// not exist yet or is newer than the existing entity.
func (c *DSTryjobStore) updateEntity(key *datastore.Key, item newerInterface, tx *datastore.Transaction, force bool, newValFn tryjobstore.NewValueFn) (interface{}, error) {
	// Update the issue if the provided one is newer.
	updateFn := func(tx *datastore.Transaction) error {
		curr := reflect.New(reflect.TypeOf(item).Elem()).Interface()
		ok, err := c.getEntity(key, curr, tx)
		if err != nil {
			return err
		}

		// If this is an in-transaction update then get the new value from the current one.
		if newValFn != nil {
			if ok {
				newVal := newValFn(curr)

				// No update required. We are done.
				if newVal == nil {
					return nil
				}
				item = newVal.(newerInterface)
			} else {
				return sklog.FmtErrorf("Unable to find item %s for transactional update.", key)
			}
		} else if ok && !force && !item.MoreRecentThan(curr) {
			// We found an item that is not more recent than the current one: Nothing to do.
			return nil
		}

		_, err = tx.Put(key, item)
		return err
	}

	// Run the transaction.
	var err error
	if tx == nil {
		_, err = c.client.RunInTransaction(context.TODO(), updateFn)
	} else {
		err = updateFn(tx)
	}

	return item, err
}

// getEntity loads the entity defined by key into target. If tx is not nil it uses the transaction.
func (c *DSTryjobStore) getEntity(key *datastore.Key, target interface{}, tx *datastore.Transaction) (bool, error) {
	var err error
	if tx == nil {
		err = c.client.Get(context.TODO(), key, target)
	} else {
		err = tx.Get(key, target)
	}

	if err != nil {
		// If we couldn't find it return nil, but no error.
		if err == datastore.ErrNoSuchEntity {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// getIssueKey returns a datastore key for the given issue id.
func (c *DSTryjobStore) getIssueKey(id int64) *datastore.Key {
	ret := ds.NewKey(ds.ISSUE)
	ret.ID = id
	return ret
}

// getTryjobKey returns a datastore key for the given buildbucketID.
func (c *DSTryjobStore) getTryjobKey(buildBucketID int64) *datastore.Key {
	ret := ds.NewKey(ds.TRYJOB)
	ret.ID = buildBucketID
	return ret
}

// getTryjobResultKey returns a key for the given tryjobResult.
func (c *DSTryjobStore) getTryjobResultKey() *datastore.Key {
	return ds.NewKey(ds.TRYJOB_RESULT)
}

// Make sure DSTryjobStore fulfills the TryjobStore interface
var _ tryjobstore.TryjobStore = (*DSTryjobStore)(nil)
