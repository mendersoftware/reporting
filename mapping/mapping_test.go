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
	"errors"
	"fmt"
	"testing"

	"github.com/mendersoftware/reporting/client/inventory"
	"github.com/mendersoftware/reporting/model"
	"github.com/mendersoftware/reporting/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewMapper(t *testing.T) {
	m := NewMapper(nil)
	assert.NotNil(t, m)
}

func TestMapInventoryAttributes(t *testing.T) {
	const tenantID = "tenantID"
	testCases := map[string]struct {
		attrs   inventory.DeviceAttributes
		update  bool
		mapping *model.Mapping
		out     inventory.DeviceAttributes
		err     error
	}{
		"ok": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: inventory.AttrScopeInventory},
				{Name: "a2", Value: "v2", Scope: inventory.AttrScopeInventory},
				{Name: "a3", Value: "v3", Scope: inventory.AttrScopeSystem},
			},
			update: true,
			mapping: &model.Mapping{
				TenantID:  tenantID,
				Inventory: []string{"a1", "a2"},
			},
			out: inventory.DeviceAttributes{
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: inventory.AttrScopeInventory},
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: inventory.AttrScopeInventory},
				{Name: "a3", Value: "v3", Scope: inventory.AttrScopeSystem},
			},
		},
		"ok, no update": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: inventory.AttrScopeInventory},
				{Name: "a2", Value: "v2", Scope: inventory.AttrScopeInventory},
				{Name: "a3", Value: "v3", Scope: inventory.AttrScopeInventory},
			},
			update: false,
			mapping: &model.Mapping{
				TenantID:  tenantID,
				Inventory: []string{"a1", "a2"},
			},
			out: inventory.DeviceAttributes{
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: inventory.AttrScopeInventory},
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: inventory.AttrScopeInventory},
				{Name: "a3", Value: "v3", Scope: inventory.AttrScopeInventory},
			},
		},
		"error": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: inventory.AttrScopeInventory},
				{Name: "a2", Value: "v2", Scope: inventory.AttrScopeInventory},
			},
			update: true,
			err:    errors.New("error"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			ds := &mocks.DataStore{}
			if tc.update {
				ds.On("UpdateAndGetMapping",
					ctx,
					tenantID,
					mock.AnythingOfType("[]string"),
				).Return(tc.mapping, tc.err)
			} else {
				ds.On("GetMapping",
					ctx,
					tenantID,
				).Return(tc.mapping, tc.err)
			}

			mapper := NewMapper(ds)
			attrs, err := mapper.MapInventoryAttributes(ctx, tenantID, tc.attrs, tc.update)
			if tc.err != nil {
				assert.Equal(t, tc.err, err)
				assert.Nil(t, attrs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.out, attrs)
			}
		})
	}
}

func TestReverseInventoryAttributes(t *testing.T) {
	const tenantID = "tenantID"
	testCases := map[string]struct {
		attrs   inventory.DeviceAttributes
		mapping *model.Mapping
		out     inventory.DeviceAttributes
		err     error
	}{
		"ok": {
			attrs: inventory.DeviceAttributes{
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: inventory.AttrScopeInventory},
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: inventory.AttrScopeInventory},
				{Name: "a3", Value: "v3", Scope: inventory.AttrScopeSystem},
			},
			mapping: &model.Mapping{
				TenantID:  tenantID,
				Inventory: []string{"a1", "a2"},
			},
			out: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: inventory.AttrScopeInventory},
				{Name: "a2", Value: "v2", Scope: inventory.AttrScopeInventory},
				{Name: "a3", Value: "v3", Scope: inventory.AttrScopeSystem},
			},
		},
		"error": {
			attrs: inventory.DeviceAttributes{
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: inventory.AttrScopeInventory},
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v2", Scope: inventory.AttrScopeInventory},
			},
			err: errors.New("error"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			ds := &mocks.DataStore{}
			ds.On("GetMapping",
				ctx,
				tenantID,
			).Return(tc.mapping, tc.err)

			mapper := NewMapper(ds)
			attrs, err := mapper.ReverseInventoryAttributes(ctx, tenantID, tc.attrs)
			if tc.err != nil {
				assert.Equal(t, tc.err, err)
				assert.Nil(t, attrs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.out, attrs)
			}
		})
	}
}

func TestGetMapping(t *testing.T) {
	const tenantID = "tenantID"
	testCases := map[string]struct {
		attrs            inventory.DeviceAttributes
		inventoryMapping []string
		mapping          *model.Mapping
		err              error
	}{
		"ok": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: inventory.AttrScopeInventory},
				{Name: "a2", Value: "v2", Scope: inventory.AttrScopeInventory},
			},
			inventoryMapping: []string{"a1", "a2"},
			mapping: &model.Mapping{
				TenantID:  tenantID,
				Inventory: []string{"a1", "a2"},
			},
		},
		"error": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: inventory.AttrScopeInventory},
				{Name: "a2", Value: "v2", Scope: inventory.AttrScopeInventory},
				{Name: "a3", Value: "v2", Scope: inventory.AttrScopeInventory},
				{Name: "a2", Value: "v2", Scope: inventory.AttrScopeInventory},
			},
			inventoryMapping: []string{"a1", "a2", "a3", "a2"},
			err:              errors.New("error"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			ds := &mocks.DataStore{}
			ds.On("UpdateAndGetMapping",
				ctx,
				tenantID,
				tc.inventoryMapping,
			).Return(tc.mapping, tc.err)

			mapper := &mapper{ds: ds}
			mapping, err := mapper.updateAndGetMapping(ctx, tenantID, tc.attrs)
			if tc.err != nil {
				assert.Equal(t, tc.err, err)
				assert.Nil(t, mapping)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mapping, mapping)
			}
		})
	}
}

func TestMapAttributes(t *testing.T) {
	testCases := map[string]struct {
		attrs   inventory.DeviceAttributes
		mapping map[string]string
		out     inventory.DeviceAttributes
	}{
		"case 1": {
			attrs: inventory.DeviceAttributes{
				{Name: "a1", Value: "v1", Scope: inventory.AttrScopeInventory},
				{Name: "a2", Value: "v2", Scope: inventory.AttrScopeInventory},
			},
			mapping: map[string]string{
				"a1": fmt.Sprintf(inventoryAttributeTemplate, 1),
				"a2": fmt.Sprintf(inventoryAttributeTemplate, 2),
			},
			out: inventory.DeviceAttributes{
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 1), Value: "v1", Scope: inventory.AttrScopeInventory},
				{Name: fmt.Sprintf(inventoryAttributeTemplate, 2), Value: "v2", Scope: inventory.AttrScopeInventory},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			out := mapAttributes(tc.attrs, tc.mapping)
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestAttributesToFields(t *testing.T) {
	testCases := map[string]struct {
		in  []string
		out map[string]string
	}{
		"case 1": {
			in: []string{"a1", "a2"},
			out: map[string]string{
				"a1": fmt.Sprintf(inventoryAttributeTemplate, 1),
				"a2": fmt.Sprintf(inventoryAttributeTemplate, 2),
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			out := attributesToFields(tc.in)
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestFieldsToAttributes(t *testing.T) {
	testCases := map[string]struct {
		in  []string
		out map[string]string
	}{
		"case 1": {
			in: []string{"a1", "a2"},
			out: map[string]string{
				fmt.Sprintf(inventoryAttributeTemplate, 1): "a1",
				fmt.Sprintf(inventoryAttributeTemplate, 2): "a2",
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			out := fieldsToAttributes(tc.in)
			assert.Equal(t, tc.out, out)
		})
	}
}
