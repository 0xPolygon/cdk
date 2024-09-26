// Code generated by mockery v2.40.1. DO NOT EDIT.

package sequencesender

import (
	datastreamer "github.com/0xPolygonHermez/zkevm-data-streamer/datastreamer"
	mock "github.com/stretchr/testify/mock"
)

// StreamClientMock is an autogenerated mock type for the StreamClient type
type StreamClientMock struct {
	mock.Mock
}

type StreamClientMock_Expecter struct {
	mock *mock.Mock
}

func (_m *StreamClientMock) EXPECT() *StreamClientMock_Expecter {
	return &StreamClientMock_Expecter{mock: &_m.Mock}
}

// ExecCommandStartBookmark provides a mock function with given fields: bookmark
func (_m *StreamClientMock) ExecCommandStartBookmark(bookmark []byte) error {
	ret := _m.Called(bookmark)

	if len(ret) == 0 {
		panic("no return value specified for ExecCommandStartBookmark")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte) error); ok {
		r0 = rf(bookmark)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StreamClientMock_ExecCommandStartBookmark_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExecCommandStartBookmark'
type StreamClientMock_ExecCommandStartBookmark_Call struct {
	*mock.Call
}

// ExecCommandStartBookmark is a helper method to define mock.On call
//   - bookmark []byte
func (_e *StreamClientMock_Expecter) ExecCommandStartBookmark(bookmark interface{}) *StreamClientMock_ExecCommandStartBookmark_Call {
	return &StreamClientMock_ExecCommandStartBookmark_Call{Call: _e.mock.On("ExecCommandStartBookmark", bookmark)}
}

func (_c *StreamClientMock_ExecCommandStartBookmark_Call) Run(run func(bookmark []byte)) *StreamClientMock_ExecCommandStartBookmark_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]byte))
	})
	return _c
}

func (_c *StreamClientMock_ExecCommandStartBookmark_Call) Return(_a0 error) *StreamClientMock_ExecCommandStartBookmark_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *StreamClientMock_ExecCommandStartBookmark_Call) RunAndReturn(run func([]byte) error) *StreamClientMock_ExecCommandStartBookmark_Call {
	_c.Call.Return(run)
	return _c
}

// SetProcessEntryFunc provides a mock function with given fields: f
func (_m *StreamClientMock) SetProcessEntryFunc(f datastreamer.ProcessEntryFunc) {
	_m.Called(f)
}

// StreamClientMock_SetProcessEntryFunc_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetProcessEntryFunc'
type StreamClientMock_SetProcessEntryFunc_Call struct {
	*mock.Call
}

// SetProcessEntryFunc is a helper method to define mock.On call
//   - f datastreamer.ProcessEntryFunc
func (_e *StreamClientMock_Expecter) SetProcessEntryFunc(f interface{}) *StreamClientMock_SetProcessEntryFunc_Call {
	return &StreamClientMock_SetProcessEntryFunc_Call{Call: _e.mock.On("SetProcessEntryFunc", f)}
}

func (_c *StreamClientMock_SetProcessEntryFunc_Call) Run(run func(f datastreamer.ProcessEntryFunc)) *StreamClientMock_SetProcessEntryFunc_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(datastreamer.ProcessEntryFunc))
	})
	return _c
}

func (_c *StreamClientMock_SetProcessEntryFunc_Call) Return() *StreamClientMock_SetProcessEntryFunc_Call {
	_c.Call.Return()
	return _c
}

func (_c *StreamClientMock_SetProcessEntryFunc_Call) RunAndReturn(run func(datastreamer.ProcessEntryFunc)) *StreamClientMock_SetProcessEntryFunc_Call {
	_c.Call.Return(run)
	return _c
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

// StreamClientMock_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type StreamClientMock_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
func (_e *StreamClientMock_Expecter) Start() *StreamClientMock_Start_Call {
	return &StreamClientMock_Start_Call{Call: _e.mock.On("Start")}
}

func (_c *StreamClientMock_Start_Call) Run(run func()) *StreamClientMock_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *StreamClientMock_Start_Call) Return(_a0 error) *StreamClientMock_Start_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *StreamClientMock_Start_Call) RunAndReturn(run func() error) *StreamClientMock_Start_Call {
	_c.Call.Return(run)
	return _c
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
