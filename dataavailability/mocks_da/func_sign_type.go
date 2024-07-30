// Code generated by mockery. DO NOT EDIT.

package mocks_da

import (
	client "github.com/0xPolygon/cdk-data-availability/client"

	mock "github.com/stretchr/testify/mock"
)

// funcSignType is an autogenerated mock type for the funcSignType type
type funcSignType struct {
	mock.Mock
}

type funcSignType_Expecter struct {
	mock *mock.Mock
}

func (_m *funcSignType) EXPECT() *funcSignType_Expecter {
	return &funcSignType_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: c
func (_m *funcSignType) Execute(c client.Client) ([]byte, error) {
	ret := _m.Called(c)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(client.Client) ([]byte, error)); ok {
		return rf(c)
	}
	if rf, ok := ret.Get(0).(func(client.Client) []byte); ok {
		r0 = rf(c)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(client.Client) error); ok {
		r1 = rf(c)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// funcSignType_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type funcSignType_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - c client.Client
func (_e *funcSignType_Expecter) Execute(c interface{}) *funcSignType_Execute_Call {
	return &funcSignType_Execute_Call{Call: _e.mock.On("Execute", c)}
}

func (_c *funcSignType_Execute_Call) Run(run func(c client.Client)) *funcSignType_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(client.Client))
	})
	return _c
}

func (_c *funcSignType_Execute_Call) Return(_a0 []byte, _a1 error) *funcSignType_Execute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *funcSignType_Execute_Call) RunAndReturn(run func(client.Client) ([]byte, error)) *funcSignType_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// newFuncSignType creates a new instance of funcSignType. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newFuncSignType(t interface {
	mock.TestingT
	Cleanup(func())
}) *funcSignType {
	mock := &funcSignType{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
