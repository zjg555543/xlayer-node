// Code generated by mockery. DO NOT EDIT.

package mock_l2_shared

import (
	context "context"

	l2_shared "github.com/0xPolygonHermez/zkevm-node/synchronizer/l2_sync/l2_shared"
	mock "github.com/stretchr/testify/mock"

	pgx "github.com/jackc/pgx/v4"
)

// PostClosedBatchChecker is an autogenerated mock type for the PostClosedBatchChecker type
type PostClosedBatchChecker struct {
	mock.Mock
}

type PostClosedBatchChecker_Expecter struct {
	mock *mock.Mock
}

func (_m *PostClosedBatchChecker) EXPECT() *PostClosedBatchChecker_Expecter {
	return &PostClosedBatchChecker_Expecter{mock: &_m.Mock}
}

// CheckPostClosedBatch provides a mock function with given fields: ctx, processData, dbTx
func (_m *PostClosedBatchChecker) CheckPostClosedBatch(ctx context.Context, processData l2_shared.ProcessData, dbTx pgx.Tx) error {
	ret := _m.Called(ctx, processData, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for CheckPostClosedBatch")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, l2_shared.ProcessData, pgx.Tx) error); ok {
		r0 = rf(ctx, processData, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PostClosedBatchChecker_CheckPostClosedBatch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CheckPostClosedBatch'
type PostClosedBatchChecker_CheckPostClosedBatch_Call struct {
	*mock.Call
}

// CheckPostClosedBatch is a helper method to define mock.On call
//   - ctx context.Context
//   - processData l2_shared.ProcessData
//   - dbTx pgx.Tx
func (_e *PostClosedBatchChecker_Expecter) CheckPostClosedBatch(ctx interface{}, processData interface{}, dbTx interface{}) *PostClosedBatchChecker_CheckPostClosedBatch_Call {
	return &PostClosedBatchChecker_CheckPostClosedBatch_Call{Call: _e.mock.On("CheckPostClosedBatch", ctx, processData, dbTx)}
}

func (_c *PostClosedBatchChecker_CheckPostClosedBatch_Call) Run(run func(ctx context.Context, processData l2_shared.ProcessData, dbTx pgx.Tx)) *PostClosedBatchChecker_CheckPostClosedBatch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(l2_shared.ProcessData), args[2].(pgx.Tx))
	})
	return _c
}

func (_c *PostClosedBatchChecker_CheckPostClosedBatch_Call) Return(_a0 error) *PostClosedBatchChecker_CheckPostClosedBatch_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *PostClosedBatchChecker_CheckPostClosedBatch_Call) RunAndReturn(run func(context.Context, l2_shared.ProcessData, pgx.Tx) error) *PostClosedBatchChecker_CheckPostClosedBatch_Call {
	_c.Call.Return(run)
	return _c
}

// NewPostClosedBatchChecker creates a new instance of PostClosedBatchChecker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPostClosedBatchChecker(t interface {
	mock.TestingT
	Cleanup(func())
}) *PostClosedBatchChecker {
	mock := &PostClosedBatchChecker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
