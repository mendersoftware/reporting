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

package deployments

import "time"

// DeviceDeployment stores a device deployment
type DeviceDeployment struct {
	ID     string  `json:"id"`
	Device *Device `json:"device"`
}

// Device contains the device-specific information for a device deployment
type Device struct {
	Created        *time.Time             `json:"created"`
	Finished       *time.Time             `json:"finished,omitempty"`
	Deleted        *time.Time             `json:"deleted,omitempty"`
	Status         string                 `json:"status"`
	DeviceId       string                 `json:"device_id"`
	DeploymentId   string                 `json:"deployment_id"`
	Id             string                 `json:"id"`
	Image          *Image                 `json:"image"`
	Request        *DeploymentNextRequest `json:"request"`
	IsLogAvailable bool                   `json:"log"`
	SubState       string                 `json:"substate,omitempty"`
	Retries        uint                   `json:"retries,omitempty"`
	Attempts       uint                   `json:"attempts,omitempty"`
}

type Image struct {
	Id                    string                 `json:"id"`
	Description           string                 `json:"description"`
	Name                  string                 `json:"name"`
	DeviceTypesCompatible []string               `json:"device_types_compatible"`
	Info                  *ArtifactInfo          `json:"info"`
	Signed                bool                   `json:"signed"`
	Provides              Provides               `json:"artifact_provides,omitempty"`
	Depends               map[string]interface{} `json:"artifact_depends,omitempty"`
	ClearsProvides        []string               `json:"clears_artifact_provides,omitempty"`
	Size                  int64                  `json:"size"`
	Modified              *time.Time             `json:"modified" valid:"-"`
}

type ArtifactInfo struct {
	Format  string `json:"format" valid:"required"`
	Version uint   `json:"version" valid:"required"`
}

type Provides map[string]string

type DeploymentNextRequest struct {
	DeviceProvides   *InstalledDeviceDeployment `json:"device_provides"`
	UpdateControlMap bool                       `json:"update_control_map"`
}

type InstalledDeviceDeployment struct {
	ArtifactName string            `json:"artifact_name"`
	DeviceType   string            `json:"device_type"`
	Provides     map[string]string `json:"artifact_provides,omitempty"`
}
