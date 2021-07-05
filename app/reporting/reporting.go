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

	"github.com/mendersoftware/reporting/model"
	"github.com/mendersoftware/reporting/store"
)

type App interface {
	InventorySearchDevices(ctx context.Context, searchParams *model.SearchParams) (interface{}, int, error)
}

type app struct {
	store store.Store
}

func NewApp(store store.Store) App {
	return &app{
		store: store,
	}
}

func (app *app) InventorySearchDevices(ctx context.Context, searchParams *model.SearchParams) (interface{}, int, error) {
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
func (a *app) storeToInventoryDevs(storeRes map[string]interface{}) ([]model.InvDevice, int, error) {
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
			return nil, errors.New("can't parse device id as neither single value nor array")
		}

		id, ok = idarr[0].(string)
		if !ok {
			return nil, errors.New("can't parse device id as neither single value nor array")
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
				Name:  n,
				Scope: s,
				Value: v,
			}

			attrs = append(attrs, a)
		}
	}

	ret.Attributes = attrs

	return ret, nil
}
