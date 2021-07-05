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
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

const (
	StatusAccepted = "accepted"
	StatusPending  = "pending"
)

const (
	attrMacAddress = "mac"
	attrSerialNo   = "serial_no"
	attrTag        = "tag"
)

type Device struct {
	ID                  *string         `json:"id"`
	TenantID            *string         `json:"tenantID,omitempty"`
	Name                *string         `json:"name,omitempty"`
	GroupName           *string         `json:"groupName,omitempty"`
	Status              *string         `json:"status,omitempty"`
	CustomAttributes    DeviceInventory `json:"customAttributes,omitempty"`
	IdentityAttributes  DeviceInventory `json:"identityAttributes,omitempty"`
	InventoryAttributes DeviceInventory `json:"inventoryAttributes,omitempty"`
	CreatedAt           *time.Time      `json:"createdAt,omitempty"`
	UpdatedAt           *time.Time      `json:"updatedAt,omitempty"`
}

func NewDevice(id string) *Device {
	return &Device{
		ID: &id,
	}
}

func (a *Device) GetID() string {
	if a.ID != nil {
		return *a.ID
	}
	return ""
}

func (a *Device) SetID(val string) *Device {
	a.ID = &val
	return a
}

func (a *Device) GetName() string {
	if a.Name != nil {
		return *a.Name
	}
	return ""
}

func (a *Device) SetName(val string) *Device {
	a.Name = &val
	return a
}

func (a *Device) GetTenantID() string {
	if a.TenantID != nil {
		return *a.TenantID
	}
	return ""
}

func (a *Device) SetTenantID(val string) *Device {
	a.TenantID = &val
	return a
}

func (a *Device) GetGroupName() string {
	if a.GroupName != nil {
		return *a.GroupName
	}
	return ""
}

func (a *Device) SetGroupName(val string) *Device {
	a.GroupName = &val
	return a
}

func (a *Device) GetStatus() string {
	if a.Status != nil {
		return *a.Status
	}
	return ""
}

func (a *Device) SetStatus(val string) *Device {
	a.Status = &val
	return a
}

func (a *Device) GetCreatedAt() time.Time {
	if a.CreatedAt != nil {
		return *a.CreatedAt
	}
	return time.Time{}
}

func (a *Device) SetCreatedAt(val time.Time) *Device {
	a.CreatedAt = &val
	return a
}

func (a *Device) GetUpdatedAt() time.Time {
	if a.UpdatedAt != nil {
		return *a.UpdatedAt
	}
	return time.Time{}
}

func (a *Device) SetUpdatedAt(val time.Time) *Device {
	a.UpdatedAt = &val
	return a
}

type DeviceInventory []*InventoryAttribute

type InventoryAttribute struct {
	Scope   string
	Name    string
	String  []string
	Numeric []float64
}

func (a *InventoryAttribute) IsStr() bool {
	return a.String != nil
}

func (a *InventoryAttribute) IsNum() bool {
	return a.Numeric != nil
}

func NewInventoryAttribute(s string) *InventoryAttribute {
	return &InventoryAttribute{
		Scope: s,
	}
}

func (a *InventoryAttribute) GetName() string {
	return a.Name
}

func (a *InventoryAttribute) SetName(val string) *InventoryAttribute {
	a.Name = val
	return a
}

func (a *InventoryAttribute) GetString() string {
	if len(a.String) > 0 {
		return a.String[0]
	}
	return ""
}

func (a *InventoryAttribute) SetString(val string) *InventoryAttribute {
	a.String = []string{val}
	a.Numeric = nil
	return a
}

func (a *InventoryAttribute) GetStrings() []string {
	return a.String
}

func (a *InventoryAttribute) SetStrings(val []string) *InventoryAttribute {
	a.String = val
	a.Numeric = nil
	return a
}

func (a *InventoryAttribute) GetNumeric() float64 {
	if len(a.Numeric) > 0 {
		return a.Numeric[0]
	}
	return float64(0)
}

func (a *InventoryAttribute) SetNumeric(val float64) *InventoryAttribute {
	a.Numeric = []float64{val}
	a.String = nil
	return a
}

func (a *InventoryAttribute) SetNumerics(val []float64) *InventoryAttribute {
	a.Numeric = val
	a.String = nil
	return a
}

func randomMacAddress() string {
	buf := make([]byte, 6)
	_, _ = rand.Read(buf)
	buf[0] |= 2
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
}

