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
	"errors"
	"strings"
	"time"
)

const (
	StatusAccepted = "accepted"
	StatusPending  = "pending"
)

type Device struct {
	ID                  *string         `json:"id"`
	TenantID            *string         `json:"tenantID,omitempty"`
	Name                *string         `json:"name,omitempty"`
	GroupName           *string         `json:"groupName,omitempty"`
	Status              *string         `json:"status,omitempty"`
	IdentityAttributes  DeviceInventory `json:"identityAttributes,omitempty"`
	InventoryAttributes DeviceInventory `json:"inventoryAttributes,omitempty"`
	MonitorAttributes   DeviceInventory `json:"monitorAttributes,omitempty"`
	SystemAttributes    DeviceInventory `json:"systemAttributes,omitempty"`
	TagsAttributes      DeviceInventory `json:"tagsAttributes,omitempty"`
	CreatedAt           *time.Time      `json:"createdAt,omitempty"`
	UpdatedAt           *time.Time      `json:"updatedAt,omitempty"`
	Meta                *DeviceMeta     `json:"-"`
}

type DeviceMeta struct {
	SeqNo       int64
	PrimaryTerm int64
}

func (d *Device) WithMeta(m *DeviceMeta) *Device {
	d.Meta = m
	return d
}

func NewDevice(id string) *Device {
	return &Device{
		ID: &id,
	}
}

func NewDeviceFromInv(tenant string, invdev *InvDevice) (*Device, error) {
	dev := NewDevice(string(invdev.ID))
	dev.SetTenantID(tenant)

	// rewrite attributes
	// special treatment for some attributes which become fields as well
	for _, invattr := range invdev.Attributes {
		attr := NewInventoryAttribute(invattr.Scope)

		attr.SetName(invattr.Name).
			SetVal(invattr.Value)

		if err := dev.AppendAttr(attr); err != nil {
			return nil, err
		}

		dev.handleSpecialAttr(attr)
	}

	return dev, nil
}

// NewDeviceFromEsSource parses the ES '_source' into a new Device
func NewDeviceFromEsSource(source map[string]interface{}) (*Device, error) {

	// for simplicity, let any type assertions just panic
	dev := NewDevice(source["id"].(string))
	dev.SetTenantID(source["tenantID"].(string))

	for k, v := range source {
		s, n, err := MaybeParseAttr(k)

		if err != nil {
			return nil, err
		}

		if n != "" {
			attr := NewInventoryAttribute(s).
				SetName(Redot(n)).
				SetVal(v)

			dev.handleSpecialAttr(attr)
			if err := dev.AppendAttr(attr); err != nil {
				return nil, err
			}
		}
	}

	return dev, nil
}

// setSpecialAttr detects if the attribute should be promoted to a Device field
func (a *Device) handleSpecialAttr(attr *InventoryAttribute) {
	if attr.Scope == scopeIdentity && attr.Name == AttrNameStatus {
		a.SetStatus(attr.GetString())
	}

	if attr.Scope == scopeSystem && attr.Name == AttrNameGroup {
		a.SetGroupName(attr.GetString())
	}
}

func (a *Device) AppendAttr(attr *InventoryAttribute) error {
	switch attr.Scope {
	case scopeIdentity:
		a.IdentityAttributes = append(a.IdentityAttributes, attr)
		return nil
	case scopeInventory:
		a.InventoryAttributes = append(a.InventoryAttributes, attr)
		return nil
	case scopeMonitor:
		a.MonitorAttributes = append(a.MonitorAttributes, attr)
		return nil
	case scopeSystem:
		a.SystemAttributes = append(a.SystemAttributes, attr)
		return nil
	case scopeTags:
		a.TagsAttributes = append(a.TagsAttributes, attr)
		return nil
	default:
		return errors.New("unknown attribute scope " + attr.Scope)
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
	Boolean []bool
}

func (a *InventoryAttribute) IsStr() bool {
	return a.String != nil
}

func (a *InventoryAttribute) IsNum() bool {
	return a.Numeric != nil
}

func (a *InventoryAttribute) IsBool() bool {
	return a.Boolean != nil
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
	a.Boolean = nil
	a.Numeric = nil
	return a
}

func (a *InventoryAttribute) GetStrings() []string {
	return a.String
}

func (a *InventoryAttribute) SetStrings(val []string) *InventoryAttribute {
	a.String = val
	a.Boolean = nil
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
	a.Boolean = nil
	a.String = nil
	return a
}

func (a *InventoryAttribute) SetNumerics(val []float64) *InventoryAttribute {
	a.Numeric = val
	a.String = nil
	a.Boolean = nil
	return a
}

func (a *InventoryAttribute) SetBoolean(val bool) *InventoryAttribute {
	a.Boolean = []bool{val}
	a.Numeric = nil
	a.String = nil
	return a
}

func (a *InventoryAttribute) SetBooleans(val []bool) *InventoryAttribute {
	a.Boolean = val
	a.Numeric = nil
	a.String = nil
	return a
}

// SetVal inspects the 'val' type and sets the correct subtype field
// useful for translating from inventory attributes (interface{})
func (a *InventoryAttribute) SetVal(val interface{}) *InventoryAttribute {
	switch val := val.(type) {
	case bool:
		a.SetBoolean(val)
	case float64:
		a.SetNumeric(val)
	case string:
		a.SetString(val)
	case []interface{}:
		switch val[0].(type) {
		case bool:
			bools := make([]bool, len(val))
			for i, v := range val {
				bools[i] = v.(bool)
			}
			a.SetBooleans(bools)
		case float64:
			nums := make([]float64, len(val))
			for i, v := range val {
				nums[i] = v.(float64)
			}
			a.SetNumerics(nums)
		case string:
			strs := make([]string, len(val))
			for i, v := range val {
				strs[i] = v.(string)
			}
			a.SetStrings(strs)
		}
	}

	return a
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

	attributes := append(d.IdentityAttributes, d.InventoryAttributes...)
	attributes = append(attributes, d.MonitorAttributes...)
	attributes = append(attributes, d.SystemAttributes...)
	attributes = append(attributes, d.TagsAttributes...)

	for _, a := range attributes {
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
	} else if a.IsNum() {
		typ = TypeNum
		val = a.Numeric
	} else if a.IsBool() {
		typ = TypeBool
		val = a.Boolean
	}

	name := ToAttr(a.Scope, a.Name, typ)

	return name, val
}

// maybeParseAttr decides if a given field is an attribute and parses
// it's name + scope
func MaybeParseAttr(field string) (string, string, error) {
	scope := ""
	name := ""

	for _, s := range []string{scopeIdentity, scopeInventory, scopeMonitor,
		scopeSystem, scopeTags} {
		if strings.HasPrefix(field, s+"_") {
			scope = s
			break
		}
	}

	if scope != "" {
		for _, s := range []string{typeStr, typeNum} {
			if strings.HasSuffix(field, "_"+s) {
				// strip the prefix/suffix
				start := strings.Index(field, "_")
				end := strings.LastIndex(field, "_")

				name = field[start+1 : end]
			}
		}
	}

	return scope, name, nil
}
