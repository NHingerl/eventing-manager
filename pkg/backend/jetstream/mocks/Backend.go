// Code generated by mockery v2.51.1. DO NOT EDIT.

package mocks

import (
	env "github.com/kyma-project/eventing-manager/pkg/env"

	mock "github.com/stretchr/testify/mock"

	nats "github.com/nats-io/nats.go"

	utils "github.com/kyma-project/eventing-manager/pkg/backend/utils"

	v1alpha2 "github.com/kyma-project/eventing-manager/api/eventing/v1alpha2"
)

// Backend is an autogenerated mock type for the Backend type
type Backend struct {
	mock.Mock
}

type Backend_Expecter struct {
	mock *mock.Mock
}

func (_m *Backend) EXPECT() *Backend_Expecter {
	return &Backend_Expecter{mock: &_m.Mock}
}

// DeleteInvalidConsumers provides a mock function with given fields: subscriptions
func (_m *Backend) DeleteInvalidConsumers(subscriptions []v1alpha2.Subscription) error {
	ret := _m.Called(subscriptions)

	if len(ret) == 0 {
		panic("no return value specified for DeleteInvalidConsumers")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]v1alpha2.Subscription) error); ok {
		r0 = rf(subscriptions)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Backend_DeleteInvalidConsumers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteInvalidConsumers'
type Backend_DeleteInvalidConsumers_Call struct {
	*mock.Call
}

// DeleteInvalidConsumers is a helper method to define mock.On call
//   - subscriptions []v1alpha2.Subscription
func (_e *Backend_Expecter) DeleteInvalidConsumers(subscriptions interface{}) *Backend_DeleteInvalidConsumers_Call {
	return &Backend_DeleteInvalidConsumers_Call{Call: _e.mock.On("DeleteInvalidConsumers", subscriptions)}
}

func (_c *Backend_DeleteInvalidConsumers_Call) Run(run func(subscriptions []v1alpha2.Subscription)) *Backend_DeleteInvalidConsumers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]v1alpha2.Subscription))
	})
	return _c
}

func (_c *Backend_DeleteInvalidConsumers_Call) Return(_a0 error) *Backend_DeleteInvalidConsumers_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Backend_DeleteInvalidConsumers_Call) RunAndReturn(run func([]v1alpha2.Subscription) error) *Backend_DeleteInvalidConsumers_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteSubscription provides a mock function with given fields: subscription
func (_m *Backend) DeleteSubscription(subscription *v1alpha2.Subscription) error {
	ret := _m.Called(subscription)

	if len(ret) == 0 {
		panic("no return value specified for DeleteSubscription")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1alpha2.Subscription) error); ok {
		r0 = rf(subscription)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Backend_DeleteSubscription_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteSubscription'
type Backend_DeleteSubscription_Call struct {
	*mock.Call
}

// DeleteSubscription is a helper method to define mock.On call
//   - subscription *v1alpha2.Subscription
func (_e *Backend_Expecter) DeleteSubscription(subscription interface{}) *Backend_DeleteSubscription_Call {
	return &Backend_DeleteSubscription_Call{Call: _e.mock.On("DeleteSubscription", subscription)}
}

func (_c *Backend_DeleteSubscription_Call) Run(run func(subscription *v1alpha2.Subscription)) *Backend_DeleteSubscription_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*v1alpha2.Subscription))
	})
	return _c
}

func (_c *Backend_DeleteSubscription_Call) Return(_a0 error) *Backend_DeleteSubscription_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Backend_DeleteSubscription_Call) RunAndReturn(run func(*v1alpha2.Subscription) error) *Backend_DeleteSubscription_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteSubscriptionsOnly provides a mock function with given fields: subscription
func (_m *Backend) DeleteSubscriptionsOnly(subscription *v1alpha2.Subscription) error {
	ret := _m.Called(subscription)

	if len(ret) == 0 {
		panic("no return value specified for DeleteSubscriptionsOnly")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1alpha2.Subscription) error); ok {
		r0 = rf(subscription)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Backend_DeleteSubscriptionsOnly_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteSubscriptionsOnly'
type Backend_DeleteSubscriptionsOnly_Call struct {
	*mock.Call
}

// DeleteSubscriptionsOnly is a helper method to define mock.On call
//   - subscription *v1alpha2.Subscription
func (_e *Backend_Expecter) DeleteSubscriptionsOnly(subscription interface{}) *Backend_DeleteSubscriptionsOnly_Call {
	return &Backend_DeleteSubscriptionsOnly_Call{Call: _e.mock.On("DeleteSubscriptionsOnly", subscription)}
}

func (_c *Backend_DeleteSubscriptionsOnly_Call) Run(run func(subscription *v1alpha2.Subscription)) *Backend_DeleteSubscriptionsOnly_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*v1alpha2.Subscription))
	})
	return _c
}

