// Code generated by mockery. DO NOT EDIT.

package mocks_bridgesync

import (
	context "context"

	common "github.com/ethereum/go-ethereum/common"

	mock "github.com/stretchr/testify/mock"

	reorgdetector "github.com/0xPolygon/cdk/reorgdetector"
)

// ReorgDetector is an autogenerated mock type for the ReorgDetector type
type ReorgDetector struct {
	mock.Mock
}

type ReorgDetector_Expecter struct {
	mock *mock.Mock
}

func (_m *ReorgDetector) EXPECT() *ReorgDetector_Expecter {
	return &ReorgDetector_Expecter{mock: &_m.Mock}
}

// AddBlockToTrack provides a mock function with given fields: ctx, id, blockNum, blockHash
func (_m *ReorgDetector) AddBlockToTrack(ctx context.Context, id string, blockNum uint64, blockHash common.Hash) error {
	ret := _m.Called(ctx, id, blockNum, blockHash)

	if len(ret) == 0 {
		panic("no return value specified for AddBlockToTrack")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uint64, common.Hash) error); ok {
		r0 = rf(ctx, id, blockNum, blockHash)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReorgDetector_AddBlockToTrack_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddBlockToTrack'
type ReorgDetector_AddBlockToTrack_Call struct {
	*mock.Call
}

// AddBlockToTrack is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - blockNum uint64
//   - blockHash common.Hash
func (_e *ReorgDetector_Expecter) AddBlockToTrack(ctx interface{}, id interface{}, blockNum interface{}, blockHash interface{}) *ReorgDetector_AddBlockToTrack_Call {
	return &ReorgDetector_AddBlockToTrack_Call{Call: _e.mock.On("AddBlockToTrack", ctx, id, blockNum, blockHash)}
}

func (_c *ReorgDetector_AddBlockToTrack_Call) Run(run func(ctx context.Context, id string, blockNum uint64, blockHash common.Hash)) *ReorgDetector_AddBlockToTrack_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(uint64), args[3].(common.Hash))
	})
	return _c
}

func (_c *ReorgDetector_AddBlockToTrack_Call) Return(_a0 error) *ReorgDetector_AddBlockToTrack_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ReorgDetector_AddBlockToTrack_Call) RunAndReturn(run func(context.Context, string, uint64, common.Hash) error) *ReorgDetector_AddBlockToTrack_Call {
	_c.Call.Return(run)
	return _c
}

// Subscribe provides a mock function with given fields: id
func (_m *ReorgDetector) Subscribe(id string) (*reorgdetector.Subscription, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for Subscribe")
	}

	var r0 *reorgdetector.Subscription
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*reorgdetector.Subscription, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(string) *reorgdetector.Subscription); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*reorgdetector.Subscription)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReorgDetector_Subscribe_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Subscribe'
type ReorgDetector_Subscribe_Call struct {
	*mock.Call
}

// Subscribe is a helper method to define mock.On call
//   - id string
func (_e *ReorgDetector_Expecter) Subscribe(id interface{}) *ReorgDetector_Subscribe_Call {
	return &ReorgDetector_Subscribe_Call{Call: _e.mock.On("Subscribe", id)}
}

func (_c *ReorgDetector_Subscribe_Call) Run(run func(id string)) *ReorgDetector_Subscribe_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ReorgDetector_Subscribe_Call) Return(_a0 *reorgdetector.Subscription, _a1 error) *ReorgDetector_Subscribe_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ReorgDetector_Subscribe_Call) RunAndReturn(run func(string) (*reorgdetector.Subscription, error)) *ReorgDetector_Subscribe_Call {
	_c.Call.Return(run)
	return _c
}

// NewReorgDetector creates a new instance of ReorgDetector. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewReorgDetector(t interface {
	mock.TestingT
	Cleanup(func())
}) *ReorgDetector {
	mock := &ReorgDetector{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
