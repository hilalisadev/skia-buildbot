package summary

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.skia.org/infra/go/testutils/unittest"
	"go.skia.org/infra/go/tiling"
	"go.skia.org/infra/golden/go/blame"
	"go.skia.org/infra/golden/go/digest_counter"
	bug_revert "go.skia.org/infra/golden/go/testutils/data_bug_revert"
	"go.skia.org/infra/golden/go/types"
	"go.skia.org/infra/golden/go/types/expectations"
)

/**
  Test outline

  We create the following trace data:
  Traces
  ------
  id   | config  | test name  | corpus(source_type) |  digests
  a      8888      test_first         gm              aaa+, bbb-
  b      565       test_first         gm              ccc?, ddd?
  c      gpu       test_first         gm              eee+
  d      8888      test_second        gm              fff-, ggg?
  e      8888      test_third         image           jjj?

  Expectations
  ------------
  test_first   aaa  pos
  test_first   bbb  neg
  test_first   ccc  unt
  test_first   ddd  unt
  test_first   eee  pos
  test_second  fff  neg

  Ignores
  -------
  config=565

  Note no entry for test_third or ggg, meaning untriaged.

  Then, we test the following conditions and make sure we get
  the expected test summaries.

  source_type=gm
    test_first - pos(aaa, eee):2  neg(bbb):1
    test_second -                 neg(fff):1   unt(ggg):1

  source_type=gm includeIgnores=true
    test_first - pos(aaa, eee):2  neg(bbb):1   unt(ccc, ddd):2
    test_second -                 neg(fff):1   unt(ggg):1

  source_type=gm includeIgnores=true testName=test_first
    test_first - pos(aaa, eee):2  neg(bbb):1   unt(ccc, ddd):2

  testname = test_first
    test_first - pos(aaa, eee):2  neg(bbb):1

  testname = test_third
    test_third -                  unt(jjj):1

  config=565&config=8888
    test_first - pos(aaa):1       neg(bbb):1
    test_second -                 neg(fff):1   unt(ggg):1
    test_third -                  unt(jjj):1

  config=565&config=8888 head=true
    test_first -                  neg(bbb):1
    test_second -                 unt(ggg):1
    test_third -                  unt(jjj):1

  config=gpu
    test_first -                  pos(eee):1

  config=unknown
    <empty>

*/

func TestSummaryMap_AllGMsWithIgnores(t *testing.T) {
	unittest.SmallTest(t)

	sum := computeHelper(t, makeTileWithIgnores(), nil, url.Values{types.CORPUS_FIELD: {"gm"}}, false)
	require.Len(t, sum, 2)
	s1 := find(sum, FirstTest)
	require.NotNil(t, s1)
	assert.Empty(t, s1.UntHashes)
	// The only 2 untriaged digests for this test ignored because they were 565
	triageCountsCorrect(t, s1, 2, 1, 0)

	s2 := find(sum, SecondTest)
	require.NotNil(t, s2)
	assert.Equal(t, types.DigestSlice{"ggg"}, s2.UntHashes)
	triageCountsCorrect(t, s2, 0, 1, 1)
	// no gms for ThirdTest
}

func TestSummaryMap_AllGMsFullTile(t *testing.T) {
	unittest.SmallTest(t)

	sum := computeHelper(t, makeFullTile(), nil, url.Values{types.CORPUS_FIELD: {"gm"}}, false)
	require.Len(t, sum, 2)
	s1 := find(sum, FirstTest)
	require.NotNil(t, s1)
	assert.Equal(t, types.DigestSlice{"ccc", "ddd"}, s1.UntHashes)
	triageCountsCorrect(t, s1, 2, 1, 2)

	s2 := find(sum, SecondTest)
	require.NotNil(t, s2)
	assert.Equal(t, types.DigestSlice{"ggg"}, s2.UntHashes)
	triageCountsCorrect(t, s2, 0, 1, 1)
	// no gms for ThirdTest
}

