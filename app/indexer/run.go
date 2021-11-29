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
	"os"
	"os/signal"
	"time"

	"github.com/mendersoftware/go-lib-micro/config"
	"golang.org/x/sys/unix"

	"github.com/mendersoftware/reporting/client/deviceauth"
	"github.com/mendersoftware/reporting/client/inventory"
	"github.com/mendersoftware/reporting/client/nats"
	rconfig "github.com/mendersoftware/reporting/config"
	"github.com/mendersoftware/reporting/model"
	"github.com/mendersoftware/reporting/store"
)

const jobsChanSize = 1000

// InitAndRun initializes the indexer and runs it
func InitAndRun(conf config.Reader, store store.Store, nats nats.Client) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	invClient := inventory.NewClient(
		conf.GetString(rconfig.SettingInventoryAddr),
	)

	devClient := deviceauth.NewClient(
		conf.GetString(rconfig.SettingDeviceAuthAddr),
	)

	indexer := NewIndexer(store, nats, devClient, invClient)
	jobs := make(chan *model.Job, jobsChanSize)

	err := indexer.GetJobs(ctx, jobs)
	if err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, unix.SIGINT, unix.SIGTERM)

	batchSize := conf.GetInt(rconfig.SettingReindexBatchSize)
	jobsList := make([]*model.Job, batchSize)
	jobsListSize := 0

	maxTimeMs := conf.GetInt(rconfig.SettingReindexMaxTimeMsec)
	tickerTimeout := time.Duration(maxTimeMs) * time.Millisecond
	ticker := time.NewTimer(tickerTimeout)

	for {
		select {
		case <-ticker.C:
			if jobsListSize > 0 {
				indexer.ProcessJobs(ctx, jobsList[0:jobsListSize])
				jobsListSize = 0
			}
			ticker.Reset(tickerTimeout)

		case job := <-jobs:
			jobsList[jobsListSize] = job
			jobsListSize++
			if jobsListSize == batchSize {
				indexer.ProcessJobs(ctx, jobsList[0:jobsListSize])
				ticker.Reset(tickerTimeout)
				jobsListSize = 0
			}

		case <-quit:
			return nil
		}
	}
}
