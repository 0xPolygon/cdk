// Code generated by mockery v2.45.1. DO NOT EDIT.

package mocks

import (
	context "context"

	pgx "github.com/jackc/pgx/v4"
	mock "github.com/stretchr/testify/mock"

	state "github.com/0xPolygon/cdk/state"
)

// stateInterfaceMock is an autogenerated mock type for the StateInterface type
type stateInterfaceMock struct {
	mock.Mock
}

type stateInterfaceMock_Expecter struct {
	mock *mock.Mock
}

func (_m *stateInterfaceMock) EXPECT() *stateInterfaceMock_Expecter {
	return &stateInterfaceMock_Expecter{mock: &_m.Mock}
}

// AddBatch provides a mock function with given fields: ctx, dbBatch, dbTx
func (_m *stateInterfaceMock) AddBatch(ctx context.Context, dbBatch *state.DBBatch, dbTx pgx.Tx) error {
	ret := _m.Called(ctx, dbBatch, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for AddBatch")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.DBBatch, pgx.Tx) error); ok {
		r0 = rf(ctx, dbBatch, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// stateInterfaceMock_AddBatch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddBatch'
type stateInterfaceMock_AddBatch_Call struct {
	*mock.Call
}

// AddBatch is a helper method to define mock.On call
//   - ctx context.Context
//   - dbBatch *state.DBBatch
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) AddBatch(ctx interface{}, dbBatch interface{}, dbTx interface{}) *stateInterfaceMock_AddBatch_Call {
	return &stateInterfaceMock_AddBatch_Call{Call: _e.mock.On("AddBatch", ctx, dbBatch, dbTx)}
}

func (_c *stateInterfaceMock_AddBatch_Call) Run(run func(ctx context.Context, dbBatch *state.DBBatch, dbTx pgx.Tx)) *stateInterfaceMock_AddBatch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.DBBatch), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_AddBatch_Call) Return(_a0 error) *stateInterfaceMock_AddBatch_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *stateInterfaceMock_AddBatch_Call) RunAndReturn(run func(context.Context, *state.DBBatch, pgx.Tx) error) *stateInterfaceMock_AddBatch_Call {
	_c.Call.Return(run)
	return _c
}

// AddGeneratedProof provides a mock function with given fields: ctx, proof, dbTx
func (_m *stateInterfaceMock) AddGeneratedProof(ctx context.Context, proof *state.Proof, dbTx pgx.Tx) error {
	ret := _m.Called(ctx, proof, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for AddGeneratedProof")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Proof, pgx.Tx) error); ok {
		r0 = rf(ctx, proof, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// stateInterfaceMock_AddGeneratedProof_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddGeneratedProof'
type stateInterfaceMock_AddGeneratedProof_Call struct {
	*mock.Call
}

// AddGeneratedProof is a helper method to define mock.On call
//   - ctx context.Context
//   - proof *state.Proof
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) AddGeneratedProof(ctx interface{}, proof interface{}, dbTx interface{}) *stateInterfaceMock_AddGeneratedProof_Call {
	return &stateInterfaceMock_AddGeneratedProof_Call{Call: _e.mock.On("AddGeneratedProof", ctx, proof, dbTx)}
}

func (_c *stateInterfaceMock_AddGeneratedProof_Call) Run(run func(ctx context.Context, proof *state.Proof, dbTx pgx.Tx)) *stateInterfaceMock_AddGeneratedProof_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Proof), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_AddGeneratedProof_Call) Return(_a0 error) *stateInterfaceMock_AddGeneratedProof_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *stateInterfaceMock_AddGeneratedProof_Call) RunAndReturn(run func(context.Context, *state.Proof, pgx.Tx) error) *stateInterfaceMock_AddGeneratedProof_Call {
	_c.Call.Return(run)
	return _c
}

