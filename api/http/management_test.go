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

package http

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mendersoftware/go-lib-micro/rest.utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/go-lib-micro/identity"
	mapp "github.com/mendersoftware/reporting/app/reporting/mocks"
	"github.com/mendersoftware/reporting/model"
)

func GenerateJWT(id identity.Identity) string {
	JWT := base64.RawURLEncoding.EncodeToString(
		[]byte(`{"alg":"HS256","typ":"JWT"}`),
	)
	b, _ := json.Marshal(id)
	JWT = JWT + "." + base64.RawURLEncoding.EncodeToString(b)
	hash := hmac.New(sha256.New, []byte("hmac-sha256-secret"))
	JWT = JWT + "." + base64.RawURLEncoding.EncodeToString(
		hash.Sum([]byte(JWT)),
	)
	return JWT
}

func TestManagementSearch(t *testing.T) {
	t.Parallel()
	var newSearchParamMatcher = func(expected *model.SearchParams) interface{} {
		return mock.MatchedBy(func(actual *model.SearchParams) bool {
			if expected.Page < 0 {
				expected.Page = ParamPageDefault
			}
			if expected.PerPage < 0 {
				expected.PerPage = ParamPerPageDefault
			}
			if assert.NotNil(t, actual) {
				return assert.Equal(t, *expected, *actual)
			}
			return false
		})
	}
	type testCase struct {
		Name string

		App    func(*testing.T, testCase) *mapp.App
		CTX    context.Context
		Params interface{} // *model.SearchParams

		Code     int
		Response interface{}
	}
	testCases := []testCase{{
		Name: "ok",

		App: func(t *testing.T, self testCase) *mapp.App {
			app := new(mapp.App)

			app.On("InventorySearchDevices",
				contextMatcher,
				newSearchParamMatcher(self.Params.(*model.SearchParams))).
				Return(self.Response, 0, nil)
			return app
		},
		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: &model.SearchParams{
			PerPage: 10,
			Page:    2,
			Filters: []model.FilterPredicate{{
				Scope:     "inventory",
				Attribute: "ip4",
				Type:      "$exists",
				Value:     true,
			}},
			Sort: []model.SortCriteria{{
				Scope:     "inventory",
				Attribute: "ip4",
				Order:     "asc",
			}},
		},

		Code: http.StatusOK,
		Response: []model.InvDevice{{
			ID: model.DeviceID("5975e1e6-49a6-4218-a46d-f181154a98cc"),
			Attributes: model.DeviceAttributes{{
				Scope: "inventory",
				Name:  "ip4",
				Value: "10.0.0.2",
			}, {
				Scope: "system",
				Name:  "group",
				Value: "develop",
			}},
			Group:     model.GroupName("dev-set"),
			CreatedTs: time.Now().Add(-time.Hour),
			UpdatedTs: time.Now().Add(-time.Minute),
			Revision:  3,
		}, {
			ID: model.DeviceID("83bce0e4-c4c0-4995-b8b7-f056da7fc8f6"),

			Attributes: model.DeviceAttributes{{
				Scope: "inventory",
				Name:  "ip4",
				Value: "10.0.0.2",
			}, {
				Scope: "system",
				Name:  "group",
				Value: "prod_horse",
			}},
			Group:     model.GroupName("prod_horse"),
			CreatedTs: time.Now().Add(-2 * time.Hour),
			UpdatedTs: time.Now().Add(-5 * time.Minute),
			Revision:  120,
		}},
	}, {
		Name: "ok, empty result",

		App: func(t *testing.T, self testCase) *mapp.App {
			app := new(mapp.App)

			app.On("InventorySearchDevices",
				contextMatcher,
				newSearchParamMatcher(self.Params.(*model.SearchParams))).
				Return([]model.InvDevice{}, 0, nil)
			return app
		},
		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: &model.SearchParams{},

		Code:     http.StatusOK,
		Response: []model.InvDevice{},
	}, {
		Name: "error, malformed request body",

		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: &model.SearchParams{
			Filters: []model.FilterPredicate{{
				Scope:     "secret-attrs",
				Type:      "$maybethiswillfindsomethinginterresting",
				Attribute: "rootpwd",
				Value:     true,
			}},
		},
		Code:     http.StatusBadRequest,
		Response: rest.Error{Err: "malformed request body: type: must be a valid value."},
	}, {
		Name: "error, internal app error",

		App: func(t *testing.T, self testCase) *mapp.App {
			app := new(mapp.App)

			app.On("InventorySearchDevices",
				contextMatcher,
				newSearchParamMatcher(self.Params.(*model.SearchParams))).
				Return(nil, 0, errors.New("internal error"))
			return app
		},
		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: &model.SearchParams{
			PerPage: 10,
			Page:    2,
			Filters: []model.FilterPredicate{{
				Scope:     "inventory",
				Attribute: "ip4",
				Type:      "$exists",
				Value:     true,
			}},
			Sort: []model.SortCriteria{{
				Scope:     "inventory",
				Attribute: "ip4",
				Order:     "asc",
			}},
		},

		Code:     http.StatusInternalServerError,
		Response: rest.Error{Err: "internal error"},
	}, {
		Name: "error, tenant ID not present",

		App: func(t *testing.T, self testCase) *mapp.App {
			return new(mapp.App)
		},
		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
			},
		),
		Params: &model.SearchParams{},

		Code:     http.StatusUnauthorized,
		Response: rest.Error{Err: "tenant claim not present in JWT"},
	}, {
		Name: "error, tenant ID not present",

		App: func(t *testing.T, self testCase) *mapp.App {
			return new(mapp.App)
		},
		CTX: identity.WithContext(context.Background(),
			&identity.Identity{
				Subject: "851f90b3-cee5-425e-8f6e-b36de1993e7e",
				Tenant:  "123456789012345678901234",
			},
		),
		Params: map[string]string{
			"filters": "foo",
		},

		Code: http.StatusBadRequest,
		Response: rest.Error{
			Err: "malformed request body: json: " +
				"cannot unmarshal string into Go struct field " +
				"SearchParams.filters of type []model.FilterPredicate",
		},
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var app *mapp.App
			if tc.App == nil {
				app = new(mapp.App)
			} else {
				app = tc.App(t, tc)
			}
			defer app.AssertExpectations(t)
			router := NewRouter(app)

			b, _ := json.Marshal(tc.Params)
			req, _ := http.NewRequest(
				http.MethodPost,
				URIManagement+URIInventorySearch,
				bytes.NewReader(b),
			)
			if id := identity.FromContext(tc.CTX); id != nil {
				req.Header.Set("Authorization", "Bearer "+GenerateJWT(*id))
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.Code, w.Code)

			switch res := tc.Response.(type) {
			case []model.InvDevice:
				b, _ := json.Marshal(res)
				assert.JSONEq(t, string(b), w.Body.String())

			case rest.Error:
				var actual rest.Error
				dec := json.NewDecoder(w.Body)
				dec.DisallowUnknownFields()
				err := dec.Decode(&actual)
				if assert.NoError(t, err, "response schema did not match expected rest.Error") {
					assert.EqualError(t, res, actual.Error())
				}

			case nil:
				assert.Empty(t, w.Body.String())

			default:
				panic("[TEST ERR] Dunno what to compare!")
			}

		})
	}
}
