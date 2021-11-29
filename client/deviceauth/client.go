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
package deviceauth

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mendersoftware/go-lib-micro/log"

	"github.com/mendersoftware/reporting/model"
	"github.com/mendersoftware/reporting/utils"
)

const (
	urlSearch      = "/api/internal/v1/devauth/tenants/:tid/devices"
	defaultTimeout = 10 * time.Second
)

//go:generate ../../x/mockgen.sh
type Client interface {
	//GetDevices uses the search endpoint to get devices just by ids (not filters)
	GetDevices(ctx context.Context, tid string, deviceIDs []string) ([]DeviceAuthDevice, error)
}

type client struct {
	client  *http.Client
	urlBase string
}

func NewClient(urlBase string) Client {
	return &client{
		client:  &http.Client{},
		urlBase: urlBase,
	}
}

func (c *client) GetDevices(
	ctx context.Context,
	tid string,
	deviceIDs []string,
) ([]DeviceAuthDevice, error) {
	l := log.FromContext(ctx)

	url := utils.JoinURL(c.urlBase, urlSearch)
	url = strings.Replace(url, ":tid", tid, 1)

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create request")
	}

	q := req.URL.Query()
	for _, deviceID := range deviceIDs {
		q.Add(model.AttrNameID, deviceID)
	}
	q.Add("page", "1")
	q.Add("per_page", strconv.Itoa(len(deviceIDs)))
	req.URL.RawQuery = q.Encode()

	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to submit %s %s", req.Method, req.URL)
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		err := errors.Errorf("%s %s request failed with status %v",
			req.Method, req.URL, rsp.Status)
		l.Errorf(err.Error())
		return nil, err
	}

	dec := json.NewDecoder(rsp.Body)
	var devDevs []DeviceAuthDevice
	if err = dec.Decode(&devDevs); err != nil {
		return nil, errors.Wrap(err, "failed to parse request body")
	}

	return devDevs, nil
}
