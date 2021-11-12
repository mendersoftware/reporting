// Copyright 2021 Northern.tech AS
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

package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"testing"
	"time"

	natsio "github.com/nats-io/nats.go"

	"github.com/mendersoftware/reporting/client/deviceauth"
	deviceauth_mocks "github.com/mendersoftware/reporting/client/deviceauth/mocks"
	"github.com/mendersoftware/reporting/client/inventory"
	inventory_mocks "github.com/mendersoftware/reporting/client/inventory/mocks"
	"github.com/mendersoftware/reporting/client/nats"
	nats_mocks "github.com/mendersoftware/reporting/client/nats/mocks"
	"github.com/mendersoftware/reporting/model"
	store_mocks "github.com/mendersoftware/reporting/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetJobsSubscriptionError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan *model.Job, 1)

	var unsubscribe nats.UnsubscribeFunc = func() error {
		return nil
	}

	subscriptionError := errors.New("subscription error")

	nats := &nats_mocks.Client{}
	nats.On("JetStreamCreateStream",
		mock.AnythingOfType("string"),
	).Return(nil)

	nats.On("JetStreamSubscribe",
		ctx,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("chan *nats.Msg"),
	).Return(unsubscribe, subscriptionError)

	defer nats.AssertExpectations(t)

	indexer := NewIndexer(nil, nats, nil, nil)
	err := indexer.GetJobs(ctx, jobs)
	assert.Equal(t, "failed to subscribe to the nats JetStream: subscription error", err.Error())

	cancel()
}

func TestGetJobs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan *model.Job, 1)

	var unsubscribe nats.UnsubscribeFunc = func() error {
		return nil
	}

	nats := &nats_mocks.Client{}
	nats.On("JetStreamCreateStream",
		mock.AnythingOfType("string"),
	).Return(nil)

	nats.On("JetStreamSubscribe",
		ctx,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(msgs chan *natsio.Msg) bool {
			job := &model.Job{Action: "index"}
			jobData, _ := json.Marshal(job)
			msgs <- &natsio.Msg{
				Data: jobData,
			}

			return true
		}),
	).Return(unsubscribe, nil)

	defer nats.AssertExpectations(t)

	indexer := NewIndexer(nil, nats, nil, nil)
	err := indexer.GetJobs(ctx, jobs)
	assert.Nil(t, err)

	time.Sleep(500 * time.Millisecond)

	job := <-jobs
	assert.Equal(t, job.Action, "index")

	cancel()
}

func TestGetJobsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan *model.Job, 1)

	var unsubscribe nats.UnsubscribeFunc = func() error {
		return nil
	}

	nats := &nats_mocks.Client{}
	nats.On("JetStreamCreateStream",
		mock.AnythingOfType("string"),
	).Return(nil)

	nats.On("JetStreamSubscribe",
		ctx,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(msgs chan *natsio.Msg) bool {
			msgs <- &natsio.Msg{
				Data: []byte(""),
			}

			return true
		}),
	).Return(unsubscribe, nil)

	defer nats.AssertExpectations(t)

	indexer := NewIndexer(nil, nats, nil, nil)
	err := indexer.GetJobs(ctx, jobs)
	assert.Nil(t, err)

	time.Sleep(500 * time.Millisecond)

	select {
	case <-jobs:
		assert.Fail(t, "unexpected message")
	default:
	}

	cancel()
}

func strptr(s string) *string {
	return &s
}

