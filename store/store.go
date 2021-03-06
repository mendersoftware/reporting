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

package store

import (
	"context"

	"github.com/mendersoftware/reporting/model"
)

//go:generate ../x/mockgen.sh
type Store interface {
	BulkIndexDevices(ctx context.Context, devices, removedDevices []*model.Device) error
	GetDevicesIndex(tid string) string
	GetDevicesRoutingKey(tid string) string
	GetDevicesIndexMapping(ctx context.Context, tid string) (map[string]interface{}, error)
	Migrate(ctx context.Context) error
	Search(ctx context.Context, query interface{}) (model.M, error)
}