func TestSummaryMap_FirstTestFullTile(t *testing.T) {
	unittest.SmallTest(t)

	sum := computeHelper(t, makeFullTile(), types.TestNameSet{FirstTest: true}, url.Values{types.CORPUS_FIELD: {"gm"}}, false)
	require.Len(t, sum, 1)
	s1 := find(sum, FirstTest)
	require.NotNil(t, s1)
	assert.Equal(t, types.DigestSlice{"ccc", "ddd"}, s1.UntHashes)
	triageCountsCorrect(t, s1, 2, 1, 2)
	// Only FirstTest results should be here, rest are known to be nil since map length == 1.
}

func TestSummaryMap_FirstTestIgnores(t *testing.T) {
	unittest.SmallTest(t)

	sum := computeHelper(t, makeTileWithIgnores(), types.TestNameSet{FirstTest: true}, nil, false)
	require.Len(t, sum, 1)
	s1 := find(sum, FirstTest)
	require.NotNil(t, s1)
	assert.Empty(t, s1.UntHashes)
	triageCountsCorrect(t, s1, 2, 1, 0)
	// Only FirstTest results should be here, rest are known to be nil since map length == 1.
}

func TestSummaryMap_8888Or565Ignores(t *testing.T) {
	unittest.SmallTest(t)

	sum := computeHelper(t, makeTileWithIgnores(), nil, url.Values{"config": {"8888", "565"}}, false)
	require.Len(t, sum, 3)
	s1 := find(sum, FirstTest)
	require.NotNil(t, s1)
	assert.Empty(t, s1.UntHashes)
	// Even though we queried for the 565, the untriaged ones won't show up because of ignores.
	triageCountsCorrect(t, s1, 1, 1, 0)

	s2 := find(sum, SecondTest)
	require.NotNil(t, s2)
	assert.Equal(t, types.DigestSlice{"ggg"}, s2.UntHashes)
	triageCountsCorrect(t, s2, 0, 1, 1)

	s3 := find(sum, ThirdTest)
	assert.NotNil(t, s3)
	assert.Equal(t, types.DigestSlice{"jjj"}, s3.UntHashes)
	triageCountsCorrect(t, s3, 0, 0, 1)
}

func TestSummaryMap_8888Or565IgnoresHead(t *testing.T) {
	unittest.SmallTest(t)

	sum := computeHelper(t, makeTileWithIgnores(), nil, url.Values{"config": {"8888", "565"}}, true)
	require.Len(t, sum, 3)
	s1 := find(sum, FirstTest)
	require.NotNil(t, s1)
	assert.Empty(t, s1.UntHashes)
	// These numbers are are a bit lower because we are only looking at head.
	// Those with missing digests should "pull forward" their last result (see ThirdTest)
	triageCountsCorrect(t, s1, 0, 1, 0)

	s2 := find(sum, SecondTest)
	require.NotNil(t, s2)
	assert.Equal(t, types.DigestSlice{"ggg"}, s2.UntHashes)
	triageCountsCorrect(t, s2, 0, 0, 1)

	s3 := find(sum, ThirdTest)
	assert.NotNil(t, s3)
	assert.Equal(t, types.DigestSlice{"jjj"}, s3.UntHashes)
	triageCountsCorrect(t, s3, 0, 0, 1)
}

func TestSummaryMap_GPUConfigIgnores(t *testing.T) {
	unittest.SmallTest(t)

	sum := computeHelper(t, makeTileWithIgnores(), nil, url.Values{"config": {"gpu"}}, false)
	require.Len(t, sum, 1)
	s1 := find(sum, FirstTest)
	require.NotNil(t, s1)
	assert.Empty(t, s1.UntHashes)
	// Only one digest should be found, and it is positive.
	triageCountsCorrect(t, s1, 1, 0, 0)
	// No other tests have gpu config
}

func TestSummaryMap_UnknownConfigIgnores(t *testing.T) {
	unittest.SmallTest(t)

	sum := computeHelper(t, makeTileWithIgnores(), nil, url.Values{"config": {"unknown"}}, false)
	require.Empty(t, sum)
}

