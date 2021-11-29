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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchParamsValidate(t *testing.T) {
	testCases := map[string]struct {
		params SearchParams
		err    error
	}{
		"ok, empty": {
			params: SearchParams{},
		},
		"ok, full example": {
			params: SearchParams{
				Filters: []FilterPredicate{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Type:      "$eq",
						Value:     "00:11:22:33:44",
					},
				},
				Sort: []SortCriteria{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
						Order:     "asc",
					},
				},
				Attributes: []SelectAttribute{
					{
						Scope:     ScopeIdentity,
						Attribute: "mac",
					},
				},
			},
		},
		"ko, filter fails validation": {
			params: SearchParams{
				Filters: []FilterPredicate{
					{
						Value: "",
					},
				},
			},
			err: errors.New("attribute: cannot be blank; scope: cannot be blank; type: cannot be blank."),
		},
		"ko, sort fails validation": {
			params: SearchParams{
				Sort: []SortCriteria{
					{
						Order: "dummy",
					},
				},
			},
			err: errors.New("attribute: cannot be blank; order: must be a valid value; scope: cannot be blank."),
		},
		"ko, attributes fails validation": {
			params: SearchParams{
				Attributes: []SelectAttribute{
					{
						Attribute: "mac",
					},
				}},
			err: errors.New("scope: cannot be blank."),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.err != nil {
				assert.EqualError(t, tc.err, err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestFilterPredicateValueType(t *testing.T) {
	testCases := map[string]struct {
		filterPredicate FilterPredicate

		typ     Type
		isArray bool
		err     error
	}{
		"ok, string": {
			filterPredicate: FilterPredicate{
				Scope:     ScopeIdentity,
				Attribute: "mac",
				Value:     "a",
			},
			typ:     TypeStr,
			isArray: false,
			err:     nil,
		},
		"ok, array of string": {
			filterPredicate: FilterPredicate{
				Scope:     ScopeIdentity,
				Attribute: "mac",
				Value:     []interface{}{"a"},
			},
			typ:     TypeStr,
			isArray: true,
			err:     nil,
		},
		"ok, number": {
			filterPredicate: FilterPredicate{
				Scope:     ScopeIdentity,
				Attribute: "mac",
				Value:     float64(1.0),
			},
			typ:     TypeNum,
			isArray: false,
			err:     nil,
		},
		"ok, array of numbers": {
			filterPredicate: FilterPredicate{
				Scope:     ScopeIdentity,
				Attribute: "mac",
				Value:     []interface{}{float64(1.0)},
			},
			typ:     TypeNum,
			isArray: true,
			err:     nil,
		},
		"ok, bool": {
			filterPredicate: FilterPredicate{
				Scope:     ScopeIdentity,
				Attribute: "mac",
				Value:     true,
			},
			typ:     TypeBool,
			isArray: false,
			err:     nil,
		},
		"ok, array of bools": {
			filterPredicate: FilterPredicate{
				Scope:     ScopeIdentity,
				Attribute: "mac",
				Value:     []interface{}{true},
			},
			typ:     TypeBool,
			isArray: true,
			err:     nil,
		},
		"ko, wrong type": {
			filterPredicate: FilterPredicate{
				Scope:     ScopeIdentity,
				Attribute: "mac",
				Value:     nil,
			},
			typ:     0,
			isArray: false,
			err:     errors.New("unknown attribute value type: <nil> <nil>"),
		},
		"ko, wrong type in array": {
			filterPredicate: FilterPredicate{
				Scope:     ScopeIdentity,
				Attribute: "mac",
				Value:     []interface{}{nil},
			},
			typ:     0,
			isArray: false,
			err:     errors.New("unknown attribute value type: <nil> <nil>"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			typ, isArray, err := tc.filterPredicate.ValueType()
			assert.Equal(t, tc.typ, typ)
			assert.Equal(t, tc.isArray, isArray)
			if tc.err == nil {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, tc.err, err.Error())
			}
		})
	}
}
