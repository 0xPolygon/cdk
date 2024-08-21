// Code generated by mockery v2.40.1. DO NOT EDIT.

package sync

import (
	context "context"

	types "github.com/ethereum/go-ethereum/core/types"
	mock "github.com/stretchr/testify/mock"
)

// EVMDownloaderMock is an autogenerated mock type for the evmDownloaderFull type
type EVMDownloaderMock struct {
	mock.Mock
}

// Download provides a mock function with given fields: ctx, fromBlock, downloadedCh
func (_m *EVMDownloaderMock) Download(ctx context.Context, fromBlock uint64, downloadedCh chan EVMBlock) {
	_m.Called(ctx, fromBlock, downloadedCh)
}

// GetBlockHeader provides a mock function with given fields: ctx, blockNum
func (_m *EVMDownloaderMock) GetBlockHeader(ctx context.Context, blockNum uint64) EVMBlockHeader {
	ret := _m.Called(ctx, blockNum)

	if len(ret) == 0 {
		panic("no return value specified for getBlockHeader")
	}

	var r0 EVMBlockHeader
	if rf, ok := ret.Get(0).(func(context.Context, uint64) EVMBlockHeader); ok {
		r0 = rf(ctx, blockNum)
	} else {
		r0 = ret.Get(0).(EVMBlockHeader)
	}

	return r0
}

// GetEventsByBlockRange provides a mock function with given fields: ctx, fromBlock, toBlock
func (_m *EVMDownloaderMock) GetEventsByBlockRange(ctx context.Context, fromBlock uint64, toBlock uint64) []EVMBlock {
	ret := _m.Called(ctx, fromBlock, toBlock)

	if len(ret) == 0 {
		panic("no return value specified for getEventsByBlockRange")
	}

	var r0 []EVMBlock
	if rf, ok := ret.Get(0).(func(context.Context, uint64, uint64) []EVMBlock); ok {
		r0 = rf(ctx, fromBlock, toBlock)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]EVMBlock)
		}
	}

	return r0
}

// GetLogs provides a mock function with given fields: ctx, fromBlock, toBlock
func (_m *EVMDownloaderMock) GetLogs(ctx context.Context, fromBlock uint64, toBlock uint64) []types.Log {
	ret := _m.Called(ctx, fromBlock, toBlock)

	if len(ret) == 0 {
		panic("no return value specified for getLogs")
	}

	var r0 []types.Log
	if rf, ok := ret.Get(0).(func(context.Context, uint64, uint64) []types.Log); ok {
		r0 = rf(ctx, fromBlock, toBlock)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.Log)
		}
	}

	return r0
}

// WaitForNewBlocks provides a mock function with given fields: ctx, lastBlockSeen
func (_m *EVMDownloaderMock) WaitForNewBlocks(ctx context.Context, lastBlockSeen uint64) uint64 {
	ret := _m.Called(ctx, lastBlockSeen)

	if len(ret) == 0 {
		panic("no return value specified for waitForNewBlocks")
	}

	var r0 uint64
	if rf, ok := ret.Get(0).(func(context.Context, uint64) uint64); ok {
		r0 = rf(ctx, lastBlockSeen)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// NewEVMDownloaderMock creates a new instance of EVMDownloaderMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEVMDownloaderMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *EVMDownloaderMock {
	mock := &EVMDownloaderMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