// AddSequence provides a mock function with given fields: ctx, sequence, dbTx
func (_m *stateInterfaceMock) AddSequence(ctx context.Context, sequence state.Sequence, dbTx pgx.Tx) error {
	ret := _m.Called(ctx, sequence, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for AddSequence")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, state.Sequence, pgx.Tx) error); ok {
		r0 = rf(ctx, sequence, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// stateInterfaceMock_AddSequence_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddSequence'
type stateInterfaceMock_AddSequence_Call struct {
	*mock.Call
}

// AddSequence is a helper method to define mock.On call
//   - ctx context.Context
//   - sequence state.Sequence
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) AddSequence(ctx interface{}, sequence interface{}, dbTx interface{}) *stateInterfaceMock_AddSequence_Call {
	return &stateInterfaceMock_AddSequence_Call{Call: _e.mock.On("AddSequence", ctx, sequence, dbTx)}
}

func (_c *stateInterfaceMock_AddSequence_Call) Run(run func(ctx context.Context, sequence state.Sequence, dbTx pgx.Tx)) *stateInterfaceMock_AddSequence_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(state.Sequence), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_AddSequence_Call) Return(_a0 error) *stateInterfaceMock_AddSequence_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *stateInterfaceMock_AddSequence_Call) RunAndReturn(run func(context.Context, state.Sequence, pgx.Tx) error) *stateInterfaceMock_AddSequence_Call {
	_c.Call.Return(run)
	return _c
}

// BeginStateTransaction provides a mock function with given fields: ctx
func (_m *stateInterfaceMock) BeginStateTransaction(ctx context.Context) (pgx.Tx, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for BeginStateTransaction")
	}

	var r0 pgx.Tx
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (pgx.Tx, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) pgx.Tx); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pgx.Tx)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// stateInterfaceMock_BeginStateTransaction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BeginStateTransaction'
type stateInterfaceMock_BeginStateTransaction_Call struct {
	*mock.Call
}

// BeginStateTransaction is a helper method to define mock.On call
//   - ctx context.Context
func (_e *stateInterfaceMock_Expecter) BeginStateTransaction(ctx interface{}) *stateInterfaceMock_BeginStateTransaction_Call {
	return &stateInterfaceMock_BeginStateTransaction_Call{Call: _e.mock.On("BeginStateTransaction", ctx)}
}

func (_c *stateInterfaceMock_BeginStateTransaction_Call) Run(run func(ctx context.Context)) *stateInterfaceMock_BeginStateTransaction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *stateInterfaceMock_BeginStateTransaction_Call) Return(_a0 pgx.Tx, _a1 error) *stateInterfaceMock_BeginStateTransaction_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *stateInterfaceMock_BeginStateTransaction_Call) RunAndReturn(run func(context.Context) (pgx.Tx, error)) *stateInterfaceMock_BeginStateTransaction_Call {
	_c.Call.Return(run)
	return _c
}

// CheckProofContainsCompleteSequences provides a mock function with given fields: ctx, proof, dbTx
func (_m *stateInterfaceMock) CheckProofContainsCompleteSequences(ctx context.Context, proof *state.Proof, dbTx pgx.Tx) (bool, error) {
	ret := _m.Called(ctx, proof, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for CheckProofContainsCompleteSequences")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Proof, pgx.Tx) (bool, error)); ok {
		return rf(ctx, proof, dbTx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *state.Proof, pgx.Tx) bool); ok {
		r0 = rf(ctx, proof, dbTx)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *state.Proof, pgx.Tx) error); ok {
		r1 = rf(ctx, proof, dbTx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// stateInterfaceMock_CheckProofContainsCompleteSequences_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CheckProofContainsCompleteSequences'
type stateInterfaceMock_CheckProofContainsCompleteSequences_Call struct {
	*mock.Call
}

// CheckProofContainsCompleteSequences is a helper method to define mock.On call
//   - ctx context.Context
//   - proof *state.Proof
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) CheckProofContainsCompleteSequences(ctx interface{}, proof interface{}, dbTx interface{}) *stateInterfaceMock_CheckProofContainsCompleteSequences_Call {
	return &stateInterfaceMock_CheckProofContainsCompleteSequences_Call{Call: _e.mock.On("CheckProofContainsCompleteSequences", ctx, proof, dbTx)}
}

func (_c *stateInterfaceMock_CheckProofContainsCompleteSequences_Call) Run(run func(ctx context.Context, proof *state.Proof, dbTx pgx.Tx)) *stateInterfaceMock_CheckProofContainsCompleteSequences_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Proof), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_CheckProofContainsCompleteSequences_Call) Return(_a0 bool, _a1 error) *stateInterfaceMock_CheckProofContainsCompleteSequences_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *stateInterfaceMock_CheckProofContainsCompleteSequences_Call) RunAndReturn(run func(context.Context, *state.Proof, pgx.Tx) (bool, error)) *stateInterfaceMock_CheckProofContainsCompleteSequences_Call {
	_c.Call.Return(run)
	return _c
}

