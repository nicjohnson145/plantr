// Code generated by mockery. DO NOT EDIT.

package storage

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockClient is an autogenerated mock type for the Client type
type MockClient struct {
	mock.Mock
}

type MockClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockClient) EXPECT() *MockClient_Expecter {
	return &MockClient_Expecter{mock: &_m.Mock}
}

// ReadChallenge provides a mock function with given fields: ctx, id
func (_m *MockClient) ReadChallenge(ctx context.Context, id string) (*Challenge, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for ReadChallenge")
	}

	var r0 *Challenge
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*Challenge, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *Challenge); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Challenge)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockClient_ReadChallenge_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ReadChallenge'
type MockClient_ReadChallenge_Call struct {
	*mock.Call
}

// ReadChallenge is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *MockClient_Expecter) ReadChallenge(ctx interface{}, id interface{}) *MockClient_ReadChallenge_Call {
	return &MockClient_ReadChallenge_Call{Call: _e.mock.On("ReadChallenge", ctx, id)}
}

func (_c *MockClient_ReadChallenge_Call) Run(run func(ctx context.Context, id string)) *MockClient_ReadChallenge_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockClient_ReadChallenge_Call) Return(_a0 *Challenge, _a1 error) *MockClient_ReadChallenge_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockClient_ReadChallenge_Call) RunAndReturn(run func(context.Context, string) (*Challenge, error)) *MockClient_ReadChallenge_Call {
	_c.Call.Return(run)
	return _c
}

// WriteChallenge provides a mock function with given fields: ctx, challenge
func (_m *MockClient) WriteChallenge(ctx context.Context, challenge *Challenge) error {
	ret := _m.Called(ctx, challenge)

	if len(ret) == 0 {
		panic("no return value specified for WriteChallenge")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *Challenge) error); ok {
		r0 = rf(ctx, challenge)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockClient_WriteChallenge_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteChallenge'
type MockClient_WriteChallenge_Call struct {
	*mock.Call
}

// WriteChallenge is a helper method to define mock.On call
//   - ctx context.Context
//   - challenge *Challenge
func (_e *MockClient_Expecter) WriteChallenge(ctx interface{}, challenge interface{}) *MockClient_WriteChallenge_Call {
	return &MockClient_WriteChallenge_Call{Call: _e.mock.On("WriteChallenge", ctx, challenge)}
}

func (_c *MockClient_WriteChallenge_Call) Run(run func(ctx context.Context, challenge *Challenge)) *MockClient_WriteChallenge_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*Challenge))
	})
	return _c
}

func (_c *MockClient_WriteChallenge_Call) Return(_a0 error) *MockClient_WriteChallenge_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockClient_WriteChallenge_Call) RunAndReturn(run func(context.Context, *Challenge) error) *MockClient_WriteChallenge_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockClient creates a new instance of MockClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockClient {
	mock := &MockClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
