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

package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/log"
	_ "github.com/mendersoftware/go-lib-micro/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/mendersoftware/reporting/model"
)

//go:generate ../x/mockgen.sh
type Store interface {
	IndexDevice(ctx context.Context, device *model.Device) error
	BulkIndexDevices(ctx context.Context, devices []*model.Device) error
	BulkRaw(ctx context.Context, items []BulkItem) (map[string]interface{}, error)

	Search(ctx context.Context, query interface{}) (model.M, error)
	GetDevice(ctx context.Context, tenant, devid string) (*model.Device, error)
	GetDevices(ctx context.Context, tenantDevs map[string][]string) ([]model.Device, error)
	UpdateDevice(ctx context.Context, tenantID, deviceID string, updateDev *model.Device) error
	Migrate(ctx context.Context) error
	GetDevIndex(ctx context.Context, tid string) (map[string]interface{}, error)
}

type StoreOption func(*store)

type store struct {
	addresses []string
	client    *es.Client
}

func NewStore(opts ...StoreOption) (Store, error) {
	store := &store{}
	for _, opt := range opts {
		opt(store)
	}

	cfg := es.Config{
		Addresses: store.addresses,
	}
	esClient, err := es.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "invalid Elasticsearch configuration")
	}

	_, err = esClient.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to Elasticsearch")
	}

	store.client = esClient
	return store, nil
}

func (s *store) IndexDevice(ctx context.Context, device *model.Device) error {
	req := esapi.IndexRequest{
		Index:      devIdx(device.GetTenantID()),
		DocumentID: device.GetID(),
		Body:       esutil.NewJSONReader(device),
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "failed to index")
	}
	defer res.Body.Close()

	return nil
}

type BulkAction struct {
	Type string
	Desc *BulkActionDesc
}

type BulkActionDesc struct {
	ID            string `json:"_id"`
	Index         string `json:"_index"`
	IfSeqNo       int64  `json:"_if_seq_no"`
	IfPrimaryTerm int64  `json:"_if_primary_term"`
	Tenant        string
}

type BulkItem struct {
	Action *BulkAction
	Doc    interface{}
}

func (bad BulkActionDesc) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID    string `json:"_id"`
		Index string `json:"_index"`
	}{
		ID:    bad.ID,
		Index: devIdx(bad.Tenant),
	})
}

func (ba BulkAction) MarshalJSON() ([]byte, error) {
	a := map[string]*BulkActionDesc{
		ba.Type: ba.Desc,
	}
	return json.Marshal(a)
}

func (bi BulkItem) Marshal() ([]byte, error) {
	action, err := json.Marshal(bi.Action)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(action)
	buf.WriteString("\n")

	if bi.Doc == nil {
		return buf.Bytes(), nil
	}

	if bi.Doc != nil {
		doc, err := json.Marshal(bi.Doc)
		if err != nil {
			return nil, err
		}
		buf.Write(doc)
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

func (s *store) BulkRaw(ctx context.Context, items []BulkItem) (map[string]interface{}, error) {
	l := log.FromContext(ctx)

	var buf *bytes.Buffer
	for _, bi := range items {
		b, err := bi.Marshal()
		if err != nil {
			return nil, err
		}

		if buf == nil {
			buf = bytes.NewBuffer(b)
		}

		buf.Write(b)
	}

	req := esapi.BulkRequest{
		Body: buf,
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to bulk index")
	}
	defer res.Body.Close()

	var storeRes map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&storeRes); err != nil {
		return nil, err
	}

	l.Debugf("bulk response %v\n", storeRes)

	return storeRes, nil
}

func (s *store) BulkIndexDevices(ctx context.Context, devices []*model.Device) error {
	data := ""
	for _, device := range devices {
		actionJSON, err := json.Marshal(BulkAction{
			Type: "index",
			Desc: &BulkActionDesc{
				ID:    device.GetID(),
				Index: devIdx(device.GetTenantID()),
			},
		})
		if err != nil {
			return err
		}
		deviceJSON, err := json.Marshal(device)
		if err != nil {
			return err
		}
		data += string(actionJSON) + "\n" + string(deviceJSON) + "\n"

	}
	req := esapi.BulkRequest{
		Body: strings.NewReader(data),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "failed to bulk index")
	}
	defer res.Body.Close()

	return nil
}

func (s *store) Migrate(ctx context.Context) error {
	req := esapi.IndicesPutIndexTemplateRequest{
		Name: indexDevices,
		Body: strings.NewReader(indexDevicesTemplate),
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "failed to put the index template")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New("failed to set up the index template")
	}

	return nil
}

func (s *store) Search(ctx context.Context, query interface{}) (model.M, error) {
	l := log.FromContext(ctx)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	if l.Logger.IsLevelEnabled(logrus.DebugLevel) {
		l.Debugf("es query:\n%v\n", buf.String())
	}

	id := identity.FromContext(ctx)

	indexName := "devices"
	if id.Tenant != "" {
		indexName = indexName + "-" + id.Tenant
	}
	resp, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(indexName),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithTrackTotalHits(true),
	)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	var ret map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}

	return ret, nil
}
func (s *store) GetDevice(ctx context.Context, tenant, devid string) (*model.Device, error) {
	//l := log.FromContext(ctx)

	id := identity.FromContext(ctx)

	req := esapi.GetRequest{
		Index:      devIdx(id.Tenant),
		DocumentID: devid,
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get device")
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return nil, nil
		} else {
			return nil, errors.Errorf(
				"failed to get device from ES, code %d", res.StatusCode,
			)

		}
	}

	var storeRes map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&storeRes); err != nil {
		return nil, err
	}

	source, ok := storeRes["_source"].(map[string]interface{})
	if !ok {
		return nil, errors.New("can't process ES _source")
	}

	return model.NewDeviceFromEsSource(source)

}