// CheckProofExistsForBatch provides a mock function with given fields: ctx, batchNumber, dbTx
func (_m *stateInterfaceMock) CheckProofExistsForBatch(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) (bool, error) {
	ret := _m.Called(ctx, batchNumber, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for CheckProofExistsForBatch")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uint64, pgx.Tx) (bool, error)); ok {
		return rf(ctx, batchNumber, dbTx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uint64, pgx.Tx) bool); ok {
		r0 = rf(ctx, batchNumber, dbTx)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, uint64, pgx.Tx) error); ok {
		r1 = rf(ctx, batchNumber, dbTx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// stateInterfaceMock_CheckProofExistsForBatch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CheckProofExistsForBatch'
type stateInterfaceMock_CheckProofExistsForBatch_Call struct {
	*mock.Call
}

// CheckProofExistsForBatch is a helper method to define mock.On call
//   - ctx context.Context
//   - batchNumber uint64
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) CheckProofExistsForBatch(ctx interface{}, batchNumber interface{}, dbTx interface{}) *stateInterfaceMock_CheckProofExistsForBatch_Call {
	return &stateInterfaceMock_CheckProofExistsForBatch_Call{Call: _e.mock.On("CheckProofExistsForBatch", ctx, batchNumber, dbTx)}
}

func (_c *stateInterfaceMock_CheckProofExistsForBatch_Call) Run(run func(ctx context.Context, batchNumber uint64, dbTx pgx.Tx)) *stateInterfaceMock_CheckProofExistsForBatch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uint64), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_CheckProofExistsForBatch_Call) Return(_a0 bool, _a1 error) *stateInterfaceMock_CheckProofExistsForBatch_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *stateInterfaceMock_CheckProofExistsForBatch_Call) RunAndReturn(run func(context.Context, uint64, pgx.Tx) (bool, error)) *stateInterfaceMock_CheckProofExistsForBatch_Call {
	_c.Call.Return(run)
	return _c
}

// CleanupGeneratedProofs provides a mock function with given fields: ctx, batchNumber, dbTx
func (_m *stateInterfaceMock) CleanupGeneratedProofs(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) error {
	ret := _m.Called(ctx, batchNumber, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for CleanupGeneratedProofs")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uint64, pgx.Tx) error); ok {
		r0 = rf(ctx, batchNumber, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// stateInterfaceMock_CleanupGeneratedProofs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CleanupGeneratedProofs'
type stateInterfaceMock_CleanupGeneratedProofs_Call struct {
	*mock.Call
}

// CleanupGeneratedProofs is a helper method to define mock.On call
//   - ctx context.Context
//   - batchNumber uint64
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) CleanupGeneratedProofs(ctx interface{}, batchNumber interface{}, dbTx interface{}) *stateInterfaceMock_CleanupGeneratedProofs_Call {
	return &stateInterfaceMock_CleanupGeneratedProofs_Call{Call: _e.mock.On("CleanupGeneratedProofs", ctx, batchNumber, dbTx)}
}

func (_c *stateInterfaceMock_CleanupGeneratedProofs_Call) Run(run func(ctx context.Context, batchNumber uint64, dbTx pgx.Tx)) *stateInterfaceMock_CleanupGeneratedProofs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uint64), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_CleanupGeneratedProofs_Call) Return(_a0 error) *stateInterfaceMock_CleanupGeneratedProofs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *stateInterfaceMock_CleanupGeneratedProofs_Call) RunAndReturn(run func(context.Context, uint64, pgx.Tx) error) *stateInterfaceMock_CleanupGeneratedProofs_Call {
	_c.Call.Return(run)
	return _c
}

