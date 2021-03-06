package search

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.skia.org/infra/go/paramtools"
	"go.skia.org/infra/go/testutils"
	"go.skia.org/infra/go/testutils/unittest"
	mock_clstore "go.skia.org/infra/golden/go/clstore/mocks"
	"go.skia.org/infra/golden/go/code_review"
	"go.skia.org/infra/golden/go/diff"
	mock_diffstore "go.skia.org/infra/golden/go/diffstore/mocks"
	"go.skia.org/infra/golden/go/digest_counter"
	"go.skia.org/infra/golden/go/expstorage/mocks"
	"go.skia.org/infra/golden/go/indexer"
	mock_index "go.skia.org/infra/golden/go/indexer/mocks"
	"go.skia.org/infra/golden/go/paramsets"
	"go.skia.org/infra/golden/go/search/common"
	"go.skia.org/infra/golden/go/search/frontend"
	"go.skia.org/infra/golden/go/search/query"
	data "go.skia.org/infra/golden/go/testutils/data_three_devices"
	"go.skia.org/infra/golden/go/tjstore"
	mock_tjstore "go.skia.org/infra/golden/go/tjstore/mocks"
	"go.skia.org/infra/golden/go/types"
	"go.skia.org/infra/golden/go/types/expectations"
)

// TODO(kjlubick) Add tests for:
//   - When a CL doesn't exist or the CL has not patchsets, patchset doesn't exist,
//     or otherwise no results.
//   - Use ignore matcher
//   - When a CL specifies a PS
//   - IncludeMaster=true
//   - UnavailableDigests is not empty
//   - DiffSever/RefDiffer error

// TestSearchThreeDevicesSunnyDay searches over the three_devices
// test data for untriaged images at head, essentially the default search.
// We expect to get two untriaged digests, with their closest positive and
// negative images (if any).
func TestSearchThreeDevicesSunnyDay(t *testing.T) {
	unittest.SmallTest(t)

	mes := &mocks.ExpectationsStore{}
	mi := &mock_index.IndexSource{}
	mds := &mock_diffstore.DiffStore{}
	defer mes.AssertExpectations(t)
	defer mi.AssertExpectations(t)
	defer mds.AssertExpectations(t)

	s := New(mds, mes, mi, nil, nil, everythingPublic)

	mes.On("Get", testutils.AnyContext).Return(data.MakeTestExpectations(), nil)

	fis := makeThreeDevicesIndex()
	mi.On("GetIndex").Return(fis)

	mds.On("UnavailableDigests", testutils.AnyContext).Return(map[types.Digest]*diff.DigestFailure{}, nil)
	// Positive match
	mds.On("Get", testutils.AnyContext, data.AlphaUntriaged1Digest, types.DigestSlice{data.AlphaGood1Digest}).
		Return(map[types.Digest]*diff.DiffMetrics{
			data.AlphaGood1Digest: makeSmallDiffMetric(),
		}, nil)
	// Negative match
	mds.On("Get", testutils.AnyContext, data.AlphaUntriaged1Digest, types.DigestSlice{data.AlphaBad1Digest}).
		Return(map[types.Digest]*diff.DiffMetrics{
			data.AlphaBad1Digest: makeBigDiffMetric(),
		}, nil)
	// Positive match
	mds.On("Get", testutils.AnyContext, data.BetaUntriaged1Digest, types.DigestSlice{data.BetaGood1Digest}).
		Return(map[types.Digest]*diff.DiffMetrics{
			data.BetaGood1Digest: makeBigDiffMetric(),
		}, nil)
	// BetaUntriaged1Digest has no negative images to compare against, so diffstore isn't queried.

	q := &query.Search{
		ChangeListID: "",
		Unt:          true,
		Head:         true,

		Metric:   diff.CombinedMetric,
		FRGBAMin: 0,
		FRGBAMax: 255,
		FDiffMax: -1,
		Sort:     query.SortAscending,
	}

	resp, err := s.Search(context.Background(), q)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, &frontend.SearchResponse{
		Commits: data.MakeTestCommits(),
		Offset:  0,
		Size:    2,
		Digests: []*frontend.SRDigest{
			// AlphaTest comes first because we are sorting by ascending
			// "combined" metric, and AlphaTest's closest match is the
			// small diff metric, whereas BetaTest's only match is the
			// big diff metric.
			{
				Test:   data.AlphaTest,
				Digest: data.AlphaUntriaged1Digest,
				Status: "untriaged",
				ParamSet: map[string][]string{
					"device":                {data.BullheadDevice},
					types.PRIMARY_KEY_FIELD: {string(data.AlphaTest)},
					types.CORPUS_FIELD:      {"gm"},
				},
				Traces: &frontend.TraceGroup{
					TileSize: 3, // 3 commits in tile
					Traces: []frontend.Trace{
						{
							Data: []frontend.Point{
								{X: 0, Y: 0, S: 1},
								{X: 1, Y: 0, S: 1},
								{X: 2, Y: 0, S: 0},
							},
							ID: data.BullheadAlphaTraceID,
							Params: map[string]string{
								"device":                data.BullheadDevice,
								types.PRIMARY_KEY_FIELD: string(data.AlphaTest),
								types.CORPUS_FIELD:      "gm",
							},
						},
					},
					Digests: []frontend.DigestStatus{
						{
							Digest: data.AlphaUntriaged1Digest,
							Status: "untriaged",
						},
						{
							Digest: data.AlphaBad1Digest,
							Status: "negative",
						},
					},
				},
				ClosestRef: common.PositiveRef,
				RefDiffs: map[common.RefClosest]*frontend.SRDiffDigest{
					common.PositiveRef: {
						DiffMetrics: makeSmallDiffMetric(),
						Digest:      data.AlphaGood1Digest,
						Status:      "positive",
						ParamSet: map[string][]string{
							"device":                {data.AnglerDevice, data.CrosshatchDevice},
							types.PRIMARY_KEY_FIELD: {string(data.AlphaTest)},
							types.CORPUS_FIELD:      {"gm"},
						},
						OccurrencesInTile: 2,
					},
					common.NegativeRef: {
						DiffMetrics: makeBigDiffMetric(),
						Digest:      data.AlphaBad1Digest,
						Status:      "negative",
						ParamSet: map[string][]string{
							"device":                {data.AnglerDevice, data.BullheadDevice, data.CrosshatchDevice},
							types.PRIMARY_KEY_FIELD: {string(data.AlphaTest)},
							types.CORPUS_FIELD:      {"gm"},
						},
						OccurrencesInTile: 6,
					},
				},
			},
			{
				Test:   data.BetaTest,
				Digest: data.BetaUntriaged1Digest,
				Status: "untriaged",
				ParamSet: map[string][]string{
					"device":                {data.CrosshatchDevice},
					types.PRIMARY_KEY_FIELD: {string(data.BetaTest)},
					types.CORPUS_FIELD:      {"gm"},
				},
				Traces: &frontend.TraceGroup{
					TileSize: 3,
					Traces: []frontend.Trace{
						{
							Data: []frontend.Point{
								{X: 0, Y: 0, S: 0},
								// Other two commits were missing
							},
							ID: data.CrosshatchBetaTraceID,
							Params: map[string]string{
								"device":                data.CrosshatchDevice,
								types.PRIMARY_KEY_FIELD: string(data.BetaTest),
								types.CORPUS_FIELD:      "gm",
							},
						},
					},
					Digests: []frontend.DigestStatus{
						{
							Digest: data.BetaUntriaged1Digest,
							Status: "untriaged",
						},
					},
				},
				ClosestRef: common.PositiveRef,
				RefDiffs: map[common.RefClosest]*frontend.SRDiffDigest{
					common.PositiveRef: {
						DiffMetrics: makeBigDiffMetric(),
						Digest:      data.BetaGood1Digest,
						Status:      "positive",
						ParamSet: map[string][]string{
							"device":                {data.AnglerDevice, data.BullheadDevice},
							types.PRIMARY_KEY_FIELD: {string(data.BetaTest)},
							types.CORPUS_FIELD:      {"gm"},
						},
						OccurrencesInTile: 6,
					},
					common.NegativeRef: nil,
				},
			},
		},
	}, resp)
}

