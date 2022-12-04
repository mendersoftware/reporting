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

package mapping

import (
	"context"
	"fmt"

	"github.com/mendersoftware/reporting/client/inventory"
	"github.com/mendersoftware/reporting/model"
	"github.com/mendersoftware/reporting/store"
)

const (
	inventoryAttributeTemplate = "attribute%d"
)

// Mapping is an interface to map and reverse attributes
type Mapper interface {
	MapInventoryAttributes(ctx context.Context, tenantID string,
		attrs inventory.DeviceAttributes, update bool) (inventory.DeviceAttributes, error)
	ReverseInventoryAttributes(ctx context.Context, tenantID string,
		attrs inventory.DeviceAttributes) (inventory.DeviceAttributes, error)
}

type mapper struct {
	ds store.DataStore
}

func NewMapper(ds store.DataStore) Mapper {
	return &mapper{
		ds: ds,
	}
}

// MapInventoryAttribute maps an inventory attribute to an ES field
func (m *mapper) MapInventoryAttributes(ctx context.Context, tenantID string,
	attrs inventory.DeviceAttributes, update bool) (inventory.DeviceAttributes, error) {
	var mapping *model.Mapping
	var err error
	if update {
		mapping, err = m.updateAndGetMapping(ctx, tenantID, attrs)
	} else {
		mapping, err = m.getMapping(ctx, tenantID)
	}
	if err != nil {
		return nil, err
	}
	attributesToFields := attributesToFields(mapping.Inventory)
	return mapAttributes(attrs, attributesToFields), nil
}

// ReverseInventoryAttribute looks up the inventory attribute name from the ES field
func (m *mapper) ReverseInventoryAttributes(ctx context.Context, tenantID string,
	attrs inventory.DeviceAttributes) (inventory.DeviceAttributes, error) {
	mapping, err := m.getMapping(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	attributesToFields := fieldsToAttributes(mapping.Inventory)
	return mapAttributes(attrs, attributesToFields), nil
}

func (m *mapper) getMapping(ctx context.Context, tenantID string) (*model.Mapping, error) {
	return m.ds.GetMapping(ctx, tenantID)
}

func (m *mapper) updateAndGetMapping(ctx context.Context, tenantID string,
	attrs inventory.DeviceAttributes) (*model.Mapping, error) {
	inventoryMapping := make([]string, 0, len(attrs))
	for i := 0; i < len(attrs); i++ {
		if attrs[i].Scope == inventory.AttrScopeInventory {
			inventoryMapping = append(inventoryMapping, attrs[i].Name)
		}
	}
	mapping, err := m.ds.UpdateAndGetMapping(ctx, tenantID, inventoryMapping)
	if err != nil {
		return nil, err
	}
	return mapping, nil
}

func mapAttributes(attrs inventory.DeviceAttributes,
	mapping map[string]string) inventory.DeviceAttributes {
	mappedAttrs := make(inventory.DeviceAttributes, 0, len(attrs))
	for i := 0; i < len(attrs); i++ {
		var attrName string
		if attrs[i].Scope != inventory.AttrScopeInventory {
			attrName = attrs[i].Name
		} else if name, ok := mapping[attrs[i].Name]; ok {
			attrName = name
		} else {
			attrName = attrs[i].Name
		}
		mappedAttr := inventory.DeviceAttribute{
			Name:  attrName,
			Value: attrs[i].Value,
			Scope: attrs[i].Scope,
		}
		mappedAttrs = append(mappedAttrs, mappedAttr)
	}
	return mappedAttrs
}

func attributesToFields(attrs []string) map[string]string {
	var attributesToFields = make(map[string]string)
	for i := 0; i < len(attrs); i++ {
		attributesToFields[attrs[i]] = fmt.Sprintf(inventoryAttributeTemplate, i+1)
	}
	return attributesToFields
}

func fieldsToAttributes(attrs []string) map[string]string {
	var fieldsToAttributes = make(map[string]string)
	for i := 0; i < len(attrs); i++ {
		fieldsToAttributes[fmt.Sprintf(inventoryAttributeTemplate, i+1)] = attrs[i]
	}
	return fieldsToAttributes
}