type mgetDocs struct {
	Docs []mgetDoc `json:"docs"`
}

type mgetDoc struct {
	ID    string `json:"_id"`
	Index string `json:"_index"`
}

func (s *store) GetDevices(ctx context.Context,
	tenantDevs map[string][]string) ([]model.Device, error) {
	l := log.FromContext(ctx)

	body := mgetDocs{
		Docs: []mgetDoc{},
	}

	for tid, devs := range tenantDevs {
		for _, d := range devs {
			body.Docs = append(body.Docs, mgetDoc{d, devIdx(tid)})
		}
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req := esapi.MgetRequest{
		Body: bytes.NewReader(data),
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to mget devices")
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(fmt.Sprintf("failed to mget devices, code %d",
			res.StatusCode))
	}

	var storeRes map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&storeRes); err != nil {
		return nil, err
	}

	if l.Logger.IsLevelEnabled(logrus.DebugLevel) {
		l.Debugf("es mget result:\n%v\n", storeRes)
	}

	ret := []model.Device{}

	// result is a list of docs
	storeDocs := storeRes["docs"].([]interface{})

	// each doc has a '_source'
	// (if found and didn't trigger an error)
	for _, doc := range storeDocs {
		docM, ok := doc.(map[string]interface{})
		if !ok {
			return nil, errors.New("can't process doc")
		}

		// if not found - has 'found = false'
		found, ok := docM["found"].(bool)
		if ok && !found {
			continue
		}

		source, ok := docM["_source"].(map[string]interface{})
		if ok {
			dev, err := model.NewDeviceFromEsSource(source)
			if err != nil {
				return nil, errors.Wrap(err, "can't parse _source into model")
			}

			dev = dev.WithMeta(&model.DeviceMeta{
				SeqNo:       int64(docM["_seq_no"].(float64)),
				PrimaryTerm: int64(docM["_primary_term"].(float64)),
			})
			ret = append(ret, *dev)
		}

		// source not parsed after all - maybe doc triggered an error
		// we allow one kind of error, index not found (yet - before first device request)
		if !ok {
			e, ok := docM["error"].(map[string]interface{})
			if !ok {
				e := fmt.Sprintf(
					"neither '_source', 'found' nor 'error' found in doc %v",
					docM)
				return nil, errors.New(e)
			}

			etyp, ok := e["type"].(string)
			if !ok {
				return nil, errors.New("found doc error, but it has no type")
			}

			if etyp != "index_not_found_exception" {
				return nil, errors.New("unexpected error " + etyp)
			}

		}
	}

	l.Debugf("es mget parsed result:\n%v\n", ret)

	return ret, nil
}

func (s *store) UpdateDevice(ctx context.Context,
	tenantID,
	deviceID string,
	updateDev *model.Device) error {
	l := log.FromContext(ctx)

	id := identity.FromContext(ctx)

	body := map[string]interface{}{
		"doc": updateDev,
	}

	// DocumentType is _doc by default
	req := esapi.UpdateRequest{
		Index:      devIdx(id.Tenant),
		DocumentID: deviceID,
		Body:       esutil.NewJSONReader(body),
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "failed to update device in ES")
	}

	defer res.Body.Close()

	var esbody map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&esbody); err != nil {
		return err
	}
	l.Debugf("ES update response %v", esbody)

	switch {
	case err != nil:
		return errors.Wrap(err, "failed to update device in ES")
	case res.IsError():
		return errors.Errorf("failed to update device in ES, code %d", res.StatusCode)
	default:
		return nil
	}
}

// GetDevIndex retrieves the "devices*" index definition for tenant 'tid'
// existing fields, incl. inventory attributes, are found under 'properties'
// see: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-index.html
func (s *store) GetDevIndex(ctx context.Context, tid string) (map[string]interface{}, error) {
	l := log.FromContext(ctx)
	idx := devIdx(tid)

	req := esapi.IndicesGetRequest{
		Index: []string{idx},
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get devices index from store, tid %s", tid)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.Errorf(
			"failed to get devices index from store, tid %s, code %d",
			tid, res.StatusCode,
		)
	}

	var indexRes map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&indexRes); err != nil {
		return nil, err
	}

	index, ok := indexRes[idx]
	if !ok {
		return nil, errors.New("can't parse index defintion response")
	}

	indexM, ok := index.(map[string]interface{})
	if !ok {
		return nil, errors.New("can't parse index defintion response")
	}

	l.Debugf("devices index for tid %s\n%s\n", tid, indexM)

	return indexM, nil
}

func WithServerAddresses(addresses []string) StoreOption {
	return func(s *store) {
		s.addresses = addresses
	}
}

// devIdx prepares "devices" index name for tenant tid
func devIdx(tid string) string {
	if tid == "" {
		return indexDevices
	}
	return indexDevices + "-" + tid
}