// TestSearchThreeDevicesQueries searches over the three_devices test data using a variety
// of queries. It only spot-checks the returned data (e.g. things are in the right order); other
// tests should do a more thorough check of the return values.
func TestSearchThreeDevicesQueries(t *testing.T) {
	unittest.SmallTest(t)

	mes := &mocks.ExpectationsStore{}
	mi := &mock_index.IndexSource{}
	mds := &mock_diffstore.DiffStore{}
	defer mes.AssertExpectations(t)
	defer mi.AssertExpectations(t)
	defer mds.AssertExpectations(t)

	s := New(mds, mes, mi, nil, nil, everythingPublic)

	mes.On("Get", testutils.AnyContext).Return(data.MakeTestExpectations(), nil)

	fis := makeThreeDevicesIndex()
	mi.On("GetIndex").Return(fis)

	mds.On("UnavailableDigests", testutils.AnyContext).Return(map[types.Digest]*diff.DigestFailure{}, nil)
	// Positive match
	mds.On("Get", testutils.AnyContext, data.AlphaUntriaged1Digest, types.DigestSlice{data.AlphaGood1Digest}).
		Return(map[types.Digest]*diff.DiffMetrics{
			data.AlphaGood1Digest: makeSmallDiffMetric(),
		}, nil)
	// Negative match
	mds.On("Get", testutils.AnyContext, data.AlphaUntriaged1Digest, types.DigestSlice{data.AlphaBad1Digest}).
		Return(map[types.Digest]*diff.DiffMetrics{
			data.AlphaBad1Digest: makeBigDiffMetric(),
		}, nil)
	// Positive match
	mds.On("Get", testutils.AnyContext, data.BetaUntriaged1Digest, types.DigestSlice{data.BetaGood1Digest}).
		Return(map[types.Digest]*diff.DiffMetrics{
			data.BetaGood1Digest: makeBigDiffMetric(),
		}, nil)

	type spotCheck struct {
		test            types.TestName
		digest          types.Digest
		labelStr        string
		closestPositive types.Digest
		closestNegative types.Digest
	}

	test := func(name string, input *query.Search, expectedOutputs []spotCheck) {
		t.Run(name, func(t *testing.T) {
			resp, err := s.Search(context.Background(), input)
			require.NoError(t, err)
			require.NotNil(t, resp)

			require.Len(t, resp.Digests, len(expectedOutputs))
			for i, actualDigest := range resp.Digests {
				expected := expectedOutputs[i]
				assert.Equal(t, expected.test, actualDigest.Test)
				assert.Equal(t, expected.digest, actualDigest.Digest)
				assert.Equal(t, expected.labelStr, actualDigest.Status)
				if expected.closestPositive == "" {
					assert.Nil(t, actualDigest.RefDiffs[common.PositiveRef])
				} else {
					cp := actualDigest.RefDiffs[common.PositiveRef]
					require.NotNil(t, cp)
					assert.Equal(t, expected.closestPositive, cp.Digest)
				}
				if expected.closestNegative == "" {
					assert.Nil(t, actualDigest.RefDiffs[common.NegativeRef])
				} else {
					cp := actualDigest.RefDiffs[common.NegativeRef]
					require.NotNil(t, cp)
					assert.Equal(t, expected.closestNegative, cp.Digest)
				}
			}
		})
	}

	test("default query, but in reverse", &query.Search{
		Unt:  true,
		Head: true,

		Metric:   diff.CombinedMetric,
		FRGBAMin: 0,
		FRGBAMax: 255,
		FDiffMax: -1,
		Sort:     query.SortDescending,
	}, []spotCheck{
		{
			test:            data.BetaTest,
			digest:          data.BetaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.BetaGood1Digest,
			closestNegative: "",
		},
		{
			test:            data.AlphaTest,
			digest:          data.AlphaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.AlphaGood1Digest,
			closestNegative: data.AlphaBad1Digest,
		},
	})

	test("the closest RGBA diff should be at least 50 units away", &query.Search{
		Unt:  true,
		Head: true,

		Metric:   diff.CombinedMetric,
		FRGBAMin: 50,
		FRGBAMax: 255,
		FDiffMax: -1,
		Sort:     query.SortDescending,
	}, []spotCheck{
		{
			test:            data.BetaTest,
			digest:          data.BetaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.BetaGood1Digest,
			closestNegative: "",
		},
	})

	// note: this matches only the makeSmallDiffMetric
	test("the closest RGBA diff should be no more than 50 units away", &query.Search{
		Unt:  true,
		Head: true,

		Metric:   diff.CombinedMetric,
		FRGBAMin: 0,
		FRGBAMax: 50,
		FDiffMax: -1,
		Sort:     query.SortDescending,
	}, []spotCheck{
		{
			test:            data.AlphaTest,
			digest:          data.AlphaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.AlphaGood1Digest,
			closestNegative: data.AlphaBad1Digest,
		},
	})

	test("combined diff metric less than 1", &query.Search{
		Unt:  true,
		Head: true,

		Metric:   diff.CombinedMetric,
		FRGBAMin: 0,
		FRGBAMax: 255,
		FDiffMax: 1,
		Sort:     query.SortDescending,
	}, []spotCheck{
		{
			test:            data.AlphaTest,
			digest:          data.AlphaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.AlphaGood1Digest,
			closestNegative: data.AlphaBad1Digest,
		},
	})

	test("percent diff metric less than 1", &query.Search{
		Unt:  true,
		Head: true,

		Metric:   diff.PercentMetric,
		FRGBAMin: 0,
		FRGBAMax: 255,
		FDiffMax: 1,
		Sort:     query.SortDescending,
	}, []spotCheck{
		{
			test:            data.AlphaTest,
			digest:          data.AlphaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.AlphaGood1Digest,
			closestNegative: data.AlphaBad1Digest,
		},
	})

	test("Fewer than 10 different pixels", &query.Search{
		Unt:  true,
		Head: true,

		Metric:   diff.PixelMetric,
		FRGBAMin: 0,
		FRGBAMax: 255,
		FDiffMax: 10,
		Sort:     query.SortDescending,
	}, []spotCheck{
		{
			test:            data.AlphaTest,
			digest:          data.AlphaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.AlphaGood1Digest,
			closestNegative: data.AlphaBad1Digest,
		},
	})

	test("Nothing has fewer than 10 different pixels and min RGBA diff >50", &query.Search{
		Unt:  true,
		Head: true,

		Metric:   diff.PixelMetric,
		FRGBAMin: 50,
		FRGBAMax: 255,
		FDiffMax: 10,
		Sort:     query.SortDescending,
	}, nil)

	test("default query, only those with a reference diff (all of them)", &query.Search{
		Unt:  true,
		Head: true,
		FRef: true,

		Metric:   diff.CombinedMetric,
		FRGBAMin: 0,
		FRGBAMax: 255,
		FDiffMax: -1,
		Sort:     query.SortAscending,
	}, []spotCheck{
		{
			test:            data.AlphaTest,
			digest:          data.AlphaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.AlphaGood1Digest,
			closestNegative: data.AlphaBad1Digest,
		},
		{
			test:            data.BetaTest,
			digest:          data.BetaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.BetaGood1Digest,
			closestNegative: "",
		},
	})

	test("starting at the second commit, we only see alpha's untriaged commit at head", &query.Search{
		Unt:          true,
		Head:         true,
		FCommitBegin: data.MakeTestCommits()[1].Hash,

		Metric:   diff.CombinedMetric,
		FRGBAMin: 0,
		FRGBAMax: 255,
		FDiffMax: -1,
		Sort:     query.SortAscending,
	}, []spotCheck{
		{
			test:            data.AlphaTest,
			digest:          data.AlphaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.AlphaGood1Digest,
			closestNegative: data.AlphaBad1Digest,
		},
	})

	test("starting at the second commit, we see both if we ignore the head restriction", &query.Search{
		Unt:          true,
		Head:         false,
		FCommitBegin: data.MakeTestCommits()[1].Hash,

		Metric:   diff.CombinedMetric,
		FRGBAMin: 0,
		FRGBAMax: 255,
		FDiffMax: -1,
		Sort:     query.SortAscending,
	}, []spotCheck{
		{
			test:            data.AlphaTest,
			digest:          data.AlphaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.AlphaGood1Digest,
			closestNegative: data.AlphaBad1Digest,
		},
		{
			test:            data.BetaTest,
			digest:          data.BetaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.BetaGood1Digest,
			closestNegative: "",
		},
	})

	test("stopping at the second commit, we only see beta's untriaged", &query.Search{
		Unt:        true,
		Head:       true,
		FCommitEnd: data.MakeTestCommits()[1].Hash,

		Metric:   diff.CombinedMetric,
		FRGBAMin: 0,
		FRGBAMax: 255,
		FDiffMax: -1,
		Sort:     query.SortAscending,
	}, []spotCheck{
		{
			test:            data.BetaTest,
			digest:          data.BetaUntriaged1Digest,
			labelStr:        "untriaged",
			closestPositive: data.BetaGood1Digest,
			closestNegative: "",
		},
	})

	test("query matches nothing", &query.Search{
		Unt:  true,
		Head: true,
		TraceValues: map[string][]string{
			"blubber": {"nothing"},
		},

		Metric:   diff.CombinedMetric,
		FRGBAMin: 0,
		FRGBAMax: 255,
		FDiffMax: -1,
		Sort:     query.SortDescending,
	}, []spotCheck{})
}