func (_c *Backend_DeleteSubscriptionsOnly_Call) Return(_a0 error) *Backend_DeleteSubscriptionsOnly_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Backend_DeleteSubscriptionsOnly_Call) RunAndReturn(run func(*v1alpha2.Subscription) error) *Backend_DeleteSubscriptionsOnly_Call {
	_c.Call.Return(run)
	return _c
}

// GetConfig provides a mock function with no fields
func (_m *Backend) GetConfig() env.NATSConfig {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetConfig")
	}

	var r0 env.NATSConfig
	if rf, ok := ret.Get(0).(func() env.NATSConfig); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(env.NATSConfig)
	}

	return r0
}

// Backend_GetConfig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetConfig'
type Backend_GetConfig_Call struct {
	*mock.Call
}

// GetConfig is a helper method to define mock.On call
func (_e *Backend_Expecter) GetConfig() *Backend_GetConfig_Call {
	return &Backend_GetConfig_Call{Call: _e.mock.On("GetConfig")}
}

func (_c *Backend_GetConfig_Call) Run(run func()) *Backend_GetConfig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Backend_GetConfig_Call) Return(_a0 env.NATSConfig) *Backend_GetConfig_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Backend_GetConfig_Call) RunAndReturn(run func() env.NATSConfig) *Backend_GetConfig_Call {
	_c.Call.Return(run)
	return _c
}

// GetJetStreamContext provides a mock function with no fields
func (_m *Backend) GetJetStreamContext() nats.JetStreamContext {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetJetStreamContext")
	}

	var r0 nats.JetStreamContext
	if rf, ok := ret.Get(0).(func() nats.JetStreamContext); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(nats.JetStreamContext)
		}
	}

	return r0
}

// Backend_GetJetStreamContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetJetStreamContext'
type Backend_GetJetStreamContext_Call struct {
	*mock.Call
}

// GetJetStreamContext is a helper method to define mock.On call
func (_e *Backend_Expecter) GetJetStreamContext() *Backend_GetJetStreamContext_Call {
	return &Backend_GetJetStreamContext_Call{Call: _e.mock.On("GetJetStreamContext")}
}

func (_c *Backend_GetJetStreamContext_Call) Run(run func()) *Backend_GetJetStreamContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Backend_GetJetStreamContext_Call) Return(_a0 nats.JetStreamContext) *Backend_GetJetStreamContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Backend_GetJetStreamContext_Call) RunAndReturn(run func() nats.JetStreamContext) *Backend_GetJetStreamContext_Call {
	_c.Call.Return(run)
	return _c
}

// GetJetStreamSubjects provides a mock function with given fields: source, subjects, typeMatching
func (_m *Backend) GetJetStreamSubjects(source string, subjects []string, typeMatching v1alpha2.TypeMatching) []string {
	ret := _m.Called(source, subjects, typeMatching)

	if len(ret) == 0 {
		panic("no return value specified for GetJetStreamSubjects")
	}

	var r0 []string
	if rf, ok := ret.Get(0).(func(string, []string, v1alpha2.TypeMatching) []string); ok {
		r0 = rf(source, subjects, typeMatching)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Backend_GetJetStreamSubjects_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetJetStreamSubjects'
type Backend_GetJetStreamSubjects_Call struct {
	*mock.Call
}

// GetJetStreamSubjects is a helper method to define mock.On call
//   - source string
//   - subjects []string
//   - typeMatching v1alpha2.TypeMatching
func (_e *Backend_Expecter) GetJetStreamSubjects(source interface{}, subjects interface{}, typeMatching interface{}) *Backend_GetJetStreamSubjects_Call {
	return &Backend_GetJetStreamSubjects_Call{Call: _e.mock.On("GetJetStreamSubjects", source, subjects, typeMatching)}
}

func (_c *Backend_GetJetStreamSubjects_Call) Run(run func(source string, subjects []string, typeMatching v1alpha2.TypeMatching)) *Backend_GetJetStreamSubjects_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].([]string), args[2].(v1alpha2.TypeMatching))
	})
	return _c
}

