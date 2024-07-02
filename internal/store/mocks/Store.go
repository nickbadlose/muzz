// Code generated by mockery v2.43.2. DO NOT EDIT.

package mockstore

import (
	context "context"

	store "github.com/nickbadlose/muzz/internal/store"
	mock "github.com/stretchr/testify/mock"
)

// Store is an autogenerated mock type for the Store type
type Store struct {
	mock.Mock
}

type Store_Expecter struct {
	mock *mock.Mock
}

func (_m *Store) EXPECT() *Store_Expecter {
	return &Store_Expecter{mock: &_m.Mock}
}

// CreateUser provides a mock function with given fields: _a0, _a1
func (_m *Store) CreateUser(_a0 context.Context, _a1 *store.CreateUserInput) (*store.User, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for CreateUser")
	}

	var r0 *store.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *store.CreateUserInput) (*store.User, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *store.CreateUserInput) *store.User); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*store.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *store.CreateUserInput) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store_CreateUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateUser'
type Store_CreateUser_Call struct {
	*mock.Call
}

// CreateUser is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *store.CreateUserInput
func (_e *Store_Expecter) CreateUser(_a0 interface{}, _a1 interface{}) *Store_CreateUser_Call {
	return &Store_CreateUser_Call{Call: _e.mock.On("CreateUser", _a0, _a1)}
}

func (_c *Store_CreateUser_Call) Run(run func(_a0 context.Context, _a1 *store.CreateUserInput)) *Store_CreateUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*store.CreateUserInput))
	})
	return _c
}

func (_c *Store_CreateUser_Call) Return(_a0 *store.User, _a1 error) *Store_CreateUser_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Store_CreateUser_Call) RunAndReturn(run func(context.Context, *store.CreateUserInput) (*store.User, error)) *Store_CreateUser_Call {
	_c.Call.Return(run)
	return _c
}

// GetUserByEmail provides a mock function with given fields: _a0, _a1
func (_m *Store) GetUserByEmail(_a0 context.Context, _a1 string) (*store.User, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetUserByEmail")
	}

	var r0 *store.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*store.User, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *store.User); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*store.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store_GetUserByEmail_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUserByEmail'
type Store_GetUserByEmail_Call struct {
	*mock.Call
}

// GetUserByEmail is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
func (_e *Store_Expecter) GetUserByEmail(_a0 interface{}, _a1 interface{}) *Store_GetUserByEmail_Call {
	return &Store_GetUserByEmail_Call{Call: _e.mock.On("GetUserByEmail", _a0, _a1)}
}

func (_c *Store_GetUserByEmail_Call) Run(run func(_a0 context.Context, _a1 string)) *Store_GetUserByEmail_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Store_GetUserByEmail_Call) Return(_a0 *store.User, _a1 error) *Store_GetUserByEmail_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Store_GetUserByEmail_Call) RunAndReturn(run func(context.Context, string) (*store.User, error)) *Store_GetUserByEmail_Call {
	_c.Call.Return(run)
	return _c
}

// GetUsers provides a mock function with given fields: _a0, _a1
func (_m *Store) GetUsers(_a0 context.Context, _a1 int) ([]*store.UserDetails, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetUsers")
	}

	var r0 []*store.UserDetails
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int) ([]*store.UserDetails, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int) []*store.UserDetails); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*store.UserDetails)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store_GetUsers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUsers'
type Store_GetUsers_Call struct {
	*mock.Call
}

// GetUsers is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 int
func (_e *Store_Expecter) GetUsers(_a0 interface{}, _a1 interface{}) *Store_GetUsers_Call {
	return &Store_GetUsers_Call{Call: _e.mock.On("GetUsers", _a0, _a1)}
}

func (_c *Store_GetUsers_Call) Run(run func(_a0 context.Context, _a1 int)) *Store_GetUsers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int))
	})
	return _c
}

func (_c *Store_GetUsers_Call) Return(_a0 []*store.UserDetails, _a1 error) *Store_GetUsers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Store_GetUsers_Call) RunAndReturn(run func(context.Context, int) ([]*store.UserDetails, error)) *Store_GetUsers_Call {
	_c.Call.Return(run)
	return _c
}

// NewStore creates a new instance of Store. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *Store {
	mock := &Store{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
