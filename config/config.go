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

package config

import (
	"github.com/mendersoftware/go-lib-micro/config"
)

const (
	// SettingListen is the config key for the listen address
	SettingListen = "listen"
	// SettingListenDefault is the default value for the listen address
	SettingListenDefault = ":8080"

	// SettingElasticsearchAddresses is the config key for the elasticsearch addresses
	SettingElasticsearchAddresses = "elasticsearch_addresses"
	// SettingElasticsearchAddressesDefault is the default value for the elasticsearch addresses
	SettingElasticsearchAddressesDefault = "http://localhost:9200"

	// SettingElasticsearchDevicesIndexName is the config key for the elasticsearch devices
	// index name
	SettingElasticsearchDevicesIndexName = "elasticsearch_devices_index_name"
	// SettingElasticsearchDevicesIndexNameDefault is the default value for the elasticsearch
	// devices index name
	SettingElasticsearchDevicesIndexNameDefault = "devices"

	// SettingElasticsearchDevicesIndexShards is the config key for the elasticsearch devices
	// index shards
	SettingElasticsearchDevicesIndexShards = "elasticsearch_devices_index_shards"
	// SettingElasticsearchDevicesIndexShardsDefault is the default value for the elasticsearch
	// devices index shards
	SettingElasticsearchDevicesIndexShardsDefault = 1

	// SettingElasticsearchDevicesIndexReplicas is the config key for the elasticsearch devices
	// index replicas
	SettingElasticsearchDevicesIndexReplicas = "elasticsearch_devices_index_replicas"
	// SettingElasticsearchDevicesIndexReplicasDefault is the default value for the
	// elasticsearch devices index replicas
	SettingElasticsearchDevicesIndexReplicasDefault = 0

	// SettingDeviceAuthAddr is the config key for the deviceauth service address
	SettingDeviceAuthAddr = "deviceauth_addr"
	// SettingDeviceAuthAddrDefault is the default value for the deviceauth service address
	SettingDeviceAuthAddrDefault = "http://mender-device-auth:8080/"

	// SettingInventoryAddr is the config key for the inventory service address
	SettingInventoryAddr = "inventory_addr"
	// SettingInventoryAddrDefault is the default value for the inventory service address
	SettingInventoryAddrDefault = "http://mender-inventory:8080/"

	// SettingNatsURI is the config key for the nats uri
	SettingNatsURI = "nats_uri"
	// SettingNatsURIDefault is the default value for the nats uri
	SettingNatsURIDefault = "nats://mender-nats:4222"

	// SettingNatsStreamName is the config key for the nats streaem name
	SettingNatsStreamName = "nats_stream_name"
	// SettingNatsStreamNameDefault is the default value for the nats stream name
	SettingNatsStreamNameDefault = "WORKFLOWS"

	// SettingNatsSubscriberTopic is the config key for the nats subscriber topic name
	SettingNatsSubscriberTopic = "nats_subscriber_topic"
	// SettingNatsSubscriberTopicDefault is the default value for the nats subscriber topic name
	SettingNatsSubscriberTopicDefault = "reporting"

	// SettingNatsSubscriberDurable is the config key for the nats subscriber durable name
	SettingNatsSubscriberDurable = "nats_subscriber_durable"
	// SettingNatsSubscriberDurableDefault is the default value for the nats subscriber durable
	// name
	SettingNatsSubscriberDurableDefault = "reporting"

	// SettingReindexBatchSize is the num of buffered requests processed together
	SettingReindexBatchSize        = "reindex_batch_size"
	SettingReindexBatchSizeDefault = 100

	// SettingReindexTimeMsec is the max time after which reindexing is triggered
	// (even if buffered requests didn't reach reindex_batch_size yet)
	SettingReindexMaxTimeMsec        = "reindex_max_time_msec"
	SettingReindexMaxTimeMsecDefault = 1000

	// SettingDebugLog is the config key for the truning on the debug log
	SettingDebugLog = "debug_log"
	// SettingDebugLogDefault is the default value for the debug log enabling
	SettingDebugLogDefault = false
)

var (
	// Defaults are the default configuration settings
	Defaults = []config.Default{
		{Key: SettingListen, Value: SettingListenDefault},
		{Key: SettingElasticsearchAddresses, Value: SettingElasticsearchAddressesDefault},
		{Key: SettingElasticsearchDevicesIndexName,
			Value: SettingElasticsearchDevicesIndexNameDefault},
		{Key: SettingElasticsearchDevicesIndexShards,
			Value: SettingElasticsearchDevicesIndexShardsDefault},
		{Key: SettingElasticsearchDevicesIndexReplicas,
			Value: SettingElasticsearchDevicesIndexReplicasDefault},
		{Key: SettingDebugLog, Value: SettingDebugLogDefault},
		{Key: SettingDeviceAuthAddr, Value: SettingDeviceAuthAddrDefault},
		{Key: SettingInventoryAddr, Value: SettingInventoryAddrDefault},
		{Key: SettingNatsURI, Value: SettingNatsURIDefault},
		{Key: SettingNatsStreamName, Value: SettingNatsStreamNameDefault},
		{Key: SettingNatsSubscriberTopic, Value: SettingNatsSubscriberTopicDefault},
		{Key: SettingNatsSubscriberDurable, Value: SettingNatsSubscriberDurableDefault},
		{Key: SettingReindexMaxTimeMsec, Value: SettingReindexMaxTimeMsecDefault},
		{Key: SettingReindexBatchSize, Value: SettingReindexBatchSizeDefault},
	}
)
