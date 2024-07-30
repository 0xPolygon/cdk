// Code generated by mockery. DO NOT EDIT.

package mocks_txbuilder

import (
	bind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	common "github.com/ethereum/go-ethereum/common"

	mock "github.com/stretchr/testify/mock"

	polygonvalidiumetrog "github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonvalidiumetrog"

	types "github.com/ethereum/go-ethereum/core/types"
)

// rollupBananaValidiumContractor is an autogenerated mock type for the rollupBananaValidiumContractor type
type rollupBananaValidiumContractor struct {
	mock.Mock
}

type rollupBananaValidiumContractor_Expecter struct {
	mock *mock.Mock
}

func (_m *rollupBananaValidiumContractor) EXPECT() *rollupBananaValidiumContractor_Expecter {
	return &rollupBananaValidiumContractor_Expecter{mock: &_m.Mock}
}

// LastAccInputHash provides a mock function with given fields: opts
func (_m *rollupBananaValidiumContractor) LastAccInputHash(opts *bind.CallOpts) ([32]byte, error) {
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

// rollupBananaValidiumContractor_LastAccInputHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LastAccInputHash'
type rollupBananaValidiumContractor_LastAccInputHash_Call struct {
	*mock.Call
}

// LastAccInputHash is a helper method to define mock.On call
//   - opts *bind.CallOpts
func (_e *rollupBananaValidiumContractor_Expecter) LastAccInputHash(opts interface{}) *rollupBananaValidiumContractor_LastAccInputHash_Call {
	return &rollupBananaValidiumContractor_LastAccInputHash_Call{Call: _e.mock.On("LastAccInputHash", opts)}
}

func (_c *rollupBananaValidiumContractor_LastAccInputHash_Call) Run(run func(opts *bind.CallOpts)) *rollupBananaValidiumContractor_LastAccInputHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*bind.CallOpts))
	})
	return _c
}

func (_c *rollupBananaValidiumContractor_LastAccInputHash_Call) Return(_a0 [32]byte, _a1 error) *rollupBananaValidiumContractor_LastAccInputHash_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *rollupBananaValidiumContractor_LastAccInputHash_Call) RunAndReturn(run func(*bind.CallOpts) ([32]byte, error)) *rollupBananaValidiumContractor_LastAccInputHash_Call {
	_c.Call.Return(run)
	return _c
}

// SequenceBatchesValidium provides a mock function with given fields: opts, batches, indexL1InfoRoot, maxSequenceTimestamp, expectedFinalAccInputHash, l2Coinbase, dataAvailabilityMessage
func (_m *rollupBananaValidiumContractor) SequenceBatchesValidium(opts *bind.TransactOpts, batches []polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData, indexL1InfoRoot uint32, maxSequenceTimestamp uint64, expectedFinalAccInputHash [32]byte, l2Coinbase common.Address, dataAvailabilityMessage []byte) (*types.Transaction, error) {
	ret := _m.Called(opts, batches, indexL1InfoRoot, maxSequenceTimestamp, expectedFinalAccInputHash, l2Coinbase, dataAvailabilityMessage)

	if len(ret) == 0 {
		panic("no return value specified for SequenceBatchesValidium")
	}

	var r0 *types.Transaction
	var r1 error
	if rf, ok := ret.Get(0).(func(*bind.TransactOpts, []polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData, uint32, uint64, [32]byte, common.Address, []byte) (*types.Transaction, error)); ok {
		return rf(opts, batches, indexL1InfoRoot, maxSequenceTimestamp, expectedFinalAccInputHash, l2Coinbase, dataAvailabilityMessage)
	}
	if rf, ok := ret.Get(0).(func(*bind.TransactOpts, []polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData, uint32, uint64, [32]byte, common.Address, []byte) *types.Transaction); ok {
		r0 = rf(opts, batches, indexL1InfoRoot, maxSequenceTimestamp, expectedFinalAccInputHash, l2Coinbase, dataAvailabilityMessage)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Transaction)
		}
	}

	if rf, ok := ret.Get(1).(func(*bind.TransactOpts, []polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData, uint32, uint64, [32]byte, common.Address, []byte) error); ok {
		r1 = rf(opts, batches, indexL1InfoRoot, maxSequenceTimestamp, expectedFinalAccInputHash, l2Coinbase, dataAvailabilityMessage)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// rollupBananaValidiumContractor_SequenceBatchesValidium_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SequenceBatchesValidium'
type rollupBananaValidiumContractor_SequenceBatchesValidium_Call struct {
	*mock.Call
}

// SequenceBatchesValidium is a helper method to define mock.On call
//   - opts *bind.TransactOpts
//   - batches []polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData
//   - indexL1InfoRoot uint32
//   - maxSequenceTimestamp uint64
//   - expectedFinalAccInputHash [32]byte
//   - l2Coinbase common.Address
//   - dataAvailabilityMessage []byte
func (_e *rollupBananaValidiumContractor_Expecter) SequenceBatchesValidium(opts interface{}, batches interface{}, indexL1InfoRoot interface{}, maxSequenceTimestamp interface{}, expectedFinalAccInputHash interface{}, l2Coinbase interface{}, dataAvailabilityMessage interface{}) *rollupBananaValidiumContractor_SequenceBatchesValidium_Call {
	return &rollupBananaValidiumContractor_SequenceBatchesValidium_Call{Call: _e.mock.On("SequenceBatchesValidium", opts, batches, indexL1InfoRoot, maxSequenceTimestamp, expectedFinalAccInputHash, l2Coinbase, dataAvailabilityMessage)}
}

func (_c *rollupBananaValidiumContractor_SequenceBatchesValidium_Call) Run(run func(opts *bind.TransactOpts, batches []polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData, indexL1InfoRoot uint32, maxSequenceTimestamp uint64, expectedFinalAccInputHash [32]byte, l2Coinbase common.Address, dataAvailabilityMessage []byte)) *rollupBananaValidiumContractor_SequenceBatchesValidium_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*bind.TransactOpts), args[1].([]polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData), args[2].(uint32), args[3].(uint64), args[4].([32]byte), args[5].(common.Address), args[6].([]byte))
	})
	return _c
}

func (_c *rollupBananaValidiumContractor_SequenceBatchesValidium_Call) Return(_a0 *types.Transaction, _a1 error) *rollupBananaValidiumContractor_SequenceBatchesValidium_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *rollupBananaValidiumContractor_SequenceBatchesValidium_Call) RunAndReturn(run func(*bind.TransactOpts, []polygonvalidiumetrog.PolygonValidiumEtrogValidiumBatchData, uint32, uint64, [32]byte, common.Address, []byte) (*types.Transaction, error)) *rollupBananaValidiumContractor_SequenceBatchesValidium_Call {
	_c.Call.Return(run)
	return _c
}

// newRollupBananaValidiumContractor creates a new instance of rollupBananaValidiumContractor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newRollupBananaValidiumContractor(t interface {
	mock.TestingT
	Cleanup(func())
}) *rollupBananaValidiumContractor {
	mock := &rollupBananaValidiumContractor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}