// Code generated by mockery. DO NOT EDIT.

package mocks_txbuilder

import (
	bind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	mock "github.com/stretchr/testify/mock"
)

// rollupBananaBaseContractor is an autogenerated mock type for the rollupBananaBaseContractor type
type rollupBananaBaseContractor struct {
	mock.Mock
}

type rollupBananaBaseContractor_Expecter struct {
	mock *mock.Mock
}

func (_m *rollupBananaBaseContractor) EXPECT() *rollupBananaBaseContractor_Expecter {
	return &rollupBananaBaseContractor_Expecter{mock: &_m.Mock}
}

// LastAccInputHash provides a mock function with given fields: opts
func (_m *rollupBananaBaseContractor) LastAccInputHash(opts *bind.CallOpts) ([32]byte, error) {
	ret := _m.Called(opts)

	if len(ret) == 0 {
		panic("no return value specified for LastAccInputHash")
	}

	var r0 [32]byte
	var r1 error
	if rf, ok := ret.Get(0).(func(*bind.CallOpts) ([32]byte, error)); ok {
		return rf(opts)
	}
	if rf, ok := ret.Get(0).(func(*bind.CallOpts) [32]byte); ok {
		r0 = rf(opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([32]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(*bind.CallOpts) error); ok {
		r1 = rf(opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// rollupBananaBaseContractor_LastAccInputHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LastAccInputHash'
type rollupBananaBaseContractor_LastAccInputHash_Call struct {
	*mock.Call
}

// LastAccInputHash is a helper method to define mock.On call
//   - opts *bind.CallOpts
func (_e *rollupBananaBaseContractor_Expecter) LastAccInputHash(opts interface{}) *rollupBananaBaseContractor_LastAccInputHash_Call {
	return &rollupBananaBaseContractor_LastAccInputHash_Call{Call: _e.mock.On("LastAccInputHash", opts)}
}

func (_c *rollupBananaBaseContractor_LastAccInputHash_Call) Run(run func(opts *bind.CallOpts)) *rollupBananaBaseContractor_LastAccInputHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*bind.CallOpts))
	})
	return _c
}

func (_c *rollupBananaBaseContractor_LastAccInputHash_Call) Return(_a0 [32]byte, _a1 error) *rollupBananaBaseContractor_LastAccInputHash_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *rollupBananaBaseContractor_LastAccInputHash_Call) RunAndReturn(run func(*bind.CallOpts) ([32]byte, error)) *rollupBananaBaseContractor_LastAccInputHash_Call {
	_c.Call.Return(run)
	return _c
}

// newRollupBananaBaseContractor creates a new instance of rollupBananaBaseContractor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newRollupBananaBaseContractor(t interface {
	mock.TestingT
	Cleanup(func())
}) *rollupBananaBaseContractor {
	mock := &rollupBananaBaseContractor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
