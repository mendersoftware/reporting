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
package model

import (
	"encoding/json"
	"time"
)

// 1:1 port of the inventory device
// for inventory api compat
const (
	AttrScopeInventory = "inventory"
	AttrScopeIdentity  = "identity"
	AttrScopeSystem    = "system"

	AttrNameID      = "id"
	AttrNameGroup   = "group"
	AttrNameStatus  = "status"
	AttrNameUpdated = "updated_ts"
	AttrNameCreated = "created_ts"
)

type DeviceID string
type GroupName string
type DeviceAttributes []InvDeviceAttribute

type InvDeviceAttribute struct {
	Name        string      `json:"name" bson:",omitempty"`
	Description *string     `json:"description,omitempty" bson:",omitempty"`
	Value       interface{} `json:"value" bson:",omitempty"`
	Scope       string      `json:"scope" bson:",omitempty"`
}

// Device wrapper
type InvDevice struct {
	//system-generated device ID
	ID DeviceID `json:"id" bson:"_id,omitempty"`

	//a map of attributes names and their values.
	Attributes DeviceAttributes `json:"attributes,omitempty" bson:"attributes,omitempty"`

	//device's group name
	Group GroupName `json:"-" bson:"group,omitempty"`

	CreatedTs time.Time `json:"-" bson:"created_ts,omitempty"`
	//Timestamp of the last attribute update.
	UpdatedTs time.Time `json:"updated_ts" bson:"updated_ts,omitempty"`

	//device object revision
	Revision uint `json:"-" bson:"revision,omitempty"`
}

func (d *DeviceAttributes) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, (*[]InvDeviceAttribute)(d))
	if err != nil {
		return err
	}
	for i := range *d {
		if (*d)[i].Scope == "" {
			(*d)[i].Scope = AttrScopeInventory
		}
	}

	return nil
}
