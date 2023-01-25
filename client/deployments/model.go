// Copyright 2023 Northern.tech AS
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
	ID         string      `json:"id"`
	Deployment *Deployment `json:"deployment"`
	Device     *Device     `json:"device"`
}

// Deployment contains the definition of a deployment
type Deployment struct {
	Id                 string                   `json:"id"`
	Name               string                   `json:"name,omitempty"`
	ArtifactName       string                   `json:"artifact_name,omitempty"`
	Devices            []string                 `json:"devices,omitempty"`
	FilterId           string                   `json:"filter_id,omitempty"`
	PhaseId            string                   `json:"phase_id,omitempty"`
	AllDevices         bool                     `json:"all_devices,omitempty"`
	ForceInstallation  bool                     `json:"force_installation,omitempty"`
	Group              string                   `json:"group"`
	Created            *time.Time               `json:"created"`
	Finished           *time.Time               `json:"finished,omitempty"`
	Artifacts          []string                 `json:"artifacts,omitempty"`
	Depends            []map[string]interface{} `json:"depends"`
	Status             string                   `json:"status"`
	InitialDeviceCount int                      `json:"initial_device_count,omitempty"`
	DeviceCount        *int                     `json:"device_count"`
	Retries            uint                     `json:"retries,omitempty"`
	Dynamic            bool                     `json:"dynamic,omitempty"`
	MaxDevices         int                      `json:"max_devices,omitempty"`
	Groups             []string                 `json:"groups,omitempty"`
	DeviceList         []string                 `json:"device_list"`
	Type               string                   `json:"type,omitempty"`
	AutogenerateDelta  bool                     `json:"autogenerate_delta,omitempty"`
}

// Device contains the device-specific information for a device deployment
type Device struct {
	Created        *time.Time `json:"created"`
	Finished       *time.Time `json:"finished,omitempty"`
	Deleted        *time.Time `json:"deleted,omitempty"`
	Status         string     `json:"status"`
	DeviceId       string     `json:"device_id"`
	DeploymentId   string     `json:"deployment_id"`
	Id             string     `json:"id"`
	Image          *Image     `json:"image"`
	IsLogAvailable bool       `json:"log"`
	SubState       string     `json:"substate,omitempty"`
	Retries        uint       `json:"retries,omitempty"`
	Attempts       uint       `json:"attempts,omitempty"`
}

type Image struct {
	Id                    string              `json:"id"`
	Description           string              `json:"description"`
	Name                  string              `json:"name"`
	DeviceTypesCompatible []string            `json:"device_types_compatible"`
	Info                  *ArtifactInfo       `json:"info"`
	Signed                bool                `json:"signed"`
	Provides              map[string]string   `json:"artifact_provides,omitempty"`
	Depends               map[string][]string `json:"artifact_depends,omitempty"`
	ClearsProvides        []string            `json:"clears_artifact_provides,omitempty"`
	Size                  int64               `json:"size"`
	Modified              *time.Time          `json:"modified" valid:"-"`
}

type ArtifactInfo struct {
	Format  string `json:"format" valid:"required"`
	Version uint   `json:"version" valid:"required"`
}

type InstalledDeviceDeployment struct {
	ArtifactName string            `json:"artifact_name"`
	DeviceType   string            `json:"device_type"`
	Provides     map[string]string `json:"artifact_provides,omitempty"`
}
