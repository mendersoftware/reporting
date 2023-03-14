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

package model

import "time"

//nolint:lll
type Deployment struct {
	ID                          string                 `json:"id"`
	TenantID                    string                 `json:"tenant_id"`
	DeviceID                    string                 `json:"device_id"`
	DeploymentID                string                 `json:"deployment_id"`
	DeploymentName              string                 `json:"deployment_name"`
	DeploymentArtifactName      string                 `json:"deployment_artifact_name"`
	DeploymentType              string                 `json:"deployment_type"`
	DeploymentCreated           *time.Time             `json:"deployment_created"`
	DeploymentFilterID          string                 `json:"deployment_filter_id,omitempty"`
	DeploymentAllDevices        bool                   `json:"deployment_all_devices"`
	DeploymentForceInstallation bool                   `json:"deployment_force_installation"`
	DeploymentGroup             string                 `json:"deployment_group,omitempty"`
	DeploymentPhased            bool                   `json:"deployment_phased"`
	DeploymentPhaseId           string                 `json:"deployment_phase_id,omitempty"`
	DeploymentRetries           uint                   `json:"deployment_retries"`
	DeploymentMaxDevices        uint                   `json:"deployment_max_devices"`
	DeploymentAutogenerateDelta bool                   `json:"deployment_autogenerate_deta"`
	DeviceCreated               *time.Time             `json:"device_created"`
	DeviceFinished              *time.Time             `json:"device_finished"`
	DeviceElapsedSeconds        uint                   `json:"device_elapsed_seconds"`
	DeviceDeleted               *time.Time             `json:"device_deleted,omitempty"`
	DeviceStatus                string                 `json:"device_status"`
	DeviceIsLogAvailable        bool                   `json:"device_is_log_available"`
	DeviceRetries               uint                   `json:"device_retries"`
	DeviceAttempts              uint                   `json:"device_attempts"`
	ImageID                     string                 `json:"image_id,omitempty"`
	ImageDescription            string                 `json:"image_description,omitempty"`
	ImageArtifactName           string                 `json:"image_artifact_name"`
	ImageDeviceTypes            []string               `json:"image_device_types"`
	ImageSigned                 bool                   `json:"image_signed"`
	ImageArtifactInfoFormat     string                 `json:"image_artifact_info_format,omitempty"`
	ImageArtifactInfoVersion    uint                   `json:"image_artifact_info_version,omitempty"`
	ImageProvides               map[string]string      `json:"image_provides,omitempty"`
	ImageDepends                map[string]interface{} `json:"image_depends,omitempty"`
	ImageClearsProvides         []string               `json:"image_clears_provides,omitempty"`
	ImageSize                   int64                  `json:"image_size,omitempty"`
}