func computeHelper(t *testing.T, tile *tiling.Tile, testNames types.TestNameSet, query url.Values, head bool) []*TriageStatus {
	dc := digest_counter.New(makeFullTile())
	blamer, err := blame.New(makeFullTile(), makeExpectations())
	require.NoError(t, err)

	tr := asSlice(t, tile.Traces)

	d := Data{
		Traces:       tr,
		Expectations: makeExpectations(),
		ByTrace:      dc.ByTrace(),
		Blamer:       blamer,

		DiffStore: nil, // diameter is disabled, so this can be nil.
	}

	return d.Calculate(testNames, query, head)
}

func asSlice(t *testing.T, traces map[tiling.TraceID]tiling.Trace) []*types.TracePair {
	xt := make([]*types.TracePair, 0, len(traces))
	for id, tr := range traces {
		gt, ok := tr.(*types.GoldenTrace)
		require.True(t, ok)
		xt = append(xt, &types.TracePair{
			ID:    id,
			Trace: gt,
		})
	}
	return xt
}

// TestSummaryMap_FullBugRevert checks the entire return value, rather
// than just the spot checks above.
func TestSummaryMap_FullBugRevert(t *testing.T) {
	unittest.SmallTest(t)

	sum := bugRevertHelper(t, url.Values{types.CORPUS_FIELD: {"gm"}}, false)
	require.Equal(t, []*TriageStatus{
		{
			Name:      bug_revert.TestOne,
			Pos:       1,
			Untriaged: 1,
			UntHashes: types.DigestSlice{bug_revert.UntriagedDigestBravo},
			Num:       2,
			Corpus:    "gm",
			Blame: []blame.WeightedBlame{
				{
					Author: bug_revert.BuggyAuthor,
					Prob:   1,
				},
			},
		},
		{
			Name:      bug_revert.TestTwo,
			Pos:       2,
			Untriaged: 2,
			UntHashes: types.DigestSlice{bug_revert.UntriagedDigestDelta, bug_revert.UntriagedDigestFoxtrot},
			Num:       4,
			Corpus:    "gm",
			Blame: []blame.WeightedBlame{
				{
					Author: bug_revert.InnocentAuthor,
					Prob:   0.5,
				},
				{
					Author: bug_revert.BuggyAuthor,
					Prob:   0.5,
				},
			},
		},
	}, sum)
}

// TestSummaryMap_FullBugRevertHead checks the entire return value of a query at head, rather
// than just the spot checks above.
func TestSummaryMap_FullBugRevertHead(t *testing.T) {
	unittest.SmallTest(t)

	sum := bugRevertHelper(t, url.Values{types.CORPUS_FIELD: {"gm"}}, true)
	require.Equal(t, []*TriageStatus{
		{
			Name:      bug_revert.TestOne,
			Pos:       1,
			Untriaged: 0,
			UntHashes: types.DigestSlice{},
			Num:       1,
			Corpus:    "gm",
			// TODO(kjlubick): If there's no untriaged images, the blame should
			// likely be empty.
			Blame: []blame.WeightedBlame{
				{
					Author: bug_revert.BuggyAuthor,
					Prob:   1,
				},
			},
		},
		{
			Name:      bug_revert.TestTwo,
			Pos:       2,
			Untriaged: 1,
			UntHashes: types.DigestSlice{bug_revert.UntriagedDigestFoxtrot},
			Num:       3,
			Corpus:    "gm",
			Blame: []blame.WeightedBlame{
				{
					Author: bug_revert.InnocentAuthor,
					Prob:   0.5,
				},
				{
					Author: bug_revert.BuggyAuthor,
					Prob:   0.5,
				},
			},
		},
	}, sum)
}

func TestSummaryMap_NoMatch(t *testing.T) {
	unittest.SmallTest(t)

	sum := bugRevertHelper(t, url.Values{types.CORPUS_FIELD: {"does-not-exist"}}, false)
	require.Empty(t, sum)
}