// CleanupLockedProofs provides a mock function with given fields: ctx, duration, dbTx
func (_m *stateInterfaceMock) CleanupLockedProofs(ctx context.Context, duration string, dbTx pgx.Tx) (int64, error) {
	ret := _m.Called(ctx, duration, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for CleanupLockedProofs")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, pgx.Tx) (int64, error)); ok {
		return rf(ctx, duration, dbTx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, pgx.Tx) int64); ok {
		r0 = rf(ctx, duration, dbTx)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, pgx.Tx) error); ok {
		r1 = rf(ctx, duration, dbTx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// stateInterfaceMock_CleanupLockedProofs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CleanupLockedProofs'
type stateInterfaceMock_CleanupLockedProofs_Call struct {
	*mock.Call
}

// CleanupLockedProofs is a helper method to define mock.On call
//   - ctx context.Context
//   - duration string
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) CleanupLockedProofs(ctx interface{}, duration interface{}, dbTx interface{}) *stateInterfaceMock_CleanupLockedProofs_Call {
	return &stateInterfaceMock_CleanupLockedProofs_Call{Call: _e.mock.On("CleanupLockedProofs", ctx, duration, dbTx)}
}

func (_c *stateInterfaceMock_CleanupLockedProofs_Call) Run(run func(ctx context.Context, duration string, dbTx pgx.Tx)) *stateInterfaceMock_CleanupLockedProofs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_CleanupLockedProofs_Call) Return(_a0 int64, _a1 error) *stateInterfaceMock_CleanupLockedProofs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *stateInterfaceMock_CleanupLockedProofs_Call) RunAndReturn(run func(context.Context, string, pgx.Tx) (int64, error)) *stateInterfaceMock_CleanupLockedProofs_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteBatchesNewerThanBatchNumber provides a mock function with given fields: ctx, batchNumber, dbTx
func (_m *stateInterfaceMock) DeleteBatchesNewerThanBatchNumber(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) error {
	ret := _m.Called(ctx, batchNumber, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for DeleteBatchesNewerThanBatchNumber")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uint64, pgx.Tx) error); ok {
		r0 = rf(ctx, batchNumber, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// stateInterfaceMock_DeleteBatchesNewerThanBatchNumber_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteBatchesNewerThanBatchNumber'
type stateInterfaceMock_DeleteBatchesNewerThanBatchNumber_Call struct {
	*mock.Call
}

// DeleteBatchesNewerThanBatchNumber is a helper method to define mock.On call
//   - ctx context.Context
//   - batchNumber uint64
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) DeleteBatchesNewerThanBatchNumber(ctx interface{}, batchNumber interface{}, dbTx interface{}) *stateInterfaceMock_DeleteBatchesNewerThanBatchNumber_Call {
	return &stateInterfaceMock_DeleteBatchesNewerThanBatchNumber_Call{Call: _e.mock.On("DeleteBatchesNewerThanBatchNumber", ctx, batchNumber, dbTx)}
}

func (_c *stateInterfaceMock_DeleteBatchesNewerThanBatchNumber_Call) Run(run func(ctx context.Context, batchNumber uint64, dbTx pgx.Tx)) *stateInterfaceMock_DeleteBatchesNewerThanBatchNumber_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uint64), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_DeleteBatchesNewerThanBatchNumber_Call) Return(_a0 error) *stateInterfaceMock_DeleteBatchesNewerThanBatchNumber_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *stateInterfaceMock_DeleteBatchesNewerThanBatchNumber_Call) RunAndReturn(run func(context.Context, uint64, pgx.Tx) error) *stateInterfaceMock_DeleteBatchesNewerThanBatchNumber_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteBatchesOlderThanBatchNumber provides a mock function with given fields: ctx, batchNumber, dbTx
func (_m *stateInterfaceMock) DeleteBatchesOlderThanBatchNumber(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) error {
	ret := _m.Called(ctx, batchNumber, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for DeleteBatchesOlderThanBatchNumber")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uint64, pgx.Tx) error); ok {
		r0 = rf(ctx, batchNumber, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// stateInterfaceMock_DeleteBatchesOlderThanBatchNumber_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteBatchesOlderThanBatchNumber'
type stateInterfaceMock_DeleteBatchesOlderThanBatchNumber_Call struct {
	*mock.Call
}

// DeleteBatchesOlderThanBatchNumber is a helper method to define mock.On call
//   - ctx context.Context
//   - batchNumber uint64
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) DeleteBatchesOlderThanBatchNumber(ctx interface{}, batchNumber interface{}, dbTx interface{}) *stateInterfaceMock_DeleteBatchesOlderThanBatchNumber_Call {
	return &stateInterfaceMock_DeleteBatchesOlderThanBatchNumber_Call{Call: _e.mock.On("DeleteBatchesOlderThanBatchNumber", ctx, batchNumber, dbTx)}
}

func (_c *stateInterfaceMock_DeleteBatchesOlderThanBatchNumber_Call) Run(run func(ctx context.Context, batchNumber uint64, dbTx pgx.Tx)) *stateInterfaceMock_DeleteBatchesOlderThanBatchNumber_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uint64), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_DeleteBatchesOlderThanBatchNumber_Call) Return(_a0 error) *stateInterfaceMock_DeleteBatchesOlderThanBatchNumber_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *stateInterfaceMock_DeleteBatchesOlderThanBatchNumber_Call) RunAndReturn(run func(context.Context, uint64, pgx.Tx) error) *stateInterfaceMock_DeleteBatchesOlderThanBatchNumber_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteGeneratedProofs provides a mock function with given fields: ctx, batchNumber, batchNumberFinal, dbTx
func (_m *stateInterfaceMock) DeleteGeneratedProofs(ctx context.Context, batchNumber uint64, batchNumberFinal uint64, dbTx pgx.Tx) error {
	ret := _m.Called(ctx, batchNumber, batchNumberFinal, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for DeleteGeneratedProofs")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uint64, uint64, pgx.Tx) error); ok {
		r0 = rf(ctx, batchNumber, batchNumberFinal, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// stateInterfaceMock_DeleteGeneratedProofs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteGeneratedProofs'
type stateInterfaceMock_DeleteGeneratedProofs_Call struct {
	*mock.Call
}

// DeleteGeneratedProofs is a helper method to define mock.On call
//   - ctx context.Context
//   - batchNumber uint64
//   - batchNumberFinal uint64
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) DeleteGeneratedProofs(ctx interface{}, batchNumber interface{}, batchNumberFinal interface{}, dbTx interface{}) *stateInterfaceMock_DeleteGeneratedProofs_Call {
	return &stateInterfaceMock_DeleteGeneratedProofs_Call{Call: _e.mock.On("DeleteGeneratedProofs", ctx, batchNumber, batchNumberFinal, dbTx)}
}

func (_c *stateInterfaceMock_DeleteGeneratedProofs_Call) Run(run func(ctx context.Context, batchNumber uint64, batchNumberFinal uint64, dbTx pgx.Tx)) *stateInterfaceMock_DeleteGeneratedProofs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uint64), args[2].(uint64), args[3].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_DeleteGeneratedProofs_Call) Return(_a0 error) *stateInterfaceMock_DeleteGeneratedProofs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *stateInterfaceMock_DeleteGeneratedProofs_Call) RunAndReturn(run func(context.Context, uint64, uint64, pgx.Tx) error) *stateInterfaceMock_DeleteGeneratedProofs_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteUngeneratedProofs provides a mock function with given fields: ctx, dbTx
func (_m *stateInterfaceMock) DeleteUngeneratedProofs(ctx context.Context, dbTx pgx.Tx) error {
	ret := _m.Called(ctx, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for DeleteUngeneratedProofs")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, pgx.Tx) error); ok {
		r0 = rf(ctx, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// stateInterfaceMock_DeleteUngeneratedProofs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteUngeneratedProofs'
type stateInterfaceMock_DeleteUngeneratedProofs_Call struct {
	*mock.Call
}

// DeleteUngeneratedProofs is a helper method to define mock.On call
//   - ctx context.Context
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) DeleteUngeneratedProofs(ctx interface{}, dbTx interface{}) *stateInterfaceMock_DeleteUngeneratedProofs_Call {
	return &stateInterfaceMock_DeleteUngeneratedProofs_Call{Call: _e.mock.On("DeleteUngeneratedProofs", ctx, dbTx)}
}

func (_c *stateInterfaceMock_DeleteUngeneratedProofs_Call) Run(run func(ctx context.Context, dbTx pgx.Tx)) *stateInterfaceMock_DeleteUngeneratedProofs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_DeleteUngeneratedProofs_Call) Return(_a0 error) *stateInterfaceMock_DeleteUngeneratedProofs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *stateInterfaceMock_DeleteUngeneratedProofs_Call) RunAndReturn(run func(context.Context, pgx.Tx) error) *stateInterfaceMock_DeleteUngeneratedProofs_Call {
	_c.Call.Return(run)
	return _c
}