func TestProcessJobs(t *testing.T) {
	const tenantID = "tenant"

	testCases := map[string]struct {
		jobs []*model.Job

		deviceauthDeviceIDs []string
		deviceauthDevices   []deviceauth.DeviceAuthDevice
		deviceauthErr       error

		inventoryDeviceIDs []string
		inventoryDevices   []inventory.Device
		inventoryErr       error

		bulkIndexDevices       []*model.Device
		bulkIndexRemoveDevices []*model.Device
		bulkIndexErr           error
	}{
		"ok": {
			jobs: []*model.Job{
				{
					Action:   "index",
					TenantID: tenantID,
					DeviceID: "1",
					Service:  model.ServiceInventory,
				},
				{
					Action:   "index",
					TenantID: tenantID,
					DeviceID: "2",
					Service:  model.ServiceInventory,
				},
				{
					Action:   "index",
					TenantID: tenantID,
					DeviceID: "3",
					Service:  model.ServiceInventory,
				},
			},

			deviceauthDeviceIDs: []string{"1", "2", "3"},
			deviceauthDevices: []deviceauth.DeviceAuthDevice{
				{
					ID:     "1",
					Status: "active",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:44",
					},
				},
				{
					ID:     "2",
					Status: "pending",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:55",
					},
				},
			},

			inventoryDeviceIDs: []string{"1", "2"},
			inventoryDevices: []inventory.Device{
				{
					ID: "1",
				},
				{
					ID: "2",
				},
			},

			bulkIndexDevices: []*model.Device{
				{
					ID:       strptr("1"),
					TenantID: strptr(tenantID),
					IdentityAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeIdentity,
							Name:   model.AttrNameStatus,
							String: []string{"active"},
						},
						{
							Scope:  model.ScopeIdentity,
							Name:   "mac",
							String: []string{"00:11:22:33:44"},
						},
					},
				},
				{
					ID:       strptr("2"),
					TenantID: strptr(tenantID),
					IdentityAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeIdentity,
							Name:   model.AttrNameStatus,
							String: []string{"pending"},
						},
						{
							Scope:  model.ScopeIdentity,
							Name:   "mac",
							String: []string{"00:11:22:33:55"},
						},
					},
				},
			},
			bulkIndexRemoveDevices: []*model.Device{
				{
					ID:       strptr("3"),
					TenantID: strptr(tenantID),
				},
			},
		},
		"ko, failure in deviceauth": {
			jobs: []*model.Job{
				{
					Action:   "index",
					TenantID: tenantID,
					DeviceID: "1",
					Service:  model.ServiceInventory,
				},
				{
					Action:   "index",
					TenantID: tenantID,
					DeviceID: "2",
					Service:  model.ServiceInventory,
				},
				{
					Action:   "index",
					TenantID: tenantID,
					DeviceID: "3",
					Service:  model.ServiceInventory,
				},
			},

			deviceauthDeviceIDs: []string{"1", "2", "3"},
			deviceauthErr:       errors.New("abc"),
		},
		"ko, failure in inventory": {
			jobs: []*model.Job{
				{
					Action:   "index",
					TenantID: tenantID,
					DeviceID: "1",
					Service:  model.ServiceInventory,
				},
				{
					Action:   "index",
					TenantID: tenantID,
					DeviceID: "2",
					Service:  model.ServiceInventory,
				},
				{
					Action:   "index",
					TenantID: tenantID,
					DeviceID: "3",
					Service:  model.ServiceInventory,
				},
			},

			deviceauthDeviceIDs: []string{"1", "2", "3"},
			deviceauthDevices: []deviceauth.DeviceAuthDevice{
				{
					ID:     "1",
					Status: "active",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:44",
					},
				},
				{
					ID:     "2",
					Status: "pending",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:55",
					},
				},
			},

			inventoryDeviceIDs: []string{"1", "2", "3"},
			inventoryErr:       errors.New("abc"),
		},
		"ko, failure in BulkIndex": {
			jobs: []*model.Job{
				{
					Action:   "index",
					TenantID: tenantID,
					DeviceID: "1",
					Service:  model.ServiceInventory,
				},
				{
					Action:   "index",
					TenantID: tenantID,
					DeviceID: "2",
					Service:  model.ServiceInventory,
				},
				{
					Action:   "index",
					TenantID: tenantID,
					DeviceID: "3",
					Service:  model.ServiceInventory,
				},
			},

			deviceauthDeviceIDs: []string{"1", "2", "3"},
			deviceauthDevices: []deviceauth.DeviceAuthDevice{
				{
					ID:     "1",
					Status: "active",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:44",
					},
				},
				{
					ID:     "2",
					Status: "pending",
					IdDataStruct: map[string]string{
						"mac": "00:11:22:33:55",
					},
				},
			},

			inventoryDeviceIDs: []string{"1", "2"},
			inventoryDevices: []inventory.Device{
				{
					ID: "1",
				},
				{
					ID: "2",
				},
			},

			bulkIndexDevices: []*model.Device{
				{
					ID:       strptr("1"),
					TenantID: strptr(tenantID),
					IdentityAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeIdentity,
							Name:   model.AttrNameStatus,
							String: []string{"active"},
						},
						{
							Scope:  model.ScopeIdentity,
							Name:   "mac",
							String: []string{"00:11:22:33:44"},
						},
					},
				},
				{
					ID:       strptr("2"),
					TenantID: strptr(tenantID),
					IdentityAttributes: model.InventoryAttributes{
						{
							Scope:  model.ScopeIdentity,
							Name:   model.AttrNameStatus,
							String: []string{"pending"},
						},
						{
							Scope:  model.ScopeIdentity,
							Name:   "mac",
							String: []string{"00:11:22:33:55"},
						},
					},
				},
			},
			bulkIndexRemoveDevices: []*model.Device{
				{
					ID:       strptr("3"),
					TenantID: strptr(tenantID),
				},
			},
			bulkIndexErr: errors.New("bulk index error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			store := &store_mocks.Store{}
			defer store.AssertExpectations(t)

			if len(tc.bulkIndexDevices) > 0 || len(tc.bulkIndexRemoveDevices) > 0 {
				store.On("BulkIndexDevices",
					ctx,
					tc.bulkIndexDevices,
					tc.bulkIndexRemoveDevices,
				).Return(tc.bulkIndexErr)
			}

			devClient := &deviceauth_mocks.Client{}
			defer devClient.AssertExpectations(t)

			devClient.On("GetDevices",
				ctx,
				tenantID,
				mock.MatchedBy(func(ids []string) bool {
					sort.Strings(ids)
					assert.Equal(t, ids, tc.deviceauthDeviceIDs)

					return true
				}),
			).Return(tc.deviceauthDevices, tc.deviceauthErr)

			invClient := &inventory_mocks.Client{}
			defer invClient.AssertExpectations(t)

			if tc.deviceauthErr == nil {
				invClient.On("GetDevices",
					ctx,
					tenantID,
					mock.MatchedBy(func(ids []string) bool {
						sort.Strings(ids)
						assert.Equal(t, ids, tc.deviceauthDeviceIDs)

						return true
					}),
				).Return(tc.inventoryDevices, tc.inventoryErr)
			}

			indexer := NewIndexer(store, nil, devClient, invClient)

			indexer.ProcessJobs(ctx, tc.jobs)
		})
	}
}
