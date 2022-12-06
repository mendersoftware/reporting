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
	"math"
	"sync"

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

type tenantMapCache struct {
	inventory        map[string]string
	inventoryReverse map[string]string
}

type mapper struct {
	ds    store.DataStore
	cache map[string]*tenantMapCache
	lock  sync.RWMutex
}

func NewMapper(ds store.DataStore) Mapper {
	return newMapper(ds)
}

func newMapper(ds store.DataStore) *mapper {
	return &mapper{
		ds:    ds,
		cache: make(map[string]*tenantMapCache),
		lock:  sync.RWMutex{},
	}
}

// MapInventoryAttribute maps an inventory attribute to an ES field
func (m *mapper) MapInventoryAttributes(ctx context.Context, tenantID string,
	attrs inventory.DeviceAttributes, update bool) (inventory.DeviceAttributes, error) {
	attributesToFieldsMap := m.lookupMapping(tenantID, attrs, false)
	if attributesToFieldsMap == nil {
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
		n := int(math.Min(float64(len(mapping.Inventory)), model.MaxMappingInventoryAttributes))
		attributesToFieldsMap = attributesToFields(mapping.Inventory[:n])
	}
	return mapAttributes(attrs, attributesToFieldsMap), nil
}

// ReverseInventoryAttribute looks up the inventory attribute name from the ES field
func (m *mapper) ReverseInventoryAttributes(ctx context.Context, tenantID string,
	attrs inventory.DeviceAttributes) (inventory.DeviceAttributes, error) {
	attributesToFieldsMap := m.lookupMapping(tenantID, attrs, true)
	if attributesToFieldsMap == nil {
		mapping, err := m.getMapping(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		n := int(math.Min(float64(len(mapping.Inventory)), model.MaxMappingInventoryAttributes))
		attributesToFieldsMap = fieldsToAttributes(mapping.Inventory[:n])
	}
	return mapAttributes(attrs, attributesToFieldsMap), nil
}

func (m *mapper) getMapping(ctx context.Context, tenantID string) (*model.Mapping, error) {
	mapping, err := m.ds.GetMapping(ctx, tenantID)
	if err == nil {
		m.cacheMapping(tenantID, mapping)
	}
	return mapping, err
}

func (m *mapper) cacheMapping(tenantID string, mapping *model.Mapping) {
	cache := &tenantMapCache{
		inventory:        make(map[string]string),
		inventoryReverse: make(map[string]string),
	}
	n := int(math.Min(float64(len(mapping.Inventory)), model.MaxMappingInventoryAttributes))
	for i, attr := range mapping.Inventory[:n] {
		attrName := fmt.Sprintf(inventoryAttributeTemplate, i+1)
		cache.inventory[attr] = attrName
		cache.inventoryReverse[attrName] = attr
	}
	m.lock.Lock()
	m.cache[tenantID] = cache
	m.lock.Unlock()
}

func (m *mapper) lookupMapping(tenantID string, attrs inventory.DeviceAttributes,
	reverse bool) map[string]string {
	m.lock.RLock()
	cache, ok := m.cache[tenantID]
	m.lock.RUnlock()
	if ok {
		var cacheAttributes map[string]string
		if reverse {
			cacheAttributes = cache.inventoryReverse
		} else {
			cacheAttributes = cache.inventory
		}
		if len(cacheAttributes) < model.MaxMappingInventoryAttributes {
			for i := 0; i < len(attrs); i++ {
				if attrs[i].Scope == inventory.AttrScopeInventory {
					if _, ok := cacheAttributes[attrs[i].Name]; !ok {
						return nil
					}
				}
			}
		}
		return cacheAttributes
	}
	return nil
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
	m.cacheMapping(tenantID, mapping)
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