// GetBatch provides a mock function with given fields: ctx, batchNumber, dbTx
func (_m *stateInterfaceMock) GetBatch(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) (*state.DBBatch, error) {
	ret := _m.Called(ctx, batchNumber, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for GetBatch")
	}

	var r0 *state.DBBatch
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uint64, pgx.Tx) (*state.DBBatch, error)); ok {
		return rf(ctx, batchNumber, dbTx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uint64, pgx.Tx) *state.DBBatch); ok {
		r0 = rf(ctx, batchNumber, dbTx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*state.DBBatch)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uint64, pgx.Tx) error); ok {
		r1 = rf(ctx, batchNumber, dbTx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// stateInterfaceMock_GetBatch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetBatch'
type stateInterfaceMock_GetBatch_Call struct {
	*mock.Call
}

// GetBatch is a helper method to define mock.On call
//   - ctx context.Context
//   - batchNumber uint64
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) GetBatch(ctx interface{}, batchNumber interface{}, dbTx interface{}) *stateInterfaceMock_GetBatch_Call {
	return &stateInterfaceMock_GetBatch_Call{Call: _e.mock.On("GetBatch", ctx, batchNumber, dbTx)}
}

func (_c *stateInterfaceMock_GetBatch_Call) Run(run func(ctx context.Context, batchNumber uint64, dbTx pgx.Tx)) *stateInterfaceMock_GetBatch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uint64), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_GetBatch_Call) Return(_a0 *state.DBBatch, _a1 error) *stateInterfaceMock_GetBatch_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *stateInterfaceMock_GetBatch_Call) RunAndReturn(run func(context.Context, uint64, pgx.Tx) (*state.DBBatch, error)) *stateInterfaceMock_GetBatch_Call {
	_c.Call.Return(run)
	return _c
}