// TestSearchThreeDevicesChangeListSunnyDay covers the case
// where two tryjobs have been run on a given CL and PS, one on the
// angler bot and one on the bullhead bot. The master branch
// looks like in the ThreeDevices data set. The outputs produced are
// Test  |  Device  | Digest
// ----------------------
// Alpha | Angler   | data.AlphaGood1Digest
// Alpha | Bullhead | data.AlphaUntriaged1Digest
// Beta  | Angler   | data.BetaGood1Digest
// Beta  | Bullhead | BetaBrandNewDigest
//
// The user has triaged the data.AlphaUntriaged1Digest as positive
// but BetaBrandNewDigest remains untriaged.
// With this setup, we do a default query (don't show master,
// only untriaged digests) and expect to see only an entry about
// BetaBrandNewDigest.
func TestSearchThreeDevicesChangeListSunnyDay(t *testing.T) {
	unittest.SmallTest(t)

	clID := "1234"
	crs := "gerrit"
	AlphaNowGoodDigest := data.AlphaUntriaged1Digest
	BetaBrandNewDigest := types.Digest("be7a03256511bec3a7453c3186bb2e07")

	mes := &mocks.ExpectationsStore{}
	issueStore := &mocks.ExpectationsStore{}
	mi := &mock_index.IndexSource{}
	mds := &mock_diffstore.DiffStore{}
	mcls := &mock_clstore.Store{}
	mtjs := &mock_tjstore.Store{}
	defer mes.AssertExpectations(t)
	defer issueStore.AssertExpectations(t)
	defer mi.AssertExpectations(t)
	defer mds.AssertExpectations(t)
	defer mcls.AssertExpectations(t)
	defer mtjs.AssertExpectations(t)

	mes.On("ForChangeList", clID, crs).Return(issueStore, nil)
	var ie expectations.Expectations
	ie.Set(data.AlphaTest, AlphaNowGoodDigest, expectations.Positive)
	issueStore.On("Get", testutils.AnyContext).Return(&ie, nil)
	mes.On("Get", testutils.AnyContext).Return(data.MakeTestExpectations(), nil)

	fis := makeThreeDevicesIndex()
	mi.On("GetIndex").Return(fis)

	mcls.On("GetPatchSets", testutils.AnyContext, clID).Return([]code_review.PatchSet{
		{
			SystemID:     "first_one",
			ChangeListID: clID,
			Order:        1,
			// All the rest are ignored
		},
		{
			SystemID:     "fourth_one",
			ChangeListID: clID,
			Order:        4,
			// All the rest are ignored
		},
	}, nil).Once() // this should be cached after fetch, as it could be expensive to retrieve.
	mcls.On("System").Return(crs)

	expectedID := tjstore.CombinedPSID{
		CL:  clID,
		CRS: crs,
		PS:  "fourth_one", // we didn't specify a PS, so it goes with the most recent
	}
	anglerGroup := map[string]string{
		"device": data.AnglerDevice,
	}
	bullheadGroup := map[string]string{
		"device": data.BullheadDevice,
	}
	options := map[string]string{
		"ext": "png",
	}

	mtjs.On("GetResults", testutils.AnyContext, expectedID).Return([]tjstore.TryJobResult{
		{
			GroupParams: anglerGroup,
			Options:     options,
			Digest:      data.AlphaGood1Digest,
			ResultParams: map[string]string{
				types.PRIMARY_KEY_FIELD: string(data.AlphaTest),
				types.CORPUS_FIELD:      "gm",
			},
		},
		{
			GroupParams: bullheadGroup,
			Options:     options,
			Digest:      AlphaNowGoodDigest,
			ResultParams: map[string]string{
				types.PRIMARY_KEY_FIELD: string(data.AlphaTest),
				types.CORPUS_FIELD:      "gm",
			},
		},
		{
			GroupParams: anglerGroup,
			Options:     options,
			Digest:      data.BetaGood1Digest,
			ResultParams: map[string]string{
				types.PRIMARY_KEY_FIELD: string(data.BetaTest),
				types.CORPUS_FIELD:      "gm",
			},
		},
		{
			GroupParams: bullheadGroup,
			Options:     options,
			Digest:      BetaBrandNewDigest,
			ResultParams: map[string]string{
				types.PRIMARY_KEY_FIELD: string(data.BetaTest),
				types.CORPUS_FIELD:      "gm",
			},
		},
	}, nil).Once() // this should be cached after fetch, as it could be expensive to retrieve.

	mds.On("UnavailableDigests", testutils.AnyContext).Return(map[types.Digest]*diff.DigestFailure{}, nil)

	mds.On("Get", testutils.AnyContext, BetaBrandNewDigest, types.DigestSlice{data.BetaGood1Digest}).
		Return(map[types.Digest]*diff.DiffMetrics{
			data.BetaGood1Digest: makeSmallDiffMetric(),
		}, nil)

	s := New(mds, mes, mi, mcls, mtjs, everythingPublic)

	q := &query.Search{
		ChangeListID:  clID,
		NewCLStore:    true,
		IncludeMaster: false,

		Unt:  true,
		Head: true,

		Metric:   diff.CombinedMetric,
		FRGBAMin: 0,
		FRGBAMax: 255,
		FDiffMax: -1,
		Sort:     query.SortAscending,
	}

	resp, err := s.Search(context.Background(), q)
	require.NoError(t, err)
	require.NotNil(t, resp)
	// make sure the group maps were not mutated.
	assert.Len(t, anglerGroup, 1)
	assert.Len(t, bullheadGroup, 1)
	assert.Len(t, options, 1)

	assert.Equal(t, &frontend.SearchResponse{
		Commits: data.MakeTestCommits(),
		Offset:  0,
		Size:    1,
		Digests: []*frontend.SRDigest{
			{
				Test:   data.BetaTest,
				Digest: BetaBrandNewDigest,
				Status: "untriaged",
				ParamSet: map[string][]string{
					"device":                {data.BullheadDevice},
					types.PRIMARY_KEY_FIELD: {string(data.BetaTest)},
					types.CORPUS_FIELD:      {"gm"},
					"ext":                   {"png"},
				},
				Traces: &frontend.TraceGroup{
					TileSize: 3,
					Traces:   []frontend.Trace{},
					Digests: []frontend.DigestStatus{
						{
							Digest: BetaBrandNewDigest,
							Status: "untriaged",
						},
					},
				},
				ClosestRef: common.PositiveRef,
				RefDiffs: map[common.RefClosest]*frontend.SRDiffDigest{
					common.PositiveRef: {
						DiffMetrics: makeSmallDiffMetric(),
						Digest:      data.BetaGood1Digest,
						Status:      "positive",
						ParamSet: map[string][]string{
							"device":                {data.AnglerDevice, data.BullheadDevice},
							types.PRIMARY_KEY_FIELD: {string(data.BetaTest)},
							types.CORPUS_FIELD:      {"gm"},
							// Note: the data from three_devices lacks an "ext" entry, so
							// we don't see one here
						},
						OccurrencesInTile: 6,
					},
					common.NegativeRef: nil,
				},
			},
		},
	}, resp)

	// Validate that we cache the .*Store values in two quick responses.
	_, err = s.Search(context.Background(), q)
	require.NoError(t, err)
}

