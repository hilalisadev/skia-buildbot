// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	gitstore_deprecated "go.skia.org/infra/go/gitstore_deprecated"

	time "time"

	vcsinfo "go.skia.org/infra/go/vcsinfo"
)

// GitStore is an autogenerated mock type for the GitStore type
type GitStore struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, hashes
func (_m *GitStore) Get(ctx context.Context, hashes []string) ([]*vcsinfo.LongCommit, error) {
	ret := _m.Called(ctx, hashes)

	var r0 []*vcsinfo.LongCommit
	if rf, ok := ret.Get(0).(func(context.Context, []string) []*vcsinfo.LongCommit); ok {
		r0 = rf(ctx, hashes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*vcsinfo.LongCommit)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, hashes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBranches provides a mock function with given fields: ctx
func (_m *GitStore) GetBranches(ctx context.Context) (map[string]*gitstore_deprecated.BranchPointer, error) {
	ret := _m.Called(ctx)

	var r0 map[string]*gitstore_deprecated.BranchPointer
	if rf, ok := ret.Get(0).(func(context.Context) map[string]*gitstore_deprecated.BranchPointer); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]*gitstore_deprecated.BranchPointer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGraph provides a mock function with given fields: ctx
func (_m *GitStore) GetGraph(ctx context.Context) (*gitstore_deprecated.CommitGraph, error) {
	ret := _m.Called(ctx)

	var r0 *gitstore_deprecated.CommitGraph
	if rf, ok := ret.Get(0).(func(context.Context) *gitstore_deprecated.CommitGraph); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitstore_deprecated.CommitGraph)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Put provides a mock function with given fields: ctx, commits
func (_m *GitStore) Put(ctx context.Context, commits []*vcsinfo.LongCommit) error {
	ret := _m.Called(ctx, commits)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*vcsinfo.LongCommit) error); ok {
		r0 = rf(ctx, commits)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PutBranches provides a mock function with given fields: ctx, branches
func (_m *GitStore) PutBranches(ctx context.Context, branches map[string]string) error {
	ret := _m.Called(ctx, branches)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, map[string]string) error); ok {
		r0 = rf(ctx, branches)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RangeByTime provides a mock function with given fields: ctx, start, end, branch
func (_m *GitStore) RangeByTime(ctx context.Context, start time.Time, end time.Time, branch string) ([]*vcsinfo.IndexCommit, error) {
	ret := _m.Called(ctx, start, end, branch)

	var r0 []*vcsinfo.IndexCommit
	if rf, ok := ret.Get(0).(func(context.Context, time.Time, time.Time, string) []*vcsinfo.IndexCommit); ok {
		r0 = rf(ctx, start, end, branch)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*vcsinfo.IndexCommit)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, time.Time, time.Time, string) error); ok {
		r1 = rf(ctx, start, end, branch)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RangeN provides a mock function with given fields: ctx, startIndex, endIndex, branch
func (_m *GitStore) RangeN(ctx context.Context, startIndex int, endIndex int, branch string) ([]*vcsinfo.IndexCommit, error) {
	ret := _m.Called(ctx, startIndex, endIndex, branch)

	var r0 []*vcsinfo.IndexCommit
	if rf, ok := ret.Get(0).(func(context.Context, int, int, string) []*vcsinfo.IndexCommit); ok {
		r0 = rf(ctx, startIndex, endIndex, branch)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*vcsinfo.IndexCommit)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int, int, string) error); ok {
		r1 = rf(ctx, startIndex, endIndex, branch)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}