// GetProofReadyToVerify provides a mock function with given fields: ctx, lastVerfiedBatchNumber, dbTx
func (_m *stateInterfaceMock) GetProofReadyToVerify(ctx context.Context, lastVerfiedBatchNumber uint64, dbTx pgx.Tx) (*state.Proof, error) {
	ret := _m.Called(ctx, lastVerfiedBatchNumber, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for GetProofReadyToVerify")
	}

	var r0 *state.Proof
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uint64, pgx.Tx) (*state.Proof, error)); ok {
		return rf(ctx, lastVerfiedBatchNumber, dbTx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uint64, pgx.Tx) *state.Proof); ok {
		r0 = rf(ctx, lastVerfiedBatchNumber, dbTx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*state.Proof)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uint64, pgx.Tx) error); ok {
		r1 = rf(ctx, lastVerfiedBatchNumber, dbTx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// stateInterfaceMock_GetProofReadyToVerify_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProofReadyToVerify'
type stateInterfaceMock_GetProofReadyToVerify_Call struct {
	*mock.Call
}

// GetProofReadyToVerify is a helper method to define mock.On call
//   - ctx context.Context
//   - lastVerfiedBatchNumber uint64
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) GetProofReadyToVerify(ctx interface{}, lastVerfiedBatchNumber interface{}, dbTx interface{}) *stateInterfaceMock_GetProofReadyToVerify_Call {
	return &stateInterfaceMock_GetProofReadyToVerify_Call{Call: _e.mock.On("GetProofReadyToVerify", ctx, lastVerfiedBatchNumber, dbTx)}
}

func (_c *stateInterfaceMock_GetProofReadyToVerify_Call) Run(run func(ctx context.Context, lastVerfiedBatchNumber uint64, dbTx pgx.Tx)) *stateInterfaceMock_GetProofReadyToVerify_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uint64), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_GetProofReadyToVerify_Call) Return(_a0 *state.Proof, _a1 error) *stateInterfaceMock_GetProofReadyToVerify_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *stateInterfaceMock_GetProofReadyToVerify_Call) RunAndReturn(run func(context.Context, uint64, pgx.Tx) (*state.Proof, error)) *stateInterfaceMock_GetProofReadyToVerify_Call {
	_c.Call.Return(run)
	return _c
}