func TestDigestDetailsThreeDevicesSunnyDay(t *testing.T) {
	unittest.SmallTest(t)

	const digestWeWantDetailsAbout = data.AlphaGood1Digest
	const testWeWantDetailsAbout = data.AlphaTest

	mes := &mocks.ExpectationsStore{}
	mi := &mock_index.IndexSource{}
	mds := &mock_diffstore.DiffStore{}
	defer mes.AssertExpectations(t)
	defer mi.AssertExpectations(t)
	defer mds.AssertExpectations(t)

	fis := makeThreeDevicesIndex()
	mi.On("GetIndex").Return(fis)

	mes.On("Get", testutils.AnyContext).Return(data.MakeTestExpectations(), nil)

	mds.On("UnavailableDigests", testutils.AnyContext).Return(map[types.Digest]*diff.DigestFailure{}, nil)

	// Positive match. Note If a digest is compared to itself, it is removed from the return value,
	// so we return an empty map.
	mds.On("Get", testutils.AnyContext, digestWeWantDetailsAbout, types.DigestSlice{data.AlphaGood1Digest}).
		Return(map[types.Digest]*diff.DiffMetrics{}, nil)
	// Negative match
	mds.On("Get", testutils.AnyContext, digestWeWantDetailsAbout, types.DigestSlice{data.AlphaBad1Digest}).
		Return(map[types.Digest]*diff.DiffMetrics{
			data.AlphaBad1Digest: makeBigDiffMetric(),
		}, nil)

	s := New(mds, mes, mi, nil, nil, everythingPublic)

	details, err := s.GetDigestDetails(context.Background(), testWeWantDetailsAbout, digestWeWantDetailsAbout, "", "")
	require.NoError(t, err)
	assert.Equal(t, &frontend.DigestDetails{
		Commits: data.MakeTestCommits(),
		Digest: &frontend.SRDigest{
			Test:   testWeWantDetailsAbout,
			Digest: digestWeWantDetailsAbout,
			Status: "positive",
			ParamSet: map[string][]string{
				"device":                {data.AnglerDevice, data.CrosshatchDevice},
				types.PRIMARY_KEY_FIELD: {string(data.AlphaTest)},
				types.CORPUS_FIELD:      {"gm"},
			},
			Traces: &frontend.TraceGroup{
				TileSize: 3, // 3 commits in tile
				Traces: []frontend.Trace{ // the digest we care about appears in two traces
					{
						Data: []frontend.Point{
							{X: 0, Y: 0, S: 1},
							{X: 1, Y: 0, S: 1},
							{X: 2, Y: 0, S: 0},
						},
						ID: data.AnglerAlphaTraceID,
						Params: map[string]string{
							"device":                data.AnglerDevice,
							types.PRIMARY_KEY_FIELD: string(data.AlphaTest),
							types.CORPUS_FIELD:      "gm",
						},
					},
					{
						Data: []frontend.Point{
							{X: 0, Y: 1, S: 1},
							{X: 1, Y: 1, S: 1},
							{X: 2, Y: 1, S: 0},
						},
						ID: data.CrosshatchAlphaTraceID,
						Params: map[string]string{
							"device":                data.CrosshatchDevice,
							types.PRIMARY_KEY_FIELD: string(data.AlphaTest),
							types.CORPUS_FIELD:      "gm",
						},
					},
				},
				Digests: []frontend.DigestStatus{
					{
						Digest: data.AlphaGood1Digest,
						Status: "positive",
					},
					{
						Digest: data.AlphaBad1Digest,
						Status: "negative",
					},
				},
			},
			ClosestRef: common.NegativeRef,
			RefDiffs: map[common.RefClosest]*frontend.SRDiffDigest{
				common.PositiveRef: nil,
				common.NegativeRef: {
					DiffMetrics: makeBigDiffMetric(),
					Digest:      data.AlphaBad1Digest,
					Status:      "negative",
					ParamSet: map[string][]string{
						"device":                {data.AnglerDevice, data.BullheadDevice, data.CrosshatchDevice},
						types.PRIMARY_KEY_FIELD: {string(data.AlphaTest)},
						types.CORPUS_FIELD:      {"gm"},
					},
					OccurrencesInTile: 6,
				},
			},
		},
	}, details)
}