func (_c *Backend_GetJetStreamSubjects_Call) Return(_a0 []string) *Backend_GetJetStreamSubjects_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Backend_GetJetStreamSubjects_Call) RunAndReturn(run func(string, []string, v1alpha2.TypeMatching) []string) *Backend_GetJetStreamSubjects_Call {
	_c.Call.Return(run)
	return _c
}

// Initialize provides a mock function with given fields: connCloseHandler
func (_m *Backend) Initialize(connCloseHandler utils.ConnClosedHandler) error {
	ret := _m.Called(connCloseHandler)

	if len(ret) == 0 {
		panic("no return value specified for Initialize")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(utils.ConnClosedHandler) error); ok {
		r0 = rf(connCloseHandler)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Backend_Initialize_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Initialize'
type Backend_Initialize_Call struct {
	*mock.Call
}

// Initialize is a helper method to define mock.On call
//   - connCloseHandler utils.ConnClosedHandler
func (_e *Backend_Expecter) Initialize(connCloseHandler interface{}) *Backend_Initialize_Call {
	return &Backend_Initialize_Call{Call: _e.mock.On("Initialize", connCloseHandler)}
}

func (_c *Backend_Initialize_Call) Run(run func(connCloseHandler utils.ConnClosedHandler)) *Backend_Initialize_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(utils.ConnClosedHandler))
	})
	return _c
}

func (_c *Backend_Initialize_Call) Return(_a0 error) *Backend_Initialize_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Backend_Initialize_Call) RunAndReturn(run func(utils.ConnClosedHandler) error) *Backend_Initialize_Call {
	_c.Call.Return(run)
	return _c
}

// Shutdown provides a mock function with no fields
func (_m *Backend) Shutdown() {
	_m.Called()
}

// Backend_Shutdown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Shutdown'
type Backend_Shutdown_Call struct {
	*mock.Call
}

// Shutdown is a helper method to define mock.On call
func (_e *Backend_Expecter) Shutdown() *Backend_Shutdown_Call {
	return &Backend_Shutdown_Call{Call: _e.mock.On("Shutdown")}
}

func (_c *Backend_Shutdown_Call) Run(run func()) *Backend_Shutdown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Backend_Shutdown_Call) Return() *Backend_Shutdown_Call {
	_c.Call.Return()
	return _c
}

func (_c *Backend_Shutdown_Call) RunAndReturn(run func()) *Backend_Shutdown_Call {
	_c.Run(run)
	return _c
}

// SyncSubscription provides a mock function with given fields: subscription
func (_m *Backend) SyncSubscription(subscription *v1alpha2.Subscription) error {
	ret := _m.Called(subscription)

	if len(ret) == 0 {
		panic("no return value specified for SyncSubscription")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1alpha2.Subscription) error); ok {
		r0 = rf(subscription)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Backend_SyncSubscription_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SyncSubscription'
type Backend_SyncSubscription_Call struct {
	*mock.Call
}

// SyncSubscription is a helper method to define mock.On call
//   - subscription *v1alpha2.Subscription
func (_e *Backend_Expecter) SyncSubscription(subscription interface{}) *Backend_SyncSubscription_Call {
	return &Backend_SyncSubscription_Call{Call: _e.mock.On("SyncSubscription", subscription)}
}

func (_c *Backend_SyncSubscription_Call) Run(run func(subscription *v1alpha2.Subscription)) *Backend_SyncSubscription_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*v1alpha2.Subscription))
	})
	return _c
}

func (_c *Backend_SyncSubscription_Call) Return(_a0 error) *Backend_SyncSubscription_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Backend_SyncSubscription_Call) RunAndReturn(run func(*v1alpha2.Subscription) error) *Backend_SyncSubscription_Call {
	_c.Call.Return(run)
	return _c
}

// NewBackend creates a new instance of Backend. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBackend(t interface {
	mock.TestingT
	Cleanup(func())
}) *Backend {
	mock := &Backend{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
