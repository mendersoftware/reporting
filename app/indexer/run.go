// Copyright 2023 Northern.tech AS
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
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"golang.org/x/sys/unix"

	"github.com/mendersoftware/go-lib-micro/config"
	"github.com/mendersoftware/go-lib-micro/log"

	"github.com/mendersoftware/reporting/client/deployments"
	"github.com/mendersoftware/reporting/client/deviceauth"
	"github.com/mendersoftware/reporting/client/inventory"
	"github.com/mendersoftware/reporting/client/nats"
	rconfig "github.com/mendersoftware/reporting/config"
	"github.com/mendersoftware/reporting/model"
	"github.com/mendersoftware/reporting/store"
)

const (
	jobsChanSize = 1000
)

// InitAndRun initializes the indexer and runs it
func InitAndRun(conf config.Reader, store store.Store, ds store.DataStore, nats nats.Client) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	invClient := inventory.NewClient(
		conf.GetString(rconfig.SettingInventoryAddr),
	)

	devClient := deviceauth.NewClient(
		conf.GetString(rconfig.SettingDeviceAuthAddr),
	)

	deplClient := deployments.NewClient(
		conf.GetString(rconfig.SettingDeploymentsAddr),
	)

	indexer := NewIndexer(store, ds, nats, devClient, invClient, deplClient)
	jobs := make(chan model.Job, jobsChanSize)

	err := indexer.GetJobs(ctx, jobs)
	if err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, unix.SIGINT, unix.SIGTERM)
	go func() {
		select {
		case <-quit:
			cancel()
		case <-ctx.Done():
		}
	}()

	batchSize := conf.GetInt(rconfig.SettingReindexBatchSize)
	if batchSize <= 0 {
		return fmt.Errorf(
			"%s: must be a positive integer",
			rconfig.SettingReindexBatchSize,
		)
	}
	workerConcurrency := conf.GetInt(rconfig.SettingWorkerConcurrency)
	if workerConcurrency <= 0 {
		return fmt.Errorf(
			"%s: must be a positive integer",
			rconfig.SettingWorkerConcurrency,
		)
	}
	dispatch := make(chan []model.Job)
	jobPool := make(chan []model.Job, workerConcurrency)
	for i := 0; i < workerConcurrency; i++ {
		jobPool <- make([]model.Job, batchSize)
		go workerRoutine(ctx, strconv.Itoa(i+1), indexer, dispatch, jobPool)
	}

	maxTimeMs := conf.GetInt(rconfig.SettingReindexMaxTimeMsec)
	tickerTimeout := time.Duration(maxTimeMs) * time.Millisecond
	ticker := time.NewTimer(tickerTimeout)
	jobsList := <-jobPool
	done := ctx.Done()
	for err == nil {
		select {
		case <-ticker.C:
			ticker.Reset(tickerTimeout)
			if len(jobsList) > 0 {
				jobsList, err = dispatchJobs(ctx, jobsList, dispatch, jobPool)
			}

		case job, open := <-jobs:
			if !open {
				return errors.New("Jetstream closed")
			}
			jobsList = append(jobsList, job)
			if len(jobsList) >= cap(jobsList) {
				ticker.Reset(tickerTimeout)
				jobsList, err = dispatchJobs(ctx, jobsList, dispatch, jobPool)
			}

		case <-done:
			err = ctx.Err()
		}
	}
	return err
}

func dispatchJobs(ctx context.Context,
	jobs []model.Job,
	dispatch chan<- []model.Job,
	jobPool <-chan []model.Job,
) (next []model.Job, err error) {
	done := ctx.Done()
	select {
	case <-done:
		return nil, ctx.Err()
	case dispatch <- jobs:
	}
	select {
	case <-done:
		return nil, ctx.Err()
	case next = <-jobPool:
	}
	return next[:0], nil
}

func workerRoutine(
	ctx context.Context,
	workerName string,
	indexer Indexer,
	jobQ <-chan []model.Job,
	jobPool chan<- []model.Job) {
	l := log.FromContext(ctx)
	l.Data["worker"] = workerName
	l.Infof("Worker %s waiting for jobs", workerName)
	ctx = log.WithContext(ctx, l)
	for jobs := range jobQ {
		l.Infof("processing %d jobs", len(jobs))
		indexer.ProcessJobs(ctx, jobs)
		jobPool <- jobs
	}
}
