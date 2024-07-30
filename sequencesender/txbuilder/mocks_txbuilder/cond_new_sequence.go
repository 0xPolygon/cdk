// Code generated by mockery. DO NOT EDIT.

package mocks_txbuilder

import (
	context "context"

	common "github.com/ethereum/go-ethereum/common"

	mock "github.com/stretchr/testify/mock"

	seqsendertypes "github.com/0xPolygon/cdk/sequencesender/seqsendertypes"

	txbuilder "github.com/0xPolygon/cdk/sequencesender/txbuilder"
)

// CondNewSequence is an autogenerated mock type for the CondNewSequence type
type CondNewSequence struct {
	mock.Mock
}

type CondNewSequence_Expecter struct {
	mock *mock.Mock
}

func (_m *CondNewSequence) EXPECT() *CondNewSequence_Expecter {
	return &CondNewSequence_Expecter{mock: &_m.Mock}
}

// NewSequenceIfWorthToSend provides a mock function with given fields: ctx, txBuilder, sequenceBatches, l2Coinbase
func (_m *CondNewSequence) NewSequenceIfWorthToSend(ctx context.Context, txBuilder txbuilder.TxBuilder, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address) (seqsendertypes.Sequence, error) {
	ret := _m.Called(ctx, txBuilder, sequenceBatches, l2Coinbase)

	if len(ret) == 0 {
		panic("no return value specified for NewSequenceIfWorthToSend")
	}

	var r0 seqsendertypes.Sequence
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, txbuilder.TxBuilder, []seqsendertypes.Batch, common.Address) (seqsendertypes.Sequence, error)); ok {
		return rf(ctx, txBuilder, sequenceBatches, l2Coinbase)
	}
	if rf, ok := ret.Get(0).(func(context.Context, txbuilder.TxBuilder, []seqsendertypes.Batch, common.Address) seqsendertypes.Sequence); ok {
		r0 = rf(ctx, txBuilder, sequenceBatches, l2Coinbase)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(seqsendertypes.Sequence)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, txbuilder.TxBuilder, []seqsendertypes.Batch, common.Address) error); ok {
		r1 = rf(ctx, txBuilder, sequenceBatches, l2Coinbase)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CondNewSequence_NewSequenceIfWorthToSend_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewSequenceIfWorthToSend'
type CondNewSequence_NewSequenceIfWorthToSend_Call struct {
	*mock.Call
}

// NewSequenceIfWorthToSend is a helper method to define mock.On call
//   - ctx context.Context
//   - txBuilder txbuilder.TxBuilder
//   - sequenceBatches []seqsendertypes.Batch
//   - l2Coinbase common.Address
func (_e *CondNewSequence_Expecter) NewSequenceIfWorthToSend(ctx interface{}, txBuilder interface{}, sequenceBatches interface{}, l2Coinbase interface{}) *CondNewSequence_NewSequenceIfWorthToSend_Call {
	return &CondNewSequence_NewSequenceIfWorthToSend_Call{Call: _e.mock.On("NewSequenceIfWorthToSend", ctx, txBuilder, sequenceBatches, l2Coinbase)}
}

func (_c *CondNewSequence_NewSequenceIfWorthToSend_Call) Run(run func(ctx context.Context, txBuilder txbuilder.TxBuilder, sequenceBatches []seqsendertypes.Batch, l2Coinbase common.Address)) *CondNewSequence_NewSequenceIfWorthToSend_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(txbuilder.TxBuilder), args[2].([]seqsendertypes.Batch), args[3].(common.Address))
	})
	return _c
}

func (_c *CondNewSequence_NewSequenceIfWorthToSend_Call) Return(_a0 seqsendertypes.Sequence, _a1 error) *CondNewSequence_NewSequenceIfWorthToSend_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CondNewSequence_NewSequenceIfWorthToSend_Call) RunAndReturn(run func(context.Context, txbuilder.TxBuilder, []seqsendertypes.Batch, common.Address) (seqsendertypes.Sequence, error)) *CondNewSequence_NewSequenceIfWorthToSend_Call {
	_c.Call.Return(run)
	return _c
}

// NewCondNewSequence creates a new instance of CondNewSequence. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCondNewSequence(t interface {
	mock.TestingT
	Cleanup(func())
}) *CondNewSequence {
	mock := &CondNewSequence{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}