func TestDigestDetailsThreeDevicesChangeList(t *testing.T) {
	unittest.SmallTest(t)

	const digestWeWantDetailsAbout = data.AlphaGood1Digest
	const testWeWantDetailsAbout = data.AlphaTest
	const testCLID = "abc12345"
	const testCRS = "gerritHub"

	mes := &mocks.ExpectationsStore{}
	mi := &mock_index.IndexSource{}
	mds := &mock_diffstore.DiffStore{}
	defer mes.AssertExpectations(t)
	defer mi.AssertExpectations(t)
	defer mds.AssertExpectations(t)

	fis := makeThreeDevicesIndex()
	mi.On("GetIndex").Return(fis)

	// Mock out some ChangeList expectations in which the digest we care about is negative
	override := &expectations.Expectations{}
	override.Set(testWeWantDetailsAbout, digestWeWantDetailsAbout, expectations.Negative)
	cles := &mocks.ExpectationsStore{}
	defer cles.AssertExpectations(t)
	cles.On("Get", testutils.AnyContext).Return(override, nil)

	mes.On("Get", testutils.AnyContext).Return(data.MakeTestExpectations(), nil)
	mes.On("ForChangeList", testCLID, testCRS).Return(cles, nil)

	mds.On("UnavailableDigests", testutils.AnyContext).Return(map[types.Digest]*diff.DigestFailure{}, nil)

	// There are no positive digests with which to compare
	// Negative match. Note If a digest is compared to itself, it is removed from the return value.
	mds.On("Get", testutils.AnyContext, digestWeWantDetailsAbout, types.DigestSlice{digestWeWantDetailsAbout, data.AlphaBad1Digest}).
		Return(map[types.Digest]*diff.DiffMetrics{
			data.AlphaBad1Digest: makeBigDiffMetric(),
		}, nil)

	s := New(mds, mes, mi, nil, nil, everythingPublic)

	details, err := s.GetDigestDetails(context.Background(), testWeWantDetailsAbout, digestWeWantDetailsAbout, testCLID, testCRS)
	require.NoError(t, err)
	assert.Equal(t, details.Digest.Status, expectations.Negative.String())
}