func bugRevertHelper(t *testing.T, query url.Values, head bool) []*TriageStatus {
	dc := digest_counter.New(bug_revert.MakeTestTile())
	blamer, err := blame.New(bug_revert.MakeTestTile(), bug_revert.MakeTestExpectations())
	require.NoError(t, err)

	tr := asSlice(t, bug_revert.MakeTestTile().Traces)

	d := Data{
		Traces:       tr,
		Expectations: bug_revert.MakeTestExpectations(),
		ByTrace:      dc.ByTrace(),
		Blamer:       blamer,

		DiffStore: nil, // diameter is disabled, so this can be nil.
	}

	return d.Calculate(nil, query, head)
}

// TestSummaryMap_OverlappingCorpora makes sure that if we have two corpora that share a test name,
// we handle things correctly.
func TestSummaryMap_OverlappingCorpora(t *testing.T) {
	unittest.SmallTest(t)

	const corpusOneUntriaged = "1114c84eaa5dde4a247c93d9b93a136e"
	const corpusTwoUntriaged = "222b0d44658ad9c451c39e38c9281d47"
	const corpusOne = "corpusOne"
	const corpusTwo = "corpusTwo"

	commits := bug_revert.MakeTestCommits()[:2]

	tile := &tiling.Tile{
		Commits:   commits,
		Scale:     0, // tile contains every data point.
		TileIndex: 0,

		Traces: map[tiling.TraceID]tiling.Trace{
			",device=alpha,name=test_one,source_type=corpusOne,": types.NewGoldenTrace(
				types.DigestSlice{
					bug_revert.GoodDigestAlfa, corpusOneUntriaged,
				},
				map[string]string{
					"device":                bug_revert.AlphaDevice,
					types.PRIMARY_KEY_FIELD: string(bug_revert.TestOne),
					types.CORPUS_FIELD:      corpusOne,
				},
			),
			",device=beta,name=test_one,source_type=corpusTwo,": types.NewGoldenTrace(
				types.DigestSlice{
					corpusTwoUntriaged, corpusTwoUntriaged,
				},
				map[string]string{
					"device":                bug_revert.BetaDevice,
					types.PRIMARY_KEY_FIELD: string(bug_revert.TestOne),
					types.CORPUS_FIELD:      corpusTwo,
				},
			),
		},
	}

	var e expectations.Expectations
	e.Set(bug_revert.TestOne, bug_revert.GoodDigestAlfa, expectations.Positive)

	dc := digest_counter.New(tile)
	blamer, err := blame.New(tile, &e)
	require.NoError(t, err)

	tr := asSlice(t, tile.Traces)

	d := Data{
		Traces:       tr,
		Expectations: &e,
		ByTrace:      dc.ByTrace(),
		Blamer:       blamer,

		DiffStore: nil, // diameter is disabled, so this can be nil.
	}

	sum := d.Calculate(nil, nil, true)
	assert.Len(t, sum, 2)
	require.Equal(t, []*TriageStatus{
		{
			Name:      bug_revert.TestOne,
			Untriaged: 1,
			UntHashes: types.DigestSlice{corpusOneUntriaged},
			Num:       1,
			Corpus:    corpusOne,
			Blame: []blame.WeightedBlame{
				{
					Author: bug_revert.InnocentAuthor,
					Prob:   0.5,
				},
				{
					Author: bug_revert.BuggyAuthor,
					Prob:   0.5,
				},
			},
		},
		{
			Name:      bug_revert.TestOne,
			Untriaged: 1,
			UntHashes: types.DigestSlice{corpusTwoUntriaged},
			Num:       1,
			Corpus:    corpusTwo,
			Blame: []blame.WeightedBlame{
				{
					Author: bug_revert.InnocentAuthor,
					Prob:   0.5,
				},
				{
					Author: bug_revert.BuggyAuthor,
					Prob:   0.5,
				},
			},
		},
	}, sum)
}

