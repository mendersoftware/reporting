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
package reporting

import (
	"context"
	"errors"
	"testing"

	"github.com/mendersoftware/reporting/model"
	mstore "github.com/mendersoftware/reporting/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var contextMatcher = mock.MatchedBy(func(_ context.Context) bool { return true })

func TestInventorySearchDevices(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name string

		Params *model.SearchParams
		Store  func(*testing.T, testCase) *mstore.Store

		Result     []model.InvDevice
		TotalCount int
		Error      error
	}
	testCases := []testCase{{
		Name: "ok",

		Params: &model.SearchParams{
			Filters: []model.FilterPredicate{{
				Attribute: "foo",
				Value:     "bar",
				Scope:     "inventory",
				Type:      "$eq",
			}},
			Sort: []model.SortCriteria{{
				Attribute: "foo",
				Scope:     "inventory",
				Order:     "desc",
			}},
			DeviceIDs: []string{"194d1060-1717-44dc-a783-00038f4a8013"},
		},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildQuery(*self.Params)
			q = q.Must(model.M{"terms": model.M{"id": self.Params.DeviceIDs}})
			store.On("Search", contextMatcher, q).
				Return(model.M{"hits": map[string]interface{}{"hits": []interface{}{
					map[string]interface{}{"_source": map[string]interface{}{
						"id":       "194d1060-1717-44dc-a783-00038f4a8013",
						"tenantID": "123456789012345678901234",
						model.ToAttr("inventory", "foo", model.TypeStr): []string{"bar"},
					}}},
					"total": map[string]interface{}{
						"value": float64(1),
					}},
				}, nil)
			return store
		},
		TotalCount: 1,
		Result: []model.InvDevice{{
			ID: "194d1060-1717-44dc-a783-00038f4a8013",
			Attributes: model.DeviceAttributes{{
				Name:  "foo",
				Value: []string{"bar"},
				Scope: "inventory",
			}},
		}},
	}, {
		Name: "ok, empty result",

		Params: &model.SearchParams{},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildQuery(*self.Params)
			store.On("Search", contextMatcher, q).
				Return(model.M{
					"hits": map[string]interface{}{
						"hits": []interface{}{},
						"total": map[string]interface{}{
							"value": float64(0),
						},
					},
				}, nil)
			return store
		},
		Result: []model.InvDevice{},
	}, {
		Name: "error, internal storage-layer error",

		Params: &model.SearchParams{},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildQuery(*self.Params)
			store.On("Search", contextMatcher, q).
				Return(nil, errors.New("internal error"))
			return store
		},
		Result: []model.InvDevice{},
		Error:  errors.New("internal error"),
	}, {
		Name: "error, parsing elastic result",

		Params: &model.SearchParams{},
		Store: func(t *testing.T, self testCase) *mstore.Store {
			store := new(mstore.Store)
			q, _ := model.BuildQuery(*self.Params)
			store.On("Search", contextMatcher, q).
				Return(model.M{
					"hits": map[string]interface{}{
						"hits": []interface{}{},
						"total": map[string]interface{}{
							"value": "doh!",
						},
					},
				}, nil)
			return store
		},
		Result: []model.InvDevice{},
		Error:  errors.New("can't process total hits value"),
	}, {
		Name: "error, invalid search parameters",

		Params: &model.SearchParams{
			Filters: []model.FilterPredicate{{
				Attribute: "foo",
				Value:     true,
				Scope:     "baz",
				Type:      "$useyourimagination",
			}},
		},
		Error: errors.New("filter type not supported"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			var store *mstore.Store
			if tc.Store == nil {
				store = new(mstore.Store)
			} else {
				store = tc.Store(t, tc)
			}
			defer store.AssertExpectations(t)

			app := NewApp(store, nil, nil)
			res, cnt, err := app.InventorySearchDevices(context.Background(), tc.Params)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.TotalCount, cnt)
				assert.Equal(t, tc.Result, res)
			}
		})
	}
}