// TestDigestDetailsThreeDevicesOldDigest represents the scenario in which a user is requesting
// data about a digest that just went off the tile.
func TestDigestDetailsThreeDevicesOldDigest(t *testing.T) {
	unittest.SmallTest(t)

	const digestWeWantDetailsAbout = types.Digest("digest-too-old")
	const testWeWantDetailsAbout = data.BetaTest

	mes := &mocks.ExpectationsStore{}
	mi := &mock_index.IndexSource{}
	mds := &mock_diffstore.DiffStore{}
	defer mes.AssertExpectations(t)
	defer mi.AssertExpectations(t)
	defer mds.AssertExpectations(t)

	fis := makeThreeDevicesIndex()
	mi.On("GetIndex").Return(fis)

	mes.On("Get", testutils.AnyContext).Return(data.MakeTestExpectations(), nil)

	mds.On("UnavailableDigests", testutils.AnyContext).Return(map[types.Digest]*diff.DigestFailure{}, nil)

	mds.On("Get", testutils.AnyContext, digestWeWantDetailsAbout, types.DigestSlice{data.BetaGood1Digest}).
		Return(map[types.Digest]*diff.DiffMetrics{
			data.BetaGood1Digest: makeSmallDiffMetric(),
		}, nil)

	s := New(mds, mes, mi, nil, nil, everythingPublic)

	d, err := s.GetDigestDetails(context.Background(), testWeWantDetailsAbout, digestWeWantDetailsAbout, "", "")
	require.NoError(t, err)
	// spot check is fine for this test because other tests do a more thorough check of the
	// whole struct.
	assert.Equal(t, digestWeWantDetailsAbout, d.Digest.Digest)
	assert.Equal(t, testWeWantDetailsAbout, d.Digest.Test)
	assert.Equal(t, map[common.RefClosest]*frontend.SRDiffDigest{
		common.PositiveRef: {
			DiffMetrics: makeSmallDiffMetric(),
			Digest:      data.BetaGood1Digest,
			Status:      "positive",
			ParamSet: paramtools.ParamSet{
				"device":                []string{data.AnglerDevice, data.BullheadDevice},
				types.PRIMARY_KEY_FIELD: []string{string(data.BetaTest)},
				types.CORPUS_FIELD:      []string{"gm"},
			},
			OccurrencesInTile: 6,
		},
		common.NegativeRef: nil,
	}, d.Digest.RefDiffs)
}

// TestDigestDetailsThreeDevicesOldDigest represents the scenario in which a user is requesting
// data about a digest that never existed. In the past, when this has happened, it has broken
// Gold until that digest went away (e.g. because a bot only uploaded a subset of images).
// Therefore, we shouldn't error the search request, because it could break all searches for
// untriaged digests.
func TestDigestDetailsThreeDevicesBadDigest(t *testing.T) {
	unittest.SmallTest(t)

	const digestWeWantDetailsAbout = types.Digest("unknown-digest")
	const testWeWantDetailsAbout = data.BetaTest

	mes := &mocks.ExpectationsStore{}
	mi := &mock_index.IndexSource{}
	mds := &mock_diffstore.DiffStore{}
	defer mes.AssertExpectations(t)
	defer mi.AssertExpectations(t)
	defer mds.AssertExpectations(t)

	fis := makeThreeDevicesIndex()
	mi.On("GetIndex").Return(fis)

	mes.On("Get", testutils.AnyContext).Return(data.MakeTestExpectations(), nil)

	mds.On("UnavailableDigests", testutils.AnyContext).Return(map[types.Digest]*diff.DigestFailure{}, nil)

	mds.On("Get", testutils.AnyContext, digestWeWantDetailsAbout, types.DigestSlice{data.BetaGood1Digest}).Return(nil, errors.New("invalid digest"))

	s := New(mds, mes, mi, nil, nil, everythingPublic)

	r, err := s.GetDigestDetails(context.Background(), testWeWantDetailsAbout, digestWeWantDetailsAbout, "", "")
	require.NoError(t, err)
	// Since we couldn't find the digest, we have nothing to compare against.
	assert.Equal(t, r.Digest.Digest, digestWeWantDetailsAbout)
	assert.Equal(t, r.Digest.ClosestRef, common.NoRef)
}

