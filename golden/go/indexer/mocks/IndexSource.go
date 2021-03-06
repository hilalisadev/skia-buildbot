// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	indexer "go.skia.org/infra/golden/go/indexer"
)

// IndexSource is an autogenerated mock type for the IndexSource type
type IndexSource struct {
	mock.Mock
}

// GetIndex provides a mock function with given fields:
func (_m *IndexSource) GetIndex() indexer.IndexSearcher {
	ret := _m.Called()

	var r0 indexer.IndexSearcher
	if rf, ok := ret.Get(0).(func() indexer.IndexSearcher); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(indexer.IndexSearcher)
		}
	}

	return r0
}
