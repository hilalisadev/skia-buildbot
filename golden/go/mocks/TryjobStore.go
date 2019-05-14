// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import tryjobstore "go.skia.org/infra/golden/go/tryjobstore"

// TryjobStore is an autogenerated mock type for the TryjobStore type
type TryjobStore struct {
	mock.Mock
}

// CommitIssueExp provides a mock function with given fields: issueID, writeFn
func (_m *TryjobStore) CommitIssueExp(issueID int64, writeFn func() error) error {
	ret := _m.Called(issueID, writeFn)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64, func() error) error); ok {
		r0 = rf(issueID, writeFn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteIssue provides a mock function with given fields: issueID
func (_m *TryjobStore) DeleteIssue(issueID int64) error {
	ret := _m.Called(issueID)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64) error); ok {
		r0 = rf(issueID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetIssue provides a mock function with given fields: issueID, loadTryjobs
func (_m *TryjobStore) GetIssue(issueID int64, loadTryjobs bool) (*tryjobstore.Issue, error) {
	ret := _m.Called(issueID, loadTryjobs)

	var r0 *tryjobstore.Issue
	if rf, ok := ret.Get(0).(func(int64, bool) *tryjobstore.Issue); ok {
		r0 = rf(issueID, loadTryjobs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tryjobstore.Issue)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64, bool) error); ok {
		r1 = rf(issueID, loadTryjobs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTryjob provides a mock function with given fields: issueID, buildBucketID
func (_m *TryjobStore) GetTryjob(issueID int64, buildBucketID int64) (*tryjobstore.Tryjob, error) {
	ret := _m.Called(issueID, buildBucketID)

	var r0 *tryjobstore.Tryjob
	if rf, ok := ret.Get(0).(func(int64, int64) *tryjobstore.Tryjob); ok {
		r0 = rf(issueID, buildBucketID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tryjobstore.Tryjob)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64, int64) error); ok {
		r1 = rf(issueID, buildBucketID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTryjobResults provides a mock function with given fields: tryjobs
func (_m *TryjobStore) GetTryjobResults(tryjobs []*tryjobstore.Tryjob) ([][]*tryjobstore.TryjobResult, error) {
	ret := _m.Called(tryjobs)

	var r0 [][]*tryjobstore.TryjobResult
	if rf, ok := ret.Get(0).(func([]*tryjobstore.Tryjob) [][]*tryjobstore.TryjobResult); ok {
		r0 = rf(tryjobs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]*tryjobstore.TryjobResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*tryjobstore.Tryjob) error); ok {
		r1 = rf(tryjobs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTryjobs provides a mock function with given fields: issueID, patchsetIDs, filterDup, loadResults
func (_m *TryjobStore) GetTryjobs(issueID int64, patchsetIDs []int64, filterDup bool, loadResults bool) ([]*tryjobstore.Tryjob, [][]*tryjobstore.TryjobResult, error) {
	ret := _m.Called(issueID, patchsetIDs, filterDup, loadResults)

	var r0 []*tryjobstore.Tryjob
	if rf, ok := ret.Get(0).(func(int64, []int64, bool, bool) []*tryjobstore.Tryjob); ok {
		r0 = rf(issueID, patchsetIDs, filterDup, loadResults)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*tryjobstore.Tryjob)
		}
	}

	var r1 [][]*tryjobstore.TryjobResult
	if rf, ok := ret.Get(1).(func(int64, []int64, bool, bool) [][]*tryjobstore.TryjobResult); ok {
		r1 = rf(issueID, patchsetIDs, filterDup, loadResults)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([][]*tryjobstore.TryjobResult)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(int64, []int64, bool, bool) error); ok {
		r2 = rf(issueID, patchsetIDs, filterDup, loadResults)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ListIssues provides a mock function with given fields: offset, size
func (_m *TryjobStore) ListIssues(offset int, size int) ([]*tryjobstore.Issue, int, error) {
	ret := _m.Called(offset, size)

	var r0 []*tryjobstore.Issue
	if rf, ok := ret.Get(0).(func(int, int) []*tryjobstore.Issue); ok {
		r0 = rf(offset, size)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*tryjobstore.Issue)
		}
	}

	var r1 int
	if rf, ok := ret.Get(1).(func(int, int) int); ok {
		r1 = rf(offset, size)
	} else {
		r1 = ret.Get(1).(int)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(int, int) error); ok {
		r2 = rf(offset, size)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// RunningTryjobs provides a mock function with given fields:
func (_m *TryjobStore) RunningTryjobs() ([]*tryjobstore.Tryjob, error) {
	ret := _m.Called()

	var r0 []*tryjobstore.Tryjob
	if rf, ok := ret.Get(0).(func() []*tryjobstore.Tryjob); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*tryjobstore.Tryjob)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateIssue provides a mock function with given fields: details, updateFn
func (_m *TryjobStore) UpdateIssue(details *tryjobstore.Issue, updateFn tryjobstore.NewValueFn) error {
	ret := _m.Called(details, updateFn)

	var r0 error
	if rf, ok := ret.Get(0).(func(*tryjobstore.Issue, tryjobstore.NewValueFn) error); ok {
		r0 = rf(details, updateFn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateTryjob provides a mock function with given fields: buildBucketID, tryjob, newValFn
func (_m *TryjobStore) UpdateTryjob(buildBucketID int64, tryjob *tryjobstore.Tryjob, newValFn tryjobstore.NewValueFn) error {
	ret := _m.Called(buildBucketID, tryjob, newValFn)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64, *tryjobstore.Tryjob, tryjobstore.NewValueFn) error); ok {
		r0 = rf(buildBucketID, tryjob, newValFn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateTryjobResult provides a mock function with given fields: results
func (_m *TryjobStore) UpdateTryjobResult(results []*tryjobstore.TryjobResult) error {
	ret := _m.Called(results)

	var r0 error
	if rf, ok := ret.Get(0).(func([]*tryjobstore.TryjobResult) error); ok {
		r0 = rf(results)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