// TestMergeSorted ensures we can combine two slices of TriageStatus to make sure
// we merge them correctly
func TestMergeSorted(t *testing.T) {
	unittest.SmallTest(t)

	first := []*TriageStatus{
		{
			Name:      FirstTest,
			Diameter:  4,
			Pos:       2,
			Neg:       3,
			Untriaged: 1,
			UntHashes: types.DigestSlice{AlphaDigest},
			Num:       6,
			Corpus:    "gm",
			Blame: []blame.WeightedBlame{
				{
					Author: "test@example.com",
					Prob:   1.00,
				},
			},
		},
		{
			Name:      SecondTest,
			Diameter:  14,
			Pos:       12,
			Neg:       13,
			Untriaged: 1,
			UntHashes: types.DigestSlice{BetaDigest},
			Num:       26,
			Corpus:    "gm",
			Blame: []blame.WeightedBlame{
				{
					Author: "other@example.com",
					Prob:   0.5,
				},
				{
					Author: "test@example.com",
					Prob:   0.5,
				},
			},
		},
	}

	second := []*TriageStatus{
		{
			Name:      FirstTest,
			Diameter:  24,
			Pos:       22,
			Neg:       23,
			Untriaged: 0,
			UntHashes: types.DigestSlice{},
			Num:       45,
			Corpus:    "gm",
			Blame:     []blame.WeightedBlame{},
		},
		{
			Name:      SecondTest,
			Diameter:  14,
			Pos:       12,
			Neg:       13,
			Untriaged: 1,
			UntHashes: types.DigestSlice{BetaDigest},
			Num:       26,
			Corpus:    "zzz",
			Blame: []blame.WeightedBlame{
				{
					Author: "other@example.com",
					Prob:   0.5,
				},
				{
					Author: "test@example.com",
					Prob:   0.5,
				},
			},
		},
		{
			Name:      ThirdTest,
			Diameter:  34,
			Pos:       32,
			Neg:       33,
			Untriaged: 1,
			UntHashes: types.DigestSlice{GammaDigest},
			Num:       66,
			Corpus:    "gm",
			Blame: []blame.WeightedBlame{
				{
					Author: "nobody@example.com",
					Prob:   1.0,
				},
			},
		},
	}

	res := MergeSorted(first, second)
	// first and second should remain unchanged
	require.Len(t, first, 2)
	assert.Equal(t, FirstTest, first[0].Name)
	assert.Equal(t, SecondTest, first[1].Name)
	require.Len(t, second, 3)
	assert.Equal(t, FirstTest, second[0].Name)
	assert.Equal(t, SecondTest, second[1].Name)
	assert.Equal(t, ThirdTest, second[2].Name)

	require.Equal(t, []*TriageStatus{second[0], first[1], second[1], second[2]}, res)
}

func find(sum []*TriageStatus, name types.TestName) *TriageStatus {
	for _, dft := range sum {
		if dft.Name == name {
			return dft
		}
	}
	return nil
}

func triageCountsCorrect(t *testing.T, ts *TriageStatus, pos, neg, unt int) {
	require.NotNil(t, ts)
	assert.Equal(t, pos, ts.Pos, "Positive count wrong")
	assert.Equal(t, neg, ts.Neg, "Negative count wrong")
	assert.Equal(t, unt, ts.Untriaged, "Untriaged count wrong")
}

const (
	FirstTest  = types.TestName("test_first")
	SecondTest = types.TestName("test_second")
	ThirdTest  = types.TestName("test_third")

	AlphaDigest = types.Digest("aaaec803b2ce49e4a541068d495ab570")
	BetaDigest  = types.Digest("bbb9ee4a034fdeddd1b65be92debe731")
	GammaDigest = types.Digest("ccc8f073f388469e0193300623691a36")
)

