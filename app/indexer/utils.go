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

import "github.com/mendersoftware/reporting/model"

func groupJobsIntoTenantDeviceServices(jobs []*model.Job) TenantDeviceServices {
	tenantsDevicesServices := make(TenantDeviceServices)
	for _, job := range jobs {
		if _, ok := tenantsDevicesServices[job.TenantID]; !ok {
			tenantsDevicesServices[job.TenantID] = make(DeviceServices)
		}
		if _, ok := tenantsDevicesServices[job.TenantID][job.DeviceID]; !ok {
			tenantsDevicesServices[job.TenantID][job.DeviceID] = make(Services)
		}
		if _, ok := tenantsDevicesServices[job.TenantID][job.DeviceID][job.Service]; !ok {
			tenantsDevicesServices[job.TenantID][job.DeviceID][job.Service] = true
		}
	}
	return tenantsDevicesServices
}