func RandomDevice(tid string) *Device {
	id := uuid.New().String()
	device := NewDevice(id)
	device.SetName("device-" + id)
	device.SetTenantID(tid)
	device.SetCreatedAt(time.Now().UTC()).SetUpdatedAt(time.Now().UTC())

	if rand.Intn(10) > 7 {
		device.SetStatus(StatusPending)
	} else {
		device.SetStatus(StatusAccepted)
	}

	groupId := rand.Intn(100)
	device.SetGroupName(fmt.Sprintf("group-%02d", groupId))

	device.CustomAttributes = DeviceInventory{
		NewInventoryAttribute(scopeCustom).
			SetName(attrTag).
			SetString(fmt.Sprintf("value-%02d", rand.Intn(100))),
	}

	macAddress := randomMacAddress()

	device.IdentityAttributes = DeviceInventory{
		NewInventoryAttribute(scopeIdentity).
			SetName(attrMacAddress).
			SetString(macAddress),

		NewInventoryAttribute(scopeIdentity).
			SetName(attrSerialNo).
			SetString(fmt.Sprintf("%012d", rand.Intn(999999999999))),
	}

	device.InventoryAttributes = DeviceInventory{
		NewInventoryAttribute(scopeInventory).
			SetName(attrMacAddress).
			SetString(macAddress),

		NewInventoryAttribute(scopeInventory).
			SetName("artifact_name").
			SetString("system-M1"),

		NewInventoryAttribute(scopeInventory).
			SetName("device_type").
			SetString("dm1"),

		NewInventoryAttribute(scopeInventory).
			SetName("hostname").
			SetString("Ambarella"),

		NewInventoryAttribute(scopeInventory).
			SetName("ipv4_bcm0").
			SetString("192.168.42.1/24"),

		NewInventoryAttribute(scopeInventory).
			SetName("ipv4_usb0").
			SetString("10.0.1.2/8"),

		NewInventoryAttribute(scopeInventory).
			SetName("ipv4_wlan0").
			SetString("192.168.1.111/24"),

		NewInventoryAttribute(scopeInventory).
			SetName("kernel").
			SetString("Linux version 4.14.181 (charles-chang@rdsuper) (gcc version 8.2.1 20180802 (Linaro GCC 8.2-2018.08~dev)) #1 SMP PREEMPT Fri Mar 12 13:21:16 CST 2021"),

		NewInventoryAttribute(scopeInventory).
			SetName("mac_bcm0").
			SetString(macAddress),

		NewInventoryAttribute(scopeInventory).
			SetName("mac_usb0").
			SetString(macAddress),

		NewInventoryAttribute(scopeInventory).
			SetName("mac_wlan0").
			SetString(macAddress),

		NewInventoryAttribute(scopeInventory).
			SetName("mem_total_kB").
			SetNumeric(1020664),

		NewInventoryAttribute(scopeInventory).
			SetName("group_id").
			SetNumeric(float64(groupId)),

		NewInventoryAttribute(scopeInventory).
			SetName("mender_bootloader_integration").
			SetString("unknown"),

		NewInventoryAttribute(scopeInventory).
			SetName("mender_client_version").
			SetString("7cb96ca"),

		NewInventoryAttribute(scopeInventory).
			SetName("network_interfaces").
			SetStrings([]string{"bcm0", "usb0", "wlan0"}),

		NewInventoryAttribute(scopeInventory).
			SetName("os").
			SetString("Ambarella Flexible Linux CV25 (2.5.7) DMS (0.0.0.21B)"),

		NewInventoryAttribute(scopeInventory).
			SetName("rootfs_type").
			SetString("ext4"),

		NewInventoryAttribute(scopeInventory).
			SetName("rootfs_type").
			SetString("ext4"),

		NewInventoryAttribute(scopeInventory).
			SetName("rootfs-image.checksum").
			SetString("dbc44ce5bd57f0c909dfb15a1efd9fd5d4e426c0fa95f18ea2876e1b8a08818f"),

		NewInventoryAttribute(scopeInventory).
			SetName("rootfs-image.version").SetString("system-M1"),
	}

	return device
}

func (d *Device) MarshalJSON() ([]byte, error) {
	// TODO: smarter encoding, without explicit rewrites?
	m := make(map[string]interface{})
	m["id"] = d.ID
	m["tenantID"] = d.TenantID
	m["name"] = d.Name
	m["groupName"] = d.GroupName
	m["status"] = d.Status
	m["createdAt"] = d.CreatedAt
	m["updatedAt"] = d.UpdatedAt

	for _, a := range d.CustomAttributes {
		name, val := a.Map()
		m[name] = val
	}

	for _, a := range d.IdentityAttributes {
		name, val := a.Map()
		m[name] = val
	}

	for _, a := range d.InventoryAttributes {
		name, val := a.Map()
		m[name] = val
	}

	return json.Marshal(m)
}

func (a *InventoryAttribute) Map() (string, interface{}) {
	var val interface{}
	var typ Type

	if a.IsStr() {
		typ = TypeStr
		val = a.String
	}

	if a.IsNum() {
		typ = TypeNum
		val = a.Numeric
	}

	name := ToAttr(a.Scope, a.Name, typ)

	return name, val
}