// makeFullTile returns a tile that matches the description at the top of the file.
func makeFullTile() *tiling.Tile {
	return &tiling.Tile{
		Traces: map[tiling.TraceID]tiling.Trace{
			// These trace ids have been shortened for test terseness.
			// A real trace id would be like "8888:gm:test_first"
			"a": &types.GoldenTrace{
				Digests: types.DigestSlice{"aaa", "bbb"},
				Keys: map[string]string{
					"config":                "8888",
					types.CORPUS_FIELD:      "gm",
					types.PRIMARY_KEY_FIELD: string(FirstTest),
				},
			},
			"b": &types.GoldenTrace{
				Digests: types.DigestSlice{"ccc", "ddd"},
				Keys: map[string]string{
					"config":                "565",
					types.CORPUS_FIELD:      "gm",
					types.PRIMARY_KEY_FIELD: string(FirstTest),
				},
			},
			"c": &types.GoldenTrace{
				Digests: types.DigestSlice{"eee", types.MISSING_DIGEST},
				Keys: map[string]string{
					"config":                "gpu",
					types.CORPUS_FIELD:      "gm",
					types.PRIMARY_KEY_FIELD: string(FirstTest),
				},
			},
			"d": &types.GoldenTrace{
				Digests: types.DigestSlice{"fff", "ggg"},
				Keys: map[string]string{
					"config":                "8888",
					types.CORPUS_FIELD:      "gm",
					types.PRIMARY_KEY_FIELD: string(SecondTest),
				},
			},
			"e": &types.GoldenTrace{
				Digests: types.DigestSlice{"jjj", types.MISSING_DIGEST},
				Keys: map[string]string{
					"config":                "8888",
					types.CORPUS_FIELD:      "image",
					types.PRIMARY_KEY_FIELD: string(ThirdTest),
				},
			},
		},
		Commits: []*tiling.Commit{
			{
				CommitTime: 42,
				Hash:       "ffffffffffffffffffffffffffffffffffffffff",
				Author:     "test@example.com",
			},
			{
				CommitTime: 45,
				Hash:       "gggggggggggggggggggggggggggggggggggggggg",
				Author:     "test@example.com",
			},
		},
		Scale:     0,
		TileIndex: 0,
	}
}

// makeTileWithIgnores() returns a tile with the ignore rule
// "config=565" applied (which as removed one trace compared to makeFullTile()).
func makeTileWithIgnores() *tiling.Tile {
	return &tiling.Tile{
		Traces: map[tiling.TraceID]tiling.Trace{
			// These trace ids have been shortened for test terseness.
			// A real trace id would be like "8888:gm:test_first"
			"a": &types.GoldenTrace{
				Digests: types.DigestSlice{"aaa", "bbb"},
				Keys: map[string]string{
					"config":                "8888",
					types.CORPUS_FIELD:      "gm",
					types.PRIMARY_KEY_FIELD: string(FirstTest),
				},
			},
			"c": &types.GoldenTrace{
				Digests: types.DigestSlice{"eee", types.MISSING_DIGEST},
				Keys: map[string]string{
					"config":                "gpu",
					types.CORPUS_FIELD:      "gm",
					types.PRIMARY_KEY_FIELD: string(FirstTest),
				},
			},
			"d": &types.GoldenTrace{
				Digests: types.DigestSlice{"fff", "ggg"},
				Keys: map[string]string{
					"config":                "8888",
					types.CORPUS_FIELD:      "gm",
					types.PRIMARY_KEY_FIELD: string(SecondTest),
				},
			},
			"e": &types.GoldenTrace{
				Digests: types.DigestSlice{"jjj", types.MISSING_DIGEST},
				Keys: map[string]string{
					"config":                "8888",
					types.CORPUS_FIELD:      "image",
					types.PRIMARY_KEY_FIELD: string(ThirdTest),
				},
			},
		},
		Commits: []*tiling.Commit{
			{
				CommitTime: 42,
				Hash:       "ffffffffffffffffffffffffffffffffffffffff",
				Author:     "test@example.com",
			},
			{
				CommitTime: 45,
				Hash:       "gggggggggggggggggggggggggggggggggggggggg",
				Author:     "test@example.com",
			},
		},
		Scale:     0,
		TileIndex: 0,
	}
}

func makeExpectations() *expectations.Expectations {
	var e expectations.Expectations
	e.Set(FirstTest, "aaa", expectations.Positive)
	e.Set(FirstTest, "bbb", expectations.Negative)
	e.Set(FirstTest, "ccc", expectations.Untriaged)
	e.Set(FirstTest, "ddd", expectations.Untriaged)
	e.Set(FirstTest, "eee", expectations.Positive)

	e.Set(SecondTest, "fff", expectations.Negative)
	return &e
}
