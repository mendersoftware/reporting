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

	natsio "github.com/nats-io/nats.go"
	"github.com/pkg/errors"

	"github.com/mendersoftware/go-lib-micro/config"
	"github.com/mendersoftware/go-lib-micro/log"

	"github.com/mendersoftware/reporting/client/deviceauth"
	rconfig "github.com/mendersoftware/reporting/config"
	"github.com/mendersoftware/reporting/model"
)

type Services map[string]bool
type DeviceServices map[string]Services
type TenantDeviceServices map[string]DeviceServices

func (i *indexer) GetJobs(ctx context.Context, jobs chan *model.Job) error {
	l := log.FromContext(ctx)

	streamName := config.Config.GetString(rconfig.SettingNatsStreamName)
	topic := config.Config.GetString(rconfig.SettingNatsSubscriberTopic)
	subject := streamName + "." + topic
	durableName := config.Config.GetString(rconfig.SettingNatsSubscriberDurable)

	channel := make(chan *natsio.Msg, 1)
	unsubscribe, err := i.nats.JetStreamSubscribe(ctx, subject, durableName, channel)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to the nats JetStream")
	}

	go func() {
		l.Info("Reindexer ready to receive messages")
		defer func() {
			_ = unsubscribe()
		}()

		for {
			select {
			case msg := <-channel:
				job := &model.Job{}
				err := json.Unmarshal(msg.Data, job)
				if err != nil {
					err = errors.Wrap(err, "failed to unmarshall message")
					l.Error(err)
					if err := msg.Term(); err != nil {
						err = errors.Wrap(err, "failed to term the message")
						l.Error(err)
					}
					continue
				}
				if err = msg.Ack(); err != nil {
					err = errors.Wrap(err, "failed to ack the message")
					l.Error(err)
				}
				jobs <- job

			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (i *indexer) ProcessJobs(ctx context.Context, jobs []*model.Job) {
	l := log.FromContext(ctx)

	devices := make([]*model.Device, 0, len(jobs))
	removedDevices := make([]*model.Device, 0, len(jobs))

	l.Debugf("Processing %d jobs", len(jobs))
	tenantsDevicesServices := groupJobsIntoTenantDeviceServices(jobs)
	for tenant, deviceServices := range tenantsDevicesServices {
		deviceIDs := make([]string, 0, len(deviceServices))
		for deviceID, _ := range deviceServices {
			deviceIDs = append(deviceIDs, deviceID)
		}
		// get devices from deviceauth
		deviceAuthDevices, err := i.devClient.GetDevices(ctx, tenant, deviceIDs)
		if err != nil {
			l.Error(errors.Wrap(err, "failed to get devices from deviceauth"))
			continue
		}
		// get devices from inventory
		inventoryDevices, err := i.invClient.GetDevices(ctx, tenant, deviceIDs)
		if err != nil {
			l.Error(errors.Wrap(err, "failed to get devices from inventory"))
			continue
		}
		// process the results
		devices = devices[:0]
		removedDevices = removedDevices[:0]
		for _, deviceID := range deviceIDs {
			var deviceAuthDevice *deviceauth.DeviceAuthDevice
			var inventoryDevice *model.InvDevice
			for _, d := range deviceAuthDevices {
				if d.ID == deviceID {
					deviceAuthDevice = &d
					break
				}
			}
			for _, d := range inventoryDevices {
				if d.ID == model.DeviceID(deviceID) {
					inventoryDevice = &d
					break
				}
			}
			if deviceAuthDevice == nil || inventoryDevice == nil {
				removedDevices = append(removedDevices, &model.Device{
					ID:       &deviceID,
					TenantID: &tenant,
				})
				continue
			}
			device, err := model.NewDeviceFromInv(tenant, inventoryDevice)
			if err != nil {
				err = errors.Wrapf(err,
					"failed to convert the inventory device for tenant %s, "+
						"device %s", tenant, deviceID)
				l.Error(err)
			}
			// data from device auth
			_ = device.AppendAttr(&model.InventoryAttribute{
				Scope:  model.AttrScopeIdentity,
				Name:   "status",
				String: []string{deviceAuthDevice.Status},
			})
			for name, value := range deviceAuthDevice.IdDataStruct {
				if err := device.AppendAttr(&model.InventoryAttribute{
					Scope:  model.AttrScopeIdentity,
					Name:   name,
					String: []string{value},
				}); err != nil {
					err = errors.Wrapf(err,
						"failed to convert identity data for tenant %s, "+
							"device %s", tenant, deviceID)
					l.Error(err)
				}
			}
			// append the device
			devices = append(devices, device)
		}
		// bulk index the device
		if len(devices) > 0 || len(removedDevices) > 0 {
			err = i.store.BulkIndexDevices(ctx, devices, removedDevices)
			if err != nil {
				err = errors.Wrap(err, "failed to bulk index the devices")
				l.Error(err)
			}
		}
	}
}
