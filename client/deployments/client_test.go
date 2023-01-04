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

package deployments

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/go-lib-micro/rest.utils"
)

func newTestServer(
	rspChan <-chan *http.Response,
	reqChan chan<- *http.Request,
) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var rsp *http.Response
		select {
		case rsp = <-rspChan:
		default:
			panic("[PROG ERR] I don't know what to respond!")
		}
		if reqChan != nil {
			bodyClone := bytes.NewBuffer(nil)
			_, _ = io.Copy(bodyClone, r.Body)
			req := r.Clone(context.TODO())
			req.Body = io.NopCloser(bodyClone)
			select {
			case reqChan <- req:
				// Only push request if test function is
				// popping from the channel.
			default:
			}
		}
		hdrs := w.Header()
		for k, v := range rsp.Header {
			for _, vv := range v {
				hdrs.Add(k, vv)
			}
		}
		w.WriteHeader(rsp.StatusCode)
		if rsp.Body != nil {
			_, _ = io.Copy(w, rsp.Body)
		}
	}
	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestGetDevices(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		CTX      context.Context
		TenantID string
		DeviceID string

		URLNoise     string
		ResponseCode int
		ResponseBody interface{}

		Res   *DeviceDeployment
		Error error
	}{{
		Name: "ok, no devices",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",
		DeviceID: "9acfe595-78ff-456a-843a-0fa08bfd7c7a",

		ResponseCode: http.StatusOK,
		ResponseBody: []DeviceDeployment{},
	}, {
		Name: "ok",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",
		DeviceID: "9acfe595-78ff-456a-843a-0fa08bfd7c7a",

		ResponseCode: http.StatusOK,
		ResponseBody: []DeviceDeployment{{
			ID: "c5e37ef5-160e-401a-aec3-9dbef94855c0",
			Device: &Device{
				Status: "success",
			},
		}},

		Res: &DeviceDeployment{
			ID: "c5e37ef5-160e-401a-aec3-9dbef94855c0",
			Device: &Device{
				Status: "success",
			},
		},
	}, {
		Name: "ok, not found",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",
		DeviceID: "9acfe595-78ff-456a-843a-0fa08bfd7c7a",

		ResponseCode: http.StatusNotFound,
	}, {
		Name: "error, context canceled",

		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(),
		Error: context.Canceled,
	}, {
		Name:     "error, nil context",
		CTX:      context.Background(),
		URLNoise: "#%%%",

		Error: errors.New("failed to create request"),
	}, {
		Name: "error, invalid response schema",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",
		DeviceID: "9acfe595-78ff-456a-843a-0fa08bfd7c7a",

		ResponseCode: http.StatusOK,
		ResponseBody: []byte("bad response"),
		Error:        errors.New("failed to parse request body"),
	}, {
		Name: "error, unexpected status code",

		CTX:      context.Background(),
		TenantID: "123456789012345678901234",
		DeviceID: "9acfe595-78ff-456a-843a-0fa08bfd7c7a",

		ResponseCode: http.StatusInternalServerError,
		ResponseBody: rest.Error{Err: "something went wrong..."},
		Error:        errors.New(`^GET .+ request failed with status 500`),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			rspChan := make(chan *http.Response, 1)
			srv := newTestServer(rspChan, nil)
			defer srv.Close()

			client := NewClient(srv.URL + tc.URLNoise)

			rsp := &http.Response{
				StatusCode: tc.ResponseCode,
			}

			switch typ := tc.ResponseBody.(type) {
			case []DeviceDeployment:
				b, _ := json.Marshal(typ)
				rsp.Body = io.NopCloser(bytes.NewReader(b))

			case rest.Error:
				b, _ := json.Marshal(typ)
				rsp.Body = io.NopCloser(bytes.NewReader(b))

			case []byte:
				rsp.Body = io.NopCloser(bytes.NewReader(typ))

			case nil:
				// pass

			default:
				panic("[PROG ERR] invalid ResponseBody type")
			}
			rspChan <- rsp
			dev, err := client.GetLatestFinishedDeployment(tc.CTX, tc.TenantID, tc.DeviceID)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"error message does not match expected pattern",
					)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Res, dev)
			}

		})
	}
}
