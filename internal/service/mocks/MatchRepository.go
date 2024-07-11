// Code generated by mockery v2.43.2. DO NOT EDIT.

package mockservice

import (
	context "context"

	muzz "github.com/nickbadlose/muzz"
	mock "github.com/stretchr/testify/mock"
)

// MatchRepository is an autogenerated mock type for the MatchRepository type
type MatchRepository struct {
	mock.Mock
}

type MatchRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MatchRepository) EXPECT() *MatchRepository_Expecter {
	return &MatchRepository_Expecter{mock: &_m.Mock}
}

// CreateSwipe provides a mock function with given fields: _a0, _a1
func (_m *MatchRepository) CreateSwipe(_a0 context.Context, _a1 *muzz.CreateSwipeInput) (*muzz.Match, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for CreateSwipe")
	}

	var r0 *muzz.Match
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *muzz.CreateSwipeInput) (*muzz.Match, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *muzz.CreateSwipeInput) *muzz.Match); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*muzz.Match)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *muzz.CreateSwipeInput) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MatchRepository_CreateSwipe_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateSwipe'
type MatchRepository_CreateSwipe_Call struct {
	*mock.Call
}

// CreateSwipe is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *muzz.CreateSwipeInput
func (_e *MatchRepository_Expecter) CreateSwipe(_a0 interface{}, _a1 interface{}) *MatchRepository_CreateSwipe_Call {
	return &MatchRepository_CreateSwipe_Call{Call: _e.mock.On("CreateSwipe", _a0, _a1)}
}

func (_c *MatchRepository_CreateSwipe_Call) Run(run func(_a0 context.Context, _a1 *muzz.CreateSwipeInput)) *MatchRepository_CreateSwipe_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*muzz.CreateSwipeInput))
	})
	return _c
}

func (_c *MatchRepository_CreateSwipe_Call) Return(_a0 *muzz.Match, _a1 error) *MatchRepository_CreateSwipe_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MatchRepository_CreateSwipe_Call) RunAndReturn(run func(context.Context, *muzz.CreateSwipeInput) (*muzz.Match, error)) *MatchRepository_CreateSwipe_Call {
	_c.Call.Return(run)
	return _c
}

// NewMatchRepository creates a new instance of MatchRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMatchRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MatchRepository {
	mock := &MatchRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}