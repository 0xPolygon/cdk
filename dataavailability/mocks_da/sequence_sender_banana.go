// Code generated by mockery. DO NOT EDIT.

package mocks_da

import (
	context "context"

	etherman "github.com/0xPolygon/cdk/etherman"

	mock "github.com/stretchr/testify/mock"
)

// SequenceSenderBanana is an autogenerated mock type for the SequenceSenderBanana type
type SequenceSenderBanana struct {
	mock.Mock
}

type SequenceSenderBanana_Expecter struct {
	mock *mock.Mock
}

func (_m *SequenceSenderBanana) EXPECT() *SequenceSenderBanana_Expecter {
	return &SequenceSenderBanana_Expecter{mock: &_m.Mock}
}

// PostSequenceBanana provides a mock function with given fields: ctx, sequence
func (_m *SequenceSenderBanana) PostSequenceBanana(ctx context.Context, sequence etherman.SequenceBanana) ([]byte, error) {
	ret := _m.Called(ctx, sequence)

	if len(ret) == 0 {
		panic("no return value specified for PostSequenceBanana")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, etherman.SequenceBanana) ([]byte, error)); ok {
		return rf(ctx, sequence)
	}
	if rf, ok := ret.Get(0).(func(context.Context, etherman.SequenceBanana) []byte); ok {
		r0 = rf(ctx, sequence)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, etherman.SequenceBanana) error); ok {
		r1 = rf(ctx, sequence)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SequenceSenderBanana_PostSequenceBanana_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PostSequenceBanana'
type SequenceSenderBanana_PostSequenceBanana_Call struct {
	*mock.Call
}

// PostSequenceBanana is a helper method to define mock.On call
//   - ctx context.Context
//   - sequence etherman.SequenceBanana
func (_e *SequenceSenderBanana_Expecter) PostSequenceBanana(ctx interface{}, sequence interface{}) *SequenceSenderBanana_PostSequenceBanana_Call {
	return &SequenceSenderBanana_PostSequenceBanana_Call{Call: _e.mock.On("PostSequenceBanana", ctx, sequence)}
}

func (_c *SequenceSenderBanana_PostSequenceBanana_Call) Run(run func(ctx context.Context, sequence etherman.SequenceBanana)) *SequenceSenderBanana_PostSequenceBanana_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(etherman.SequenceBanana))
	})
	return _c
}

func (_c *SequenceSenderBanana_PostSequenceBanana_Call) Return(_a0 []byte, _a1 error) *SequenceSenderBanana_PostSequenceBanana_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SequenceSenderBanana_PostSequenceBanana_Call) RunAndReturn(run func(context.Context, etherman.SequenceBanana) ([]byte, error)) *SequenceSenderBanana_PostSequenceBanana_Call {
	_c.Call.Return(run)
	return _c
}

// NewSequenceSenderBanana creates a new instance of SequenceSenderBanana. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSequenceSenderBanana(t interface {
	mock.TestingT
	Cleanup(func())
}) *SequenceSenderBanana {
	mock := &SequenceSenderBanana{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
