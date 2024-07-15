// Code generated by mockery v2.22.1. DO NOT EDIT.

package localbridgesync

import (
	context "context"

	types "github.com/ethereum/go-ethereum/core/types"
	mock "github.com/stretchr/testify/mock"
)

// DownloaderMock is an autogenerated mock type for the downloaderInterface type
type DownloaderMock struct {
	mock.Mock
}

// appendLog provides a mock function with given fields: b, l
func (_m *DownloaderMock) appendLog(b *block, l types.Log) {
	_m.Called(b, l)
}

// getBlockHeader provides a mock function with given fields: ctx, blockNum
func (_m *DownloaderMock) getBlockHeader(ctx context.Context, blockNum uint64) blockHeader {
	ret := _m.Called(ctx, blockNum)

	var r0 blockHeader
	if rf, ok := ret.Get(0).(func(context.Context, uint64) blockHeader); ok {
		r0 = rf(ctx, blockNum)
	} else {
		r0 = ret.Get(0).(blockHeader)
	}

	return r0
}

// getEventsByBlockRange provides a mock function with given fields: ctx, fromBlock, toBlock
func (_m *DownloaderMock) getEventsByBlockRange(ctx context.Context, fromBlock uint64, toBlock uint64) []block {
	ret := _m.Called(ctx, fromBlock, toBlock)

	var r0 []block
	if rf, ok := ret.Get(0).(func(context.Context, uint64, uint64) []block); ok {
		r0 = rf(ctx, fromBlock, toBlock)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]block)
		}
	}

	return r0
}

// getLogs provides a mock function with given fields: ctx, fromBlock, toBlock
func (_m *DownloaderMock) getLogs(ctx context.Context, fromBlock uint64, toBlock uint64) []types.Log {
	ret := _m.Called(ctx, fromBlock, toBlock)

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

// syncBlockChunkSize provides a mock function with given fields:
func (_m *DownloaderMock) syncBlockChunkSize() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// waitForNewBlocks provides a mock function with given fields: ctx, lastBlockSeen
func (_m *DownloaderMock) waitForNewBlocks(ctx context.Context, lastBlockSeen uint64) uint64 {
	ret := _m.Called(ctx, lastBlockSeen)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(context.Context, uint64) uint64); ok {
		r0 = rf(ctx, lastBlockSeen)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

type mockConstructorTestingTNewDownloaderMock interface {
	mock.TestingT
	Cleanup(func())
}

// NewDownloaderMock creates a new instance of DownloaderMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDownloaderMock(t mockConstructorTestingTNewDownloaderMock) *DownloaderMock {
	mock := &DownloaderMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}