// GetProofsToAggregate provides a mock function with given fields: ctx, dbTx
func (_m *stateInterfaceMock) GetProofsToAggregate(ctx context.Context, dbTx pgx.Tx) (*state.Proof, *state.Proof, error) {
	ret := _m.Called(ctx, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for GetProofsToAggregate")
	}

	var r0 *state.Proof
	var r1 *state.Proof
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, pgx.Tx) (*state.Proof, *state.Proof, error)); ok {
		return rf(ctx, dbTx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, pgx.Tx) *state.Proof); ok {
		r0 = rf(ctx, dbTx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*state.Proof)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, pgx.Tx) *state.Proof); ok {
		r1 = rf(ctx, dbTx)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*state.Proof)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, pgx.Tx) error); ok {
		r2 = rf(ctx, dbTx)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// stateInterfaceMock_GetProofsToAggregate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProofsToAggregate'
type stateInterfaceMock_GetProofsToAggregate_Call struct {
	*mock.Call
}

// GetProofsToAggregate is a helper method to define mock.On call
//   - ctx context.Context
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) GetProofsToAggregate(ctx interface{}, dbTx interface{}) *stateInterfaceMock_GetProofsToAggregate_Call {
	return &stateInterfaceMock_GetProofsToAggregate_Call{Call: _e.mock.On("GetProofsToAggregate", ctx, dbTx)}
}

func (_c *stateInterfaceMock_GetProofsToAggregate_Call) Run(run func(ctx context.Context, dbTx pgx.Tx)) *stateInterfaceMock_GetProofsToAggregate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_GetProofsToAggregate_Call) Return(_a0 *state.Proof, _a1 *state.Proof, _a2 error) *stateInterfaceMock_GetProofsToAggregate_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *stateInterfaceMock_GetProofsToAggregate_Call) RunAndReturn(run func(context.Context, pgx.Tx) (*state.Proof, *state.Proof, error)) *stateInterfaceMock_GetProofsToAggregate_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateGeneratedProof provides a mock function with given fields: ctx, proof, dbTx
func (_m *stateInterfaceMock) UpdateGeneratedProof(ctx context.Context, proof *state.Proof, dbTx pgx.Tx) error {
	ret := _m.Called(ctx, proof, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for UpdateGeneratedProof")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *state.Proof, pgx.Tx) error); ok {
		r0 = rf(ctx, proof, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// stateInterfaceMock_UpdateGeneratedProof_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateGeneratedProof'
type stateInterfaceMock_UpdateGeneratedProof_Call struct {
	*mock.Call
}

// UpdateGeneratedProof is a helper method to define mock.On call
//   - ctx context.Context
//   - proof *state.Proof
//   - dbTx pgx.Tx
func (_e *stateInterfaceMock_Expecter) UpdateGeneratedProof(ctx interface{}, proof interface{}, dbTx interface{}) *stateInterfaceMock_UpdateGeneratedProof_Call {
	return &stateInterfaceMock_UpdateGeneratedProof_Call{Call: _e.mock.On("UpdateGeneratedProof", ctx, proof, dbTx)}
}

func (_c *stateInterfaceMock_UpdateGeneratedProof_Call) Run(run func(ctx context.Context, proof *state.Proof, dbTx pgx.Tx)) *stateInterfaceMock_UpdateGeneratedProof_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*state.Proof), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *stateInterfaceMock_UpdateGeneratedProof_Call) Return(_a0 error) *stateInterfaceMock_UpdateGeneratedProof_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *stateInterfaceMock_UpdateGeneratedProof_Call) RunAndReturn(run func(context.Context, *state.Proof, pgx.Tx) error) *stateInterfaceMock_UpdateGeneratedProof_Call {
	_c.Call.Return(run)
	return _c
}

// newStateInterfaceMock creates a new instance of stateInterfaceMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newStateInterfaceMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *stateInterfaceMock {
	mock := &stateInterfaceMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
