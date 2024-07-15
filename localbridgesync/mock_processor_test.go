// Code generated by mockery v2.22.1. DO NOT EDIT.

package localbridgesync

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// ProcessorMock is an autogenerated mock type for the processorInterface type
type ProcessorMock struct {
	mock.Mock
}

// getLastProcessedBlock provides a mock function with given fields: ctx
func (_m *ProcessorMock) getLastProcessedBlock(ctx context.Context) (uint64, error) {
	ret := _m.Called(ctx)

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (uint64, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) uint64); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// reorg provides a mock function with given fields: firstReorgedBlock
func (_m *ProcessorMock) reorg(firstReorgedBlock uint64) error {
	ret := _m.Called(firstReorgedBlock)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint64) error); ok {
		r0 = rf(firstReorgedBlock)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// storeBridgeEvents provides a mock function with given fields: blockNum, block
func (_m *ProcessorMock) storeBridgeEvents(blockNum uint64, block bridgeEvents) error {
	ret := _m.Called(blockNum, block)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint64, bridgeEvents) error); ok {
		r0 = rf(blockNum, block)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewProcessorMock interface {
	mock.TestingT
	Cleanup(func())
}

// NewProcessorMock creates a new instance of ProcessorMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewProcessorMock(t mockConstructorTestingTNewProcessorMock) *ProcessorMock {
	mock := &ProcessorMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}