func TestDigestDetailsThreeDevicesBadTest(t *testing.T) {
	unittest.SmallTest(t)

	const digestWeWantDetailsAbout = data.AlphaGood1Digest
	const testWeWantDetailsAbout = types.TestName("invalid test")

	mi := &mock_index.IndexSource{}
	defer mi.AssertExpectations(t)

	fis := makeThreeDevicesIndex()
	mi.On("GetIndex").Return(fis)

	s := New(nil, nil, mi, nil, nil, everythingPublic)

	_, err := s.GetDigestDetails(context.Background(), testWeWantDetailsAbout, digestWeWantDetailsAbout, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown")
}

func TestDigestDetailsThreeDevicesBadTestAndDigest(t *testing.T) {
	unittest.SmallTest(t)

	const digestWeWantDetailsAbout = types.Digest("invalid digest")
	const testWeWantDetailsAbout = types.TestName("invalid test")

	mi := &mock_index.IndexSource{}
	defer mi.AssertExpectations(t)

	fis := makeThreeDevicesIndex()
	mi.On("GetIndex").Return(fis)

	s := New(nil, nil, mi, nil, nil, everythingPublic)

	_, err := s.GetDigestDetails(context.Background(), testWeWantDetailsAbout, digestWeWantDetailsAbout, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown")
}

func TestDiffDigestsSunnyDay(t *testing.T) {
	unittest.SmallTest(t)

	const testWeWantDetailsAbout = data.AlphaTest
	const leftDigest = data.AlphaUntriaged1Digest
	const rightDigest = data.AlphaGood1Digest

	mes := &mocks.ExpectationsStore{}
	mi := &mock_index.IndexSource{}
	mds := &mock_diffstore.DiffStore{}
	defer mes.AssertExpectations(t)
	defer mi.AssertExpectations(t)
	defer mds.AssertExpectations(t)

	fis := makeThreeDevicesIndex()
	mi.On("GetIndex").Return(fis)

	mes.On("Get", testutils.AnyContext).Return(data.MakeTestExpectations(), nil)

	mds.On("Get", testutils.AnyContext, leftDigest, types.DigestSlice{rightDigest}).
		Return(map[types.Digest]*diff.DiffMetrics{
			rightDigest: makeSmallDiffMetric(),
		}, nil)

	s := New(mds, mes, mi, nil, nil, everythingPublic)

	cd, err := s.DiffDigests(context.Background(), testWeWantDetailsAbout, leftDigest, rightDigest, "", "")
	require.NoError(t, err)
	assert.Equal(t, &frontend.DigestComparison{
		Left: &frontend.SRDigest{
			Test:   testWeWantDetailsAbout,
			Digest: leftDigest,
			Status: expectations.Untriaged.String(),
			ParamSet: paramtools.ParamSet{
				"device":                []string{data.BullheadDevice},
				types.PRIMARY_KEY_FIELD: []string{string(data.AlphaTest)},
				types.CORPUS_FIELD:      []string{"gm"},
			},
		},
		Right: &frontend.SRDiffDigest{
			Digest:      rightDigest,
			Status:      expectations.Positive.String(),
			DiffMetrics: makeSmallDiffMetric(),
			ParamSet: paramtools.ParamSet{
				"device":                []string{data.AnglerDevice, data.CrosshatchDevice},
				types.PRIMARY_KEY_FIELD: []string{string(data.AlphaTest)},
				types.CORPUS_FIELD:      []string{"gm"},
			},
		},
	}, cd)
}

func TestDiffDigestsChangeList(t *testing.T) {
	unittest.SmallTest(t)

	const testWeWantDetailsAbout = data.AlphaTest
	const leftDigest = data.AlphaUntriaged1Digest
	const rightDigest = data.AlphaGood1Digest
	const clID = "abc12354"
	const crs = "gerritHub"

	mes := &mocks.ExpectationsStore{}
	mi := &mock_index.IndexSource{}
	mds := &mock_diffstore.DiffStore{}
	issueStore := &mocks.ExpectationsStore{}
	defer mes.AssertExpectations(t)
	defer mi.AssertExpectations(t)
	defer mds.AssertExpectations(t)
	defer issueStore.AssertExpectations(t)

	fis := makeThreeDevicesIndex()
	mi.On("GetIndex").Return(fis)

	mes.On("ForChangeList", clID, crs).Return(issueStore, nil)
	var ie expectations.Expectations
	ie.Set(data.AlphaTest, leftDigest, expectations.Negative)
	issueStore.On("Get", testutils.AnyContext).Return(&ie, nil)
	mes.On("Get", testutils.AnyContext).Return(data.MakeTestExpectations(), nil)

	mds.On("Get", testutils.AnyContext, leftDigest, types.DigestSlice{rightDigest}).
		Return(map[types.Digest]*diff.DiffMetrics{
			rightDigest: makeSmallDiffMetric(),
		}, nil)

	s := New(mds, mes, mi, nil, nil, everythingPublic)

	cd, err := s.DiffDigests(context.Background(), testWeWantDetailsAbout, leftDigest, rightDigest, clID, crs)
	require.NoError(t, err)
	assert.Equal(t, cd.Left.Status, expectations.Negative.String())
}

// TestUntriagedUnignoredTryJobExclusiveDigestsSunnyDay models the case where a set of TryJobs has
// produced five digests that were "untriaged on master" (and one good digest). We are testing that
// we can properly deduce which are untriaged, "newly seen" and unignored. One of these untriaged
// digests was already seen on master (data.AlphaUntriaged1Digest), one was already triaged negative
// for this CL (gammaNegativeTryJobDigest), and one trace matched an ignore rule (deltaIgnoredTryJobDigest). Thus,
// We only expect tjUntriagedAlpha and tjUntriagedBeta to be reported to us.
func TestUntriagedUnignoredTryJobExclusiveDigestsSunnyDay(t *testing.T) {
	unittest.SmallTest(t)

	const clID = "44474"
	const crs = "github"
	expectedID := tjstore.CombinedPSID{
		CL:  clID,
		CRS: crs,
		PS:  "abcdef",
	}

	const alphaUntriagedTryJobDigest = types.Digest("aaaa65e567de97c8a62918401731c7ec")
	const betaUntriagedTryJobDigest = types.Digest("bbbb34f7c915a1ac3a5ba524c741946c")
	const gammaNegativeTryJobDigest = types.Digest("cccc41bf4584e51be99e423707157277")
	const deltaIgnoredTryJobDigest = types.Digest("dddd84e51be99e42370715727765e563")

	mes := &mocks.ExpectationsStore{}
	mi := &mock_index.IndexSource{}
	issueStore := &mocks.ExpectationsStore{}
	mtjs := &mock_tjstore.Store{}
	defer mes.AssertExpectations(t)
	defer mi.AssertExpectations(t)
	defer issueStore.AssertExpectations(t)
	defer mtjs.AssertExpectations(t)

	cpxTile := types.NewComplexTile(data.MakeTestTile())
	reduced := data.MakeTestTile()
	delete(reduced.Traces, data.BullheadBetaTraceID)
	// The following rule exclusively matches BullheadBetaTraceID, for which the tryjob produced
	// deltaIgnoredTryJobDigest
	cpxTile.SetIgnoreRules(reduced, paramtools.ParamMatcher{
		{
			"device":                []string{data.BullheadDevice},
			types.PRIMARY_KEY_FIELD: []string{string(data.BetaTest)},
		},
	})
	dc := digest_counter.New(data.MakeTestTile())
	fis, err := indexer.SearchIndexForTesting(cpxTile, [2]digest_counter.DigestCounter{dc, dc}, [2]paramsets.ParamSummary{}, mes, nil)
	require.NoError(t, err)
	mi.On("GetIndex").Return(fis)

	// Set up the expectations such that for this CL, we have one extra expectation - marking
	// gammaNegativeTryJobDigest negative (it would be untriaged on master).
	mes.On("ForChangeList", clID, crs).Return(issueStore, nil)
	var ie expectations.Expectations
	ie.Set(data.AlphaTest, gammaNegativeTryJobDigest, expectations.Negative)
	issueStore.On("Get", testutils.AnyContext).Return(&ie, nil)
	mes.On("Get", testutils.AnyContext).Return(data.MakeTestExpectations(), nil)

	anglerGroup := map[string]string{
		"device": data.AnglerDevice,
	}
	bullheadGroup := map[string]string{
		"device": data.BullheadDevice,
	}
	crosshatchGroup := map[string]string{
		"device": data.CrosshatchDevice,
	}
	options := map[string]string{
		"ext": "png",
	}
	mtjs.On("GetResults", testutils.AnyContext, expectedID).Return([]tjstore.TryJobResult{
		{
			GroupParams: anglerGroup,
			Options:     options,
			Digest:      betaUntriagedTryJobDigest,
			ResultParams: map[string]string{
				types.PRIMARY_KEY_FIELD: string(data.AlphaTest),
				types.CORPUS_FIELD:      "gm",
			},
		},
		{
			GroupParams: bullheadGroup,
			Options:     options,
			Digest:      data.AlphaUntriaged1Digest,
			ResultParams: map[string]string{
				types.PRIMARY_KEY_FIELD: string(data.AlphaTest),
				types.CORPUS_FIELD:      "gm",
			},
		},
		{
			GroupParams: anglerGroup,
			Options:     options,
			Digest:      alphaUntriagedTryJobDigest,
			ResultParams: map[string]string{
				types.PRIMARY_KEY_FIELD: string(data.BetaTest),
				types.CORPUS_FIELD:      "gm",
			},
		},
		{
			GroupParams: bullheadGroup,
			Options:     options,
			Digest:      deltaIgnoredTryJobDigest,
			ResultParams: map[string]string{
				types.PRIMARY_KEY_FIELD: string(data.BetaTest),
				types.CORPUS_FIELD:      "gm",
			},
		},
		{
			GroupParams: crosshatchGroup,
			Options:     options,
			Digest:      gammaNegativeTryJobDigest,
			ResultParams: map[string]string{
				types.PRIMARY_KEY_FIELD: string(data.AlphaTest),
				types.CORPUS_FIELD:      "gm",
			},
		},
		{
			GroupParams: crosshatchGroup,
			Options:     options,
			Digest:      data.BetaGood1Digest,
			ResultParams: map[string]string{
				types.PRIMARY_KEY_FIELD: string(data.BetaTest),
				types.CORPUS_FIELD:      "gm",
			},
		},
	}, nil).Once()

	s := New(nil, mes, mi, nil, mtjs, everythingPublic)

	dl, err := s.UntriagedUnignoredTryJobExclusiveDigests(context.Background(), expectedID)
	require.NoError(t, err)
	assert.Equal(t, []types.Digest{alphaUntriagedTryJobDigest, betaUntriagedTryJobDigest}, dl.Digests)
}

var everythingPublic = paramtools.ParamSet{}

// makeThreeDevicesIndex returns a search index corresponding to the three_devices_data
// (which currently has nothing ignored).
func makeThreeDevicesIndex() *indexer.SearchIndex {
	cpxTile := types.NewComplexTile(data.MakeTestTile())
	dc := digest_counter.New(data.MakeTestTile())
	ps := paramsets.NewParamSummary(data.MakeTestTile(), dc)
	si, err := indexer.SearchIndexForTesting(cpxTile, [2]digest_counter.DigestCounter{dc, dc}, [2]paramsets.ParamSummary{ps, ps}, nil, nil)
	if err != nil {
		// Something is horribly broken with our test data/setup
		panic(err.Error())
	}
	return si
}

// This is arbitrary data.
func makeSmallDiffMetric() *diff.DiffMetrics {
	return &diff.DiffMetrics{
		NumDiffPixels:    8,
		PixelDiffPercent: 0.02,
		MaxRGBADiffs:     [4]int{0, 48, 12, 0},
		DimDiffer:        false,
		Diffs: map[string]float32{
			diff.CombinedMetric: 0.0005,
			"percent":           0.02,
			"pixel":             8,
		},
	}
}

func makeBigDiffMetric() *diff.DiffMetrics {
	return &diff.DiffMetrics{
		NumDiffPixels:    88812,
		PixelDiffPercent: 98.68,
		MaxRGBADiffs:     [4]int{102, 51, 13, 0},
		DimDiffer:        true,
		Diffs: map[string]float32{
			diff.CombinedMetric: 4.7,
			"percent":           98.68,
			"pixel":             88812,
		},
	}
}
