// Copyright 2022 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	context "context"

	model "github.com/mendersoftware/reporting/model"
	mock "github.com/stretchr/testify/mock"
)

// Store is an autogenerated mock type for the Store type
type Store struct {
	mock.Mock
}

// AggregateDeployments provides a mock function with given fields: ctx, query
func (_m *Store) AggregateDeployments(ctx context.Context, query model.Query) (model.M, error) {
	ret := _m.Called(ctx, query)

	var r0 model.M
	if rf, ok := ret.Get(0).(func(context.Context, model.Query) model.M); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.M)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.Query) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AggregateDevices provides a mock function with given fields: ctx, query
func (_m *Store) AggregateDevices(ctx context.Context, query model.Query) (model.M, error) {
	ret := _m.Called(ctx, query)

	var r0 model.M
	if rf, ok := ret.Get(0).(func(context.Context, model.Query) model.M); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.M)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.Query) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BulkIndexDeployments provides a mock function with given fields: ctx, deployments
func (_m *Store) BulkIndexDeployments(ctx context.Context, deployments []*model.Deployment) error {
	ret := _m.Called(ctx, deployments)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*model.Deployment) error); ok {
		r0 = rf(ctx, deployments)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// BulkIndexDevices provides a mock function with given fields: ctx, devices, removedDevices
func (_m *Store) BulkIndexDevices(ctx context.Context, devices []*model.Device, removedDevices []*model.Device) error {
	ret := _m.Called(ctx, devices, removedDevices)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*model.Device, []*model.Device) error); ok {
		r0 = rf(ctx, devices, removedDevices)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetDeploymentsIndex provides a mock function with given fields: tid
func (_m *Store) GetDeploymentsIndex(tid string) string {
	ret := _m.Called(tid)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(tid)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetDeploymentsIndexMapping provides a mock function with given fields: ctx, tid
func (_m *Store) GetDeploymentsIndexMapping(ctx context.Context, tid string) (map[string]interface{}, error) {
	ret := _m.Called(ctx, tid)

	var r0 map[string]interface{}
	if rf, ok := ret.Get(0).(func(context.Context, string) map[string]interface{}); ok {
		r0 = rf(ctx, tid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDeploymentsRoutingKey provides a mock function with given fields: tid
func (_m *Store) GetDeploymentsRoutingKey(tid string) string {
	ret := _m.Called(tid)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(tid)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetDevicesIndex provides a mock function with given fields: tid
func (_m *Store) GetDevicesIndex(tid string) string {
	ret := _m.Called(tid)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(tid)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetDevicesIndexMapping provides a mock function with given fields: ctx, tid
func (_m *Store) GetDevicesIndexMapping(ctx context.Context, tid string) (map[string]interface{}, error) {
	ret := _m.Called(ctx, tid)

	var r0 map[string]interface{}
	if rf, ok := ret.Get(0).(func(context.Context, string) map[string]interface{}); ok {
		r0 = rf(ctx, tid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDevicesRoutingKey provides a mock function with given fields: tid
func (_m *Store) GetDevicesRoutingKey(tid string) string {
	ret := _m.Called(tid)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(tid)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Migrate provides a mock function with given fields: ctx
func (_m *Store) Migrate(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Ping provides a mock function with given fields: ctx
func (_m *Store) Ping(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SearchDeployments provides a mock function with given fields: ctx, query
func (_m *Store) SearchDeployments(ctx context.Context, query model.Query) (model.M, error) {
	ret := _m.Called(ctx, query)

	var r0 model.M
	if rf, ok := ret.Get(0).(func(context.Context, model.Query) model.M); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.M)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.Query) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SearchDevices provides a mock function with given fields: ctx, query
func (_m *Store) SearchDevices(ctx context.Context, query model.Query) (model.M, error) {
	ret := _m.Called(ctx, query)

	var r0 model.M
	if rf, ok := ret.Get(0).(func(context.Context, model.Query) model.M); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.M)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.Query) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
