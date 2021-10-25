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
	"strings"
	"time"

	"github.com/panjf2000/ants/v2"

	"github.com/mendersoftware/go-lib-micro/log"

	"github.com/mendersoftware/reporting/client/inventory"
	"github.com/mendersoftware/reporting/model"
	"github.com/mendersoftware/reporting/store"
)

var (
	l = log.New(nil)

	ErrReindexChannelFull = errors.New("reindex input channel is full")
)

type reindexReq struct {
	Tenant   string
	Device   string
	Services []string
}

type Reindexer interface {
	Run() error
	Handle(r reindexReq) error
}

type reindexer struct {
	inChan    chan reindexReq
	store     store.Store
	inventory inventory.Client
	conf      *ReindexerConfig
}

type ReindexerConfig struct {
	NumWorkers  int
	BatchSize   int
	MaxTimeMsec int
	BuffLen     int
}

func NewReindexer(conf *ReindexerConfig, client inventory.Client, store store.Store) *reindexer {
	return &reindexer{
		inventory: client,
		store:     store,
		conf:      conf,
	}
}

func (ri *reindexer) Run() error {
	l.Debug("starting reindexer")
	c1 := buffer(ri.conf.BuffLen)
	ri.inChan = c1

	c2 := batch(c1, ri.conf.BatchSize, ri.conf.MaxTimeMsec)
	c3 := squash(c2)
	c4 := fetch(c3, ri.inventory, ri.store)
	c5 := merge_updates(c4)
	err := update(c5, ri.store, ri.conf.NumWorkers)
	return err
}

func (ri *reindexer) Handle(r reindexReq) error {
	l.Debug("reindexer.Handle")
	select {
	case ri.inChan <- r:
		l.Debugf("reindexer.Handle buffered request, chan len %v", len(ri.inChan))
		return nil
	default:
		return ErrReindexChannelFull
	}
}

// buffer simply creates the input buffer
func buffer(length int) chan reindexReq {
	l.Debug("spawning buffer() stage")
	out := make(chan reindexReq, length)

	return out
}

// batch groups incoming reindex requests into batches
func batch(inchan chan reindexReq, batchSize int, maxMsec int) chan []reindexReq {
	l.Debug("spawning batch() stage")
	out := make(chan []reindexReq)

	go func() {
		defer close(out)
		var batch []reindexReq
		tick := time.NewTicker(time.Millisecond * time.Duration(maxMsec))
		for {
			select {
			case r, ok := <-inchan:
				if ok {
					batch = append(batch, r)
					if len(batch) == batchSize {
						l.Debugf("counter, got batch: %v\n", batch)
						send := append([]reindexReq(nil), batch...)
						batch = nil
						out <- send
					}
				} else {
					break
				}
			case <-tick.C:
				l.Debugf("ticker, got batch: %v\n", batch)
				if len(batch) > 0 {
					send := append([]reindexReq(nil), batch...)
					out <- send
				}
				batch = nil
				tick.Reset(time.Second * 2)
			}
		}
	}()
	return out
}

// squash squashes reindex requests for a device from individual services into a single one
// it will save us some ES io at the final bulk update stage
func squash(inchan chan []reindexReq) chan []reindexReq {
	l.Debug("spawning squash() stage")
	out := make(chan []reindexReq)

	go func() {
		defer close(out)
		for batch := range inchan {
			l.Debugf("squash recv %v\n", batch)
			squashed := []reindexReq{}

			//map tid:did:services
			m := map[string][]string{}

			for _, req := range batch {
				k := req.Tenant + ":" + req.Device
				services, ok := m[k]

				if !ok {
					m[k] = append([]string{}, req.Services...)
				} else {
					found := false
					for _, s := range services {
						// we know there's just 1 service here for now
						if s == req.Services[0] {
							found = true
						}
					}
					if !found {
						m[k] = append(m[k], req.Services[0])
					}
				}
			}

			for k, v := range m {
				key := strings.Split(k, ":")
				squashed = append(squashed,
					reindexReq{
						Tenant:   key[0],
						Device:   key[1],
						Services: v})
			}

			out <- squashed
		}
	}()
	return out
}

// fetch pulls all the representations of a given device from service APIs within the reindexRequest
// for subsequent merging/update preparation
func fetch(inchan chan []reindexReq, client inventory.Client, store store.Store) chan []mergeJob {
	l.Debug("spawning fetch() stage")
	out := make(chan []mergeJob)

	go func() {
		for batch := range inchan {
			l.Debugf("fetch recv %v\n", batch)

			// output merge jobs, organized for easier access
			// as tenant:device:job maps
			jobs := map[string]map[string]mergeJob{}

			// inventory, and services in general have to be queried
			// tenant by tenant,so do that
			// (most popular convention in our APIs)
			tenantDevs := map[string][]string{}

			for _, r := range batch {
				j := mergeJob{
					Tenant: r.Tenant,
					Device: r.Device,
					// we know we can only have inventory for now
					// later, find out which sources asked for reindex
					SrcInventory: &mergeSrcInventory{},
					SrcElastic:   &mergeSrcElastic{},
				}

				// preinit output jobs
				if _, ok := jobs[r.Tenant]; !ok {
					jobs[r.Tenant] = map[string]mergeJob{}
				}

				jobs[r.Tenant][r.Device] = j

				// and prep the per tenant device lookup
				if devs, ok := tenantDevs[r.Tenant]; !ok {
					tenantDevs[r.Tenant] = []string{r.Device}
				} else {
					tenantDevs[r.Tenant] = append(devs, r.Device)
				}
			}

			// TODO async scatter/gather?
			for tenant, devs := range tenantDevs {
				invDevs, err := client.GetDevices(context.TODO(), tenant, devs)
				if err != nil {
					l.Debugf("fetch inventory error %v for devs %v",
						err,
						tenantDevs)
					continue
				} else {
					l.Debugf("fetch inventory got devs %v \n", invDevs)
					for _, d := range invDevs {
						dev := d
						jobs[tenant][string(d.ID)].SrcInventory.device =
							&dev
					}
				}
			}

			// through elastic multiget request, all devs across all tenants
			// can be pulled in one go
			esDevs, err := store.GetDevices(context.TODO(), tenantDevs)
			if err != nil {
				l.Debugf("fetch elastic error %v for devs %v", err, esDevs)
				continue
			} else {
				l.Debugf("fetch elastic got devs %+#v \n", esDevs)
			}

			for _, d := range esDevs {
				jobs[*d.TenantID][*d.ID].SrcElastic.device = &d
			}

			// flatten the list of merge jobs
			retJobs := []mergeJob{}
			for _, dev := range jobs {
				for _, job := range dev {
					retJobs = append(retJobs, job)
				}
			}

			out <- retJobs
		}
	}()

	return out
}

