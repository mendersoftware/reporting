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

package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/pkg/errors"

	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/log"
	_ "github.com/mendersoftware/go-lib-micro/log"

	"github.com/mendersoftware/reporting/model"
	"github.com/mendersoftware/reporting/store"
)

type StoreOption func(*opensearchStore)

type opensearchStore struct {
	addresses            []string
	devicesIndexName     string
	devicesIndexShards   int
	devicesIndexReplicas int
	client               *opensearch.Client
}

func NewStore(opts ...StoreOption) (store.Store, error) {
	store := &opensearchStore{}
	for _, opt := range opts {
		opt(store)
	}

	cfg := opensearch.Config{
		Addresses: store.addresses,
	}
	osClient, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "invalid OpenSearch configuration")
	}

	store.client = osClient
	return store, nil
}

func WithServerAddresses(addresses []string) StoreOption {
	return func(s *opensearchStore) {
		s.addresses = addresses
	}
}

func WithDevicesIndexName(indexName string) StoreOption {
	return func(s *opensearchStore) {
		s.devicesIndexName = indexName
	}
}

func WithDevicesIndexShards(indexShards int) StoreOption {
	return func(s *opensearchStore) {
		s.devicesIndexShards = indexShards
	}
}

func WithDevicesIndexReplicas(indexReplicas int) StoreOption {
	return func(s *opensearchStore) {
		s.devicesIndexReplicas = indexReplicas
	}
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
	Routing       string `json:"routing"`
	Tenant        string
}

type BulkItem struct {
	Action *BulkAction
	Doc    interface{}
}

func (bad BulkActionDesc) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID      string `json:"_id"`
		Index   string `json:"_index"`
		Routing string `json:"routing"`
	}{
		ID:      bad.ID,
		Index:   bad.Index,
		Routing: bad.Routing,
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

func (s *opensearchStore) BulkIndexDevices(ctx context.Context, devices []*model.Device,
	removedDevices []*model.Device) error {
	var data strings.Builder

	for _, device := range devices {
		actionJSON, err := json.Marshal(BulkAction{
			Type: "index",
			Desc: &BulkActionDesc{
				ID:      device.GetID(),
				Index:   s.GetDevicesIndex(device.GetTenantID()),
				Routing: s.GetDevicesRoutingKey(device.GetTenantID()),
			},
		})
		if err != nil {
			return err
		}
		deviceJSON, err := json.Marshal(device)
		if err != nil {
			return err
		}
		data.WriteString(string(actionJSON) + "\n" + string(deviceJSON) + "\n")
	}
	for _, device := range removedDevices {
		actionJSON, err := json.Marshal(BulkAction{
			Type: "delete",
			Desc: &BulkActionDesc{
				ID:      device.GetID(),
				Index:   s.GetDevicesIndex(device.GetTenantID()),
				Routing: s.GetDevicesRoutingKey(device.GetTenantID()),
			},
		})
		if err != nil {
			return err
		}
		data.WriteString(string(actionJSON) + "\n")
	}

	dataString := data.String()

	l := log.FromContext(ctx)
	l.Debugf("opensearch request: %s", dataString)

	req := opensearchapi.BulkRequest{
		Body: strings.NewReader(dataString),
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "failed to bulk index")
	}
	defer res.Body.Close()

	return nil
}

func (s *opensearchStore) Migrate(ctx context.Context) error {
	indexName := s.GetDevicesIndex("")
	err := s.migratePutIndexTemplate(ctx, indexName)
	if err == nil {
		err = s.migrateCreateIndex(ctx, indexName)
	}
	return err
}

func (s *opensearchStore) migratePutIndexTemplate(ctx context.Context, indexName string) error {
	l := log.FromContext(ctx)
	l.Infof("put the index template for %s", indexName)

	template := fmt.Sprintf(indexDevicesTemplate,
		indexName,
		s.devicesIndexShards,
		s.devicesIndexReplicas,
	)
	req := opensearchapi.IndicesPutIndexTemplateRequest{
		Name: indexName,
		Body: strings.NewReader(template),
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "failed to put the index template")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("failed to set up the index template")
	}
	return nil
}

func (s *opensearchStore) migrateCreateIndex(ctx context.Context, indexName string) error {
	l := log.FromContext(ctx)
	l.Infof("verify if the index %s exists", indexName)

	req := opensearchapi.IndicesExistsRequest{
		Index: []string{indexName},
	}
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return errors.Wrap(err, "failed to verify the index")
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		l.Infof("create the index %s", indexName)

		req := opensearchapi.IndicesCreateRequest{
			Index: indexName,
		}
		res, err := req.Do(ctx, s.client)
		if err != nil {
			return errors.Wrap(err, "failed to create the index")
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return errors.New("failed to create the index")
		}
	} else if res.StatusCode != http.StatusOK {
		return errors.New("failed to verify the index")
	}

	return nil
}

func (s *opensearchStore) Ping(ctx context.Context) error {
	pingRequest := s.client.Ping.WithContext(ctx)
	_, err := s.client.Ping(pingRequest)
	return errors.Wrap(err, "failed to ping opensearch")
}

func (s *opensearchStore) Search(ctx context.Context, query interface{}) (model.M, error) {
	l := log.FromContext(ctx)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	l.Debugf("es query: %v", buf.String())

	id := identity.FromContext(ctx)

	searchRequests := []func(*opensearchapi.SearchRequest){
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(s.GetDevicesIndex(id.Tenant)),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithTrackTotalHits(true),
	}
	routingKey := s.GetDevicesRoutingKey(id.Tenant)
	if routingKey != "" {
		searchRequests = append(searchRequests, s.client.Search.WithRouting(routingKey))
	}
	resp, err := s.client.Search(searchRequests...)
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

	l.Debugf("opensearch response: %v", ret)

	return ret, nil
}

// GetDevicesIndexMapping retrieves the "devices*" index definition for tenant 'tid'
// existing fields, incl. inventory attributes, are found under 'properties'
// see: https://opensearch.org/docs/latest/api-reference/index-apis/get-index/
func (s *opensearchStore) GetDevicesIndexMapping(ctx context.Context,
	tid string) (map[string]interface{}, error) {
	l := log.FromContext(ctx)
	idx := s.GetDevicesIndex(tid)

	req := opensearchapi.IndicesGetRequest{
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

// GetDevicesIndex returns the index name for the tenant tid
func (s *opensearchStore) GetDevicesIndex(tid string) string {
	return s.devicesIndexName
}

// GetDevicesRoutingKey returns the routing key for the tenant tid
func (s *opensearchStore) GetDevicesRoutingKey(tid string) string {
	return tid
}
