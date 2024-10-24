// Code generated by mockery v2.39.0. DO NOT EDIT.

package mocks

import (
	datastreamer "github.com/0xPolygonHermez/zkevm-data-streamer/datastreamer"
	mock "github.com/stretchr/testify/mock"
)

// StreamClientMock is an autogenerated mock type for the StreamClient type
type StreamClientMock struct {
	mock.Mock
}

// ExecCommandGetBookmark provides a mock function with given fields: fromBookmark
func (_m *StreamClientMock) ExecCommandGetBookmark(fromBookmark []byte) (datastreamer.FileEntry, error) {
	ret := _m.Called(fromBookmark)

	if len(ret) == 0 {
		panic("no return value specified for ExecCommandGetBookmark")
	}

	var r0 datastreamer.FileEntry
	var r1 error
	if rf, ok := ret.Get(0).(func([]byte) (datastreamer.FileEntry, error)); ok {
		return rf(fromBookmark)
	}
	if rf, ok := ret.Get(0).(func([]byte) datastreamer.FileEntry); ok {
		r0 = rf(fromBookmark)
	} else {
		r0 = ret.Get(0).(datastreamer.FileEntry)
	}

	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(fromBookmark)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ExecCommandGetEntry provides a mock function with given fields: fromEntry
func (_m *StreamClientMock) ExecCommandGetEntry(fromEntry uint64) (datastreamer.FileEntry, error) {
	ret := _m.Called(fromEntry)

	if len(ret) == 0 {
		panic("no return value specified for ExecCommandGetEntry")
	}

	var r0 datastreamer.FileEntry
	var r1 error
	if rf, ok := ret.Get(0).(func(uint64) (datastreamer.FileEntry, error)); ok {
		return rf(fromEntry)
	}
	if rf, ok := ret.Get(0).(func(uint64) datastreamer.FileEntry); ok {
		r0 = rf(fromEntry)
	} else {
		r0 = ret.Get(0).(datastreamer.FileEntry)
	}

	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(fromEntry)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ExecCommandGetHeader provides a mock function with given fields:
func (_m *StreamClientMock) ExecCommandGetHeader() (datastreamer.HeaderEntry, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ExecCommandGetHeader")
	}

	var r0 datastreamer.HeaderEntry
	var r1 error
	if rf, ok := ret.Get(0).(func() (datastreamer.HeaderEntry, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() datastreamer.HeaderEntry); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(datastreamer.HeaderEntry)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ExecCommandStart provides a mock function with given fields: fromEntry
func (_m *StreamClientMock) ExecCommandStart(fromEntry uint64) error {
	ret := _m.Called(fromEntry)

	if len(ret) == 0 {
		panic("no return value specified for ExecCommandStart")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(uint64) error); ok {
		r0 = rf(fromEntry)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ExecCommandStartBookmark provides a mock function with given fields: fromBookmark
func (_m *StreamClientMock) ExecCommandStartBookmark(fromBookmark []byte) error {
	ret := _m.Called(fromBookmark)

	if len(ret) == 0 {
		panic("no return value specified for ExecCommandStartBookmark")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte) error); ok {
		r0 = rf(fromBookmark)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ExecCommandStop provides a mock function with given fields:
func (_m *StreamClientMock) ExecCommandStop() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ExecCommandStop")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetFromStream provides a mock function with given fields:
func (_m *StreamClientMock) GetFromStream() uint64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetFromStream")
	}

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetTotalEntries provides a mock function with given fields:
func (_m *StreamClientMock) GetTotalEntries() uint64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetTotalEntries")
	}

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// IsStarted provides a mock function with given fields:
func (_m *StreamClientMock) IsStarted() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsStarted")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ResetProcessEntryFunc provides a mock function with given fields:
func (_m *StreamClientMock) ResetProcessEntryFunc() {
	_m.Called()
}

// SetProcessEntryFunc provides a mock function with given fields: f
func (_m *StreamClientMock) SetProcessEntryFunc(f datastreamer.ProcessEntryFunc) {
	_m.Called(f)
}

// Start provides a mock function with given fields:
func (_m *StreamClientMock) Start() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Start")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewStreamClientMock creates a new instance of StreamClientMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStreamClientMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *StreamClientMock {
	mock := &StreamClientMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
