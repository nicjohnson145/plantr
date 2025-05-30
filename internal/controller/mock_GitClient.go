// Code generated by mockery. DO NOT EDIT.

package controller

import (
	fs "io/fs"

	mock "github.com/stretchr/testify/mock"
)

// MockGitClient is an autogenerated mock type for the GitClient type
type MockGitClient struct {
	mock.Mock
}

type MockGitClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockGitClient) EXPECT() *MockGitClient_Expecter {
	return &MockGitClient_Expecter{mock: &_m.Mock}
}

// CloneAtCommit provides a mock function with given fields: url, commit
func (_m *MockGitClient) CloneAtCommit(url string, commit string) (fs.FS, error) {
	ret := _m.Called(url, commit)

	if len(ret) == 0 {
		panic("no return value specified for CloneAtCommit")
	}

	var r0 fs.FS
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (fs.FS, error)); ok {
		return rf(url, commit)
	}
	if rf, ok := ret.Get(0).(func(string, string) fs.FS); ok {
		r0 = rf(url, commit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(fs.FS)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(url, commit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockGitClient_CloneAtCommit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CloneAtCommit'
type MockGitClient_CloneAtCommit_Call struct {
	*mock.Call
}

// CloneAtCommit is a helper method to define mock.On call
//   - url string
//   - commit string
func (_e *MockGitClient_Expecter) CloneAtCommit(url interface{}, commit interface{}) *MockGitClient_CloneAtCommit_Call {
	return &MockGitClient_CloneAtCommit_Call{Call: _e.mock.On("CloneAtCommit", url, commit)}
}

func (_c *MockGitClient_CloneAtCommit_Call) Run(run func(url string, commit string)) *MockGitClient_CloneAtCommit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *MockGitClient_CloneAtCommit_Call) Return(_a0 fs.FS, _a1 error) *MockGitClient_CloneAtCommit_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockGitClient_CloneAtCommit_Call) RunAndReturn(run func(string, string) (fs.FS, error)) *MockGitClient_CloneAtCommit_Call {
	_c.Call.Return(run)
	return _c
}

// GetLatestCommit provides a mock function with given fields: url
func (_m *MockGitClient) GetLatestCommit(url string) (string, error) {
	ret := _m.Called(url)

	if len(ret) == 0 {
		panic("no return value specified for GetLatestCommit")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(url)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(url)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(url)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockGitClient_GetLatestCommit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLatestCommit'
type MockGitClient_GetLatestCommit_Call struct {
	*mock.Call
}

// GetLatestCommit is a helper method to define mock.On call
//   - url string
func (_e *MockGitClient_Expecter) GetLatestCommit(url interface{}) *MockGitClient_GetLatestCommit_Call {
	return &MockGitClient_GetLatestCommit_Call{Call: _e.mock.On("GetLatestCommit", url)}
}

func (_c *MockGitClient_GetLatestCommit_Call) Run(run func(url string)) *MockGitClient_GetLatestCommit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockGitClient_GetLatestCommit_Call) Return(_a0 string, _a1 error) *MockGitClient_GetLatestCommit_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockGitClient_GetLatestCommit_Call) RunAndReturn(run func(string) (string, error)) *MockGitClient_GetLatestCommit_Call {
	_c.Call.Return(run)
	return _c
}

// GetLatestRelease provides a mock function with given fields: url
func (_m *MockGitClient) GetLatestRelease(url string) (string, error) {
	ret := _m.Called(url)

	if len(ret) == 0 {
		panic("no return value specified for GetLatestRelease")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(url)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(url)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(url)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockGitClient_GetLatestRelease_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLatestRelease'
type MockGitClient_GetLatestRelease_Call struct {
	*mock.Call
}

// GetLatestRelease is a helper method to define mock.On call
//   - url string
func (_e *MockGitClient_Expecter) GetLatestRelease(url interface{}) *MockGitClient_GetLatestRelease_Call {
	return &MockGitClient_GetLatestRelease_Call{Call: _e.mock.On("GetLatestRelease", url)}
}

func (_c *MockGitClient_GetLatestRelease_Call) Run(run func(url string)) *MockGitClient_GetLatestRelease_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockGitClient_GetLatestRelease_Call) Return(_a0 string, _a1 error) *MockGitClient_GetLatestRelease_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockGitClient_GetLatestRelease_Call) RunAndReturn(run func(string) (string, error)) *MockGitClient_GetLatestRelease_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockGitClient creates a new instance of MockGitClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockGitClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGitClient {
	mock := &MockGitClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
