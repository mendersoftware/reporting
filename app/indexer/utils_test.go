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
	"testing"

	"github.com/mendersoftware/reporting/model"
	"github.com/stretchr/testify/assert"
)

func TestGroupJobsIntoTenantDeviceServices(t *testing.T) {
	jobs := []*model.Job{
		{
			TenantID: "t1",
			DeviceID: "d1",
			Service:  model.ServiceInventory,
		},
		{
			TenantID: "t1",
			DeviceID: "d1",
			Service:  model.ServiceDeviceauth,
		},
		{
			TenantID: "t1",
			DeviceID: "d2",
			Service:  model.ServiceInventory,
		},
		{
			TenantID: "t2",
			DeviceID: "d1",
			Service:  model.ServiceInventory,
		},
	}

	tenantDevicesServices := groupJobsIntoTenantDeviceServices(jobs)
	expected := TenantDeviceServices{
		"t1": DeviceServices{
			"d1": Services{
				model.ServiceInventory:  true,
				model.ServiceDeviceauth: true,
			},
			"d2": Services{
				model.ServiceInventory: true,
			},
		},
		"t2": DeviceServices{
			"d1": Services{
				model.ServiceInventory: true,
			},
		},
	}

	assert.Equal(t, expected, tenantDevicesServices)
}
