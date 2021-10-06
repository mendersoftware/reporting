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
	"sort"

	"github.com/mendersoftware/go-lib-micro/log"

	"github.com/mendersoftware/reporting/client/inventory"
	"github.com/mendersoftware/reporting/model"
	"github.com/mendersoftware/reporting/store"
)

const (
	SvcInventory  = "inventory"
	SvcDeviceauth = "deviceauth"
)

var (
	knownServices = []string{SvcInventory, SvcDeviceauth}

	ErrUnknownService = errors.New("unknown service name")
)

//nolint:lll
//go:generate ../../x/mockgen.sh
type App interface {
	GetSearchableInvAttrs(ctx context.Context, tid string) ([]model.InvFilterAttr, error)
	InventorySearchDevices(ctx context.Context, searchParams *model.SearchParams) ([]model.InvDevice, int, error)
	Reindex(ctx context.Context, tenantID, devID string, service string) error
}

type app struct {
	store     store.Store
	invClient inventory.Client
	reindexer Reindexer
}

func NewApp(store store.Store, client inventory.Client, ri Reindexer) App {
	return &app{
		store:     store,
		invClient: client,
		reindexer: ri,
	}
}

func (app *app) InventorySearchDevices(
	ctx context.Context,
	searchParams *model.SearchParams,
) ([]model.InvDevice, int, error) {
	query, err := model.BuildQuery(*searchParams)
	if err != nil {
		return nil, 0, err
	}

	if len(searchParams.DeviceIDs) > 0 {
		query = query.Must(model.M{
			"terms": model.M{
				"id": searchParams.DeviceIDs,
			},
		})
	}

	esRes, err := app.store.Search(ctx, query)

	if err != nil {
		return nil, 0, err
	}

	res, total, err := app.storeToInventoryDevs(esRes)
	if err != nil {
		return nil, 0, err
	}

	return res, total, err
}

// storeToInventoryDevs translates ES results directly to iventory devices
func (a *app) storeToInventoryDevs(
	storeRes map[string]interface{},
) ([]model.InvDevice, int, error) {
	devs := []model.InvDevice{}

	hitsM, ok := storeRes["hits"].(map[string]interface{})
	if !ok {
		return nil, 0, errors.New("can't process store hits map")
	}

	hitsTotalM, ok := hitsM["total"].(map[string]interface{})
	if !ok {
		return nil, 0, errors.New("can't process total hits struct")
	}

	total, ok := hitsTotalM["value"].(float64)
	if !ok {
		return nil, 0, errors.New("can't process total hits value")
	}

	hitsS, ok := hitsM["hits"].([]interface{})
	if !ok {
		return nil, 0, errors.New("can't process store hits slice")
	}

	for _, v := range hitsS {
		res, err := a.storeToInventoryDev(v)
		if err != nil {
			return nil, 0, err
		}

		devs = append(devs, *res)
	}

	return devs, int(total), nil
}

func (a *app) storeToInventoryDev(storeRes interface{}) (*model.InvDevice, error) {
	resM, ok := storeRes.(map[string]interface{})
	if !ok {
		return nil, errors.New("can't process individual hit")
	}

	// if query has a 'fields' clause, use 'fields' instead of '_source'
	sourceM, ok := resM["_source"].(map[string]interface{})
	if !ok {
		sourceM, ok = resM["fields"].(map[string]interface{})
		if !ok {
			return nil, errors.New("can't process hit's '_source' nor 'fields'")
		}
	}

	// if query has a 'fields' clause, all results will be arrays incl. device id, so extract it
	id, ok := sourceM["id"].(string)
	if !ok {
		idarr, ok := sourceM["id"].([]interface{})
		if !ok {
			return nil, errors.New(
				"can't parse device id as neither single value nor array",
			)
		}

		id, ok = idarr[0].(string)
		if !ok {
			return nil, errors.New(
				"can't parse device id as neither single value nor array",
			)
		}
	}

	ret := &model.InvDevice{
		ID: model.DeviceID(id),
	}

	attrs := []model.InvDeviceAttribute{}

	for k, v := range sourceM {
		s, n, err := model.MaybeParseAttr(k)

		if err != nil {
			return nil, err
		}

		if n != "" {
			a := model.InvDeviceAttribute{
				Name:  model.Redot(n),
				Scope: s,
				Value: v,
			}

			attrs = append(attrs, a)
		}
	}

	ret.Attributes = attrs

	return ret, nil
}

func (app *app) Reindex(ctx context.Context, tenantID, devID string, service string) error {
	l := log.FromContext(ctx)
	l.Debugf("triggered reindexing for device %v:%v", tenantID, devID)

	known := false
	for _, s := range knownServices {
		if service == s {
			known = true
		}
	}
	if !known {
		return ErrUnknownService
	}

	err := app.reindexer.Handle(
		reindexReq{
			Tenant:   tenantID,
			Device:   devID,
			Services: []string{service}})

	return err
}

func (app *app) GetSearchableInvAttrs(
	ctx context.Context,
	tid string,
) ([]model.InvFilterAttr, error) {
	l := log.FromContext(ctx)

	index, err := app.store.GetDevIndex(ctx, tid)
	if err != nil {
		return nil, err
	}

	// inventory attributes are under 'mappings.properties'
	mappings, ok := index["mappings"]
	if !ok {
		return nil, errors.New("can't parse index mappings")
	}

	mappingsM, ok := mappings.(map[string]interface{})
	if !ok {
		return nil, errors.New("can't parse index mappings")
	}

	props, ok := mappingsM["properties"]
	if !ok {
		return nil, errors.New("can't parse index properties")
	}

	propsM, ok := props.(map[string]interface{})
	if !ok {
		return nil, errors.New("can't parse index properties")
	}

	ret := []model.InvFilterAttr{}

	for k := range propsM {
		s, n, err := model.MaybeParseAttr(k)

		if err != nil {
			return nil, err
		}

		if n != "" {
			ret = append(ret, model.InvFilterAttr{Name: n, Scope: s, Count: 1})
		}
	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[j].Scope > ret[i].Scope {
			return true
		}

		if ret[j].Scope < ret[i].Scope {
			return false
		}

		return ret[j].Name > ret[i].Name
	})

	l.Debugf("parsed searchable attributes %v\n", ret)

	return ret, nil
}
