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

package indexer

import (
	"context"
	"encoding/json"

	natsio "github.com/nats-io/nats.go"
	"github.com/pkg/errors"

	"github.com/mendersoftware/go-lib-micro/config"
	"github.com/mendersoftware/go-lib-micro/log"

	"github.com/mendersoftware/reporting/client/deviceauth"
	"github.com/mendersoftware/reporting/client/inventory"
	rconfig "github.com/mendersoftware/reporting/config"
	"github.com/mendersoftware/reporting/model"
)

type Services map[string]bool
type DeviceServices map[string]Services
type TenantDeviceServices map[string]DeviceServices

func (i *indexer) GetJobs(ctx context.Context, jobs chan *model.Job) error {
	l := log.FromContext(ctx)

	streamName := config.Config.GetString(rconfig.SettingNatsStreamName)
	err := i.nats.JetStreamCreateStream(streamName)
	if err != nil {
		return errors.Wrap(err, "failed to create the nats JetStream stream")
	}

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
		for deviceID := range deviceServices {
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
			var inventoryDevice *inventory.Device
			for _, d := range deviceAuthDevices {
				if d.ID == deviceID {
					deviceAuthDevice = &d
					break
				}
			}
			for _, d := range inventoryDevices {
				if d.ID == inventory.DeviceID(deviceID) {
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
			device := i.processJobDevice(ctx, tenant, deviceAuthDevice, inventoryDevice)
			if device == nil {
				continue
			}
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

func (i *indexer) processJobDevice(
	ctx context.Context,
	tenant string,
	deviceAuthDevice *deviceauth.DeviceAuthDevice,
	inventoryDevice *inventory.Device,
) *model.Device {
	l := log.FromContext(ctx)
	//
	device := model.NewDevice(tenant, string(inventoryDevice.ID))
	// data from inventory
	device.SetUpdatedAt(inventoryDevice.UpdatedTs)
	attributes, err := i.mapper.MapInventoryAttributes(ctx, tenant,
		inventoryDevice.Attributes, true, false)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to map inventory data for tenant %s, "+
				"device %s", tenant, inventoryDevice.ID)
		l.Warn(err)
	} else {
		for _, invattr := range attributes {
			attr := model.NewInventoryAttribute(invattr.Scope).
				SetName(invattr.Name).
				SetVal(invattr.Value)
			if err := device.AppendAttr(attr); err != nil {
				err = errors.Wrapf(err,
					"failed to convert inventory data for tenant %s, "+
						"device %s", tenant, inventoryDevice.ID)
				l.Warn(err)
			}
		}
	}
	// data from device auth
	_ = device.AppendAttr(&model.InventoryAttribute{
		Scope:  model.ScopeIdentity,
		Name:   model.AttrNameStatus,
		String: []string{deviceAuthDevice.Status},
	})
	for name, value := range deviceAuthDevice.IdDataStruct {
		attr := model.NewInventoryAttribute(model.ScopeIdentity).
			SetName(name).
			SetVal(value)
		if err := device.AppendAttr(attr); err != nil {
			err = errors.Wrapf(err,
				"failed to convert identity data for tenant %s, "+
					"device %s", tenant, inventoryDevice.ID)
			l.Warn(err)
		}
	}
	// latest deployment
	deviceDeployment, err := i.deplClient.GetLatestFinishedDeployment(ctx, tenant,
		string(inventoryDevice.ID))
	if err != nil {
		l.Error(errors.Wrap(err, "failed to get device deployments from deployments"))
		return nil
	} else if deviceDeployment != nil {
		_ = device.AppendAttr(&model.InventoryAttribute{
			Scope:  model.ScopeSystem,
			Name:   model.AttrNameLatestDeploymentStatus,
			String: []string{deviceDeployment.Device.Status},
		})
	}
	// return the device
	return device
}
