// Code generated by mockery. DO NOT EDIT.

package mocks_txbuilder

import (
	bind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	common "github.com/ethereum/go-ethereum/common"

	mock "github.com/stretchr/testify/mock"

	polygonvalidiumetrog "github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonvalidiumetrog"

	types "github.com/ethereum/go-ethereum/core/types"
)

// rollupBananaZKEVMContractor is an autogenerated mock type for the rollupBananaZKEVMContractor type
type rollupBananaZKEVMContractor struct {
	mock.Mock
}

type rollupBananaZKEVMContractor_Expecter struct {
	mock *mock.Mock
}

func (_m *rollupBananaZKEVMContractor) EXPECT() *rollupBananaZKEVMContractor_Expecter {
	return &rollupBananaZKEVMContractor_Expecter{mock: &_m.Mock}
}

// LastAccInputHash provides a mock function with given fields: opts
func (_m *rollupBananaZKEVMContractor) LastAccInputHash(opts *bind.CallOpts) ([32]byte, error) {
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

// rollupBananaZKEVMContractor_LastAccInputHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LastAccInputHash'
type rollupBananaZKEVMContractor_LastAccInputHash_Call struct {
	*mock.Call
}

// LastAccInputHash is a helper method to define mock.On call
//   - opts *bind.CallOpts
func (_e *rollupBananaZKEVMContractor_Expecter) LastAccInputHash(opts interface{}) *rollupBananaZKEVMContractor_LastAccInputHash_Call {
	return &rollupBananaZKEVMContractor_LastAccInputHash_Call{Call: _e.mock.On("LastAccInputHash", opts)}
}

func (_c *rollupBananaZKEVMContractor_LastAccInputHash_Call) Run(run func(opts *bind.CallOpts)) *rollupBananaZKEVMContractor_LastAccInputHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*bind.CallOpts))
	})
	return _c
}

func (_c *rollupBananaZKEVMContractor_LastAccInputHash_Call) Return(_a0 [32]byte, _a1 error) *rollupBananaZKEVMContractor_LastAccInputHash_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *rollupBananaZKEVMContractor_LastAccInputHash_Call) RunAndReturn(run func(*bind.CallOpts) ([32]byte, error)) *rollupBananaZKEVMContractor_LastAccInputHash_Call {
	_c.Call.Return(run)
	return _c
}

// SequenceBatches provides a mock function with given fields: opts, batches, indexL1InfoRoot, maxSequenceTimestamp, expectedFinalAccInputHash, l2Coinbase
func (_m *rollupBananaZKEVMContractor) SequenceBatches(opts *bind.TransactOpts, batches []polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, indexL1InfoRoot uint32, maxSequenceTimestamp uint64, expectedFinalAccInputHash [32]byte, l2Coinbase common.Address) (*types.Transaction, error) {
	ret := _m.Called(opts, batches, indexL1InfoRoot, maxSequenceTimestamp, expectedFinalAccInputHash, l2Coinbase)

	if len(ret) == 0 {
		panic("no return value specified for SequenceBatches")
	}

	var r0 *types.Transaction
	var r1 error
	if rf, ok := ret.Get(0).(func(*bind.TransactOpts, []polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, uint32, uint64, [32]byte, common.Address) (*types.Transaction, error)); ok {
		return rf(opts, batches, indexL1InfoRoot, maxSequenceTimestamp, expectedFinalAccInputHash, l2Coinbase)
	}
	if rf, ok := ret.Get(0).(func(*bind.TransactOpts, []polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, uint32, uint64, [32]byte, common.Address) *types.Transaction); ok {
		r0 = rf(opts, batches, indexL1InfoRoot, maxSequenceTimestamp, expectedFinalAccInputHash, l2Coinbase)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Transaction)
		}
	}

	if rf, ok := ret.Get(1).(func(*bind.TransactOpts, []polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, uint32, uint64, [32]byte, common.Address) error); ok {
		r1 = rf(opts, batches, indexL1InfoRoot, maxSequenceTimestamp, expectedFinalAccInputHash, l2Coinbase)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// rollupBananaZKEVMContractor_SequenceBatches_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SequenceBatches'
type rollupBananaZKEVMContractor_SequenceBatches_Call struct {
	*mock.Call
}

// SequenceBatches is a helper method to define mock.On call
//   - opts *bind.TransactOpts
//   - batches []polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData
//   - indexL1InfoRoot uint32
//   - maxSequenceTimestamp uint64
//   - expectedFinalAccInputHash [32]byte
//   - l2Coinbase common.Address
func (_e *rollupBananaZKEVMContractor_Expecter) SequenceBatches(opts interface{}, batches interface{}, indexL1InfoRoot interface{}, maxSequenceTimestamp interface{}, expectedFinalAccInputHash interface{}, l2Coinbase interface{}) *rollupBananaZKEVMContractor_SequenceBatches_Call {
	return &rollupBananaZKEVMContractor_SequenceBatches_Call{Call: _e.mock.On("SequenceBatches", opts, batches, indexL1InfoRoot, maxSequenceTimestamp, expectedFinalAccInputHash, l2Coinbase)}
}

func (_c *rollupBananaZKEVMContractor_SequenceBatches_Call) Run(run func(opts *bind.TransactOpts, batches []polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, indexL1InfoRoot uint32, maxSequenceTimestamp uint64, expectedFinalAccInputHash [32]byte, l2Coinbase common.Address)) *rollupBananaZKEVMContractor_SequenceBatches_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*bind.TransactOpts), args[1].([]polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData), args[2].(uint32), args[3].(uint64), args[4].([32]byte), args[5].(common.Address))
	})
	return _c
}

func (_c *rollupBananaZKEVMContractor_SequenceBatches_Call) Return(_a0 *types.Transaction, _a1 error) *rollupBananaZKEVMContractor_SequenceBatches_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *rollupBananaZKEVMContractor_SequenceBatches_Call) RunAndReturn(run func(*bind.TransactOpts, []polygonvalidiumetrog.PolygonRollupBaseEtrogBatchData, uint32, uint64, [32]byte, common.Address) (*types.Transaction, error)) *rollupBananaZKEVMContractor_SequenceBatches_Call {
	_c.Call.Return(run)
	return _c
}

// newRollupBananaZKEVMContractor creates a new instance of rollupBananaZKEVMContractor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newRollupBananaZKEVMContractor(t interface {
	mock.TestingT
	Cleanup(func())
}) *rollupBananaZKEVMContractor {
	mock := &rollupBananaZKEVMContractor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}