// mergeJob aggregates all the fetched representations of a device
// (inventory API + other service APIs + ES)
// if a representation is null - service didn't ask for an update
type mergeJob struct {
	Tenant       string
	Device       string
	SrcInventory *mergeSrcInventory
	SrcElastic   *mergeSrcElastic
}

type mergeSrcInventory struct {
	device *model.InvDevice
}

type mergeSrcElastic struct {
	device *model.Device
}

// later: more sources/services
// type mergeSrcMonitoring struct {
// 	device *model.MonitoringDevice
// }, etc

// merge_updates merges all the available service representations of a device into one final update
// suitable for writing to es
func merge_updates(inchan chan []mergeJob) chan []store.BulkItem {
	l.Debug("spawning merge_updates() stage")

	out := make(chan []store.BulkItem)
	go func() {
		for batch := range inchan {
			l.Debugf("merge_updates recv %v\n", batch)

			var bulkItems []store.BulkItem
			for _, job := range batch {
				item, _ := merge(&job)
				bulkItems = append(bulkItems, *item)
			}

			out <- bulkItems
		}
	}()
	return out
}

// merge merges all the update sources into an update object
// for now it's just inventory
func merge(j *mergeJob) (*store.BulkItem, error) {
	now := time.Now()

	action := &store.BulkAction{
		Desc: &store.BulkActionDesc{
			Tenant: j.Tenant,
			ID:     j.Device,
		},
	}

	item := &store.BulkItem{
		Action: action,
	}

	switch {
	case j.SrcInventory.device == nil:
		item.Action.Type = "delete"

		if j.SrcElastic.device != nil {
			// concurrency control
			item.Action.Desc.IfSeqNo = j.SrcElastic.device.Meta.SeqNo
			item.Action.Desc.IfPrimaryTerm = j.SrcElastic.device.Meta.PrimaryTerm
		}
	case j.SrcElastic.device == nil:
		newdev, _ := model.NewDeviceFromInv(j.Tenant, j.SrcInventory.device)

		newdev.SetCreatedAt(now)
		newdev.SetUpdatedAt(now)
		item.Doc = newdev
		item.Action.Type = "create"

	default:
		newdev, _ := model.NewDeviceFromInv(j.Tenant, j.SrcInventory.device)

		newdev.SetUpdatedAt(now)

		item.Doc = newdev
		item.Action.Type = "index"

		// concurrency control
		item.Action.Desc.IfSeqNo = j.SrcElastic.device.Meta.SeqNo
		item.Action.Desc.IfPrimaryTerm = j.SrcElastic.device.Meta.PrimaryTerm
	}

	return item, nil
}

// bulk executes bulk update jobs for a device batch
func update(inchan chan []store.BulkItem, store store.Store, numWorkers int) error {
	l.Debug("spawning update() stage")

	p, err := ants.NewPool(numWorkers)
	if err != nil {
		return err
	}

	go func() {
		for bulkItems := range inchan {
			l.Debugf("update recv %v\n", bulkItems)

			err := p.Submit(func() {
				res, err := store.BulkRaw(context.TODO(), bulkItems)
				if err != nil {
					l.Errorf("BulkRaw failed for bulkItems %v with error %v",
						bulkItems,
						err)
				}

				l.Debugf("bulk response %v", res)

				// inspect the bulk response and at least emit warnings
				// (future: requeue conflicting devices?)
				handleBulkResponse(res)
			})
			if err != nil {
				l.Errorf("failed to submit bulk update to pool %v\n", bulkItems)
			}
		}
	}()

	return nil
}

func handleBulkResponse(res map[string]interface{}) {
	hasErrs := res["errors"].(bool)
	l.Debugf("bulk response hasErrs %v", hasErrs)

	if hasErrs {
		items := res["items"].([]interface{})

		// FIXME: steal the struct def from esapi.BulkIndexer
		// or write our own
		for _, item := range items {
			action := item.(map[string]interface{})

			for _, v := range action {
				valM := v.(map[string]interface{})

				for kk, vv := range valM {
					var id, idx string

					if kk == "_id" {
						id = vv.(string)
					}
					if kk == "_index" {
						idx = vv.(string)
					}

					if kk == "error" {
						l.Warnf("bulk update failed for dev %v:%v, %v\n",
							id,
							idx,
							valM)
					}
				}
			}
		}
	}
}
