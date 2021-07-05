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

package indexer

import (
	"context"

	"github.com/mendersoftware/go-lib-micro/config"
	"github.com/mendersoftware/reporting/model"
	"github.com/mendersoftware/reporting/store"
)

const batchSize = 200

// InitAndRun initializes the indexer and runs it
func InitAndRun(conf config.Reader, store store.Store, devices int64, tid string) error {
	ctx := context.Background()

	devicesToIndex := make([]*model.Device, 0, batchSize)

	for i := int64(1); i <= devices; i++ {
		device := model.RandomDevice(tid)
		devicesToIndex = append(devicesToIndex, device)
		if len(devicesToIndex) == batchSize {
			err := store.BulkIndexDevices(ctx, devicesToIndex)
			if err != nil {
				return err
			}
			devicesToIndex = devicesToIndex[:0]
		}
	}
	if len(devicesToIndex) > 0 {
		err := store.BulkIndexDevices(ctx, devicesToIndex)
		if err != nil {
			return err
		}
	}
	return nil
}
