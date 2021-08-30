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

	SettingInventoryAddr        = "inventory_addr"
	SettingInventoryAddrDefault = "http://mender-inventory:8080/"

	// SettingReindexBatchSize is the num of buffered requests processed together
	SettingReindexBatchSize        = "reindex_batch_size"
	SettingReindexBatchSizeDefault = 20

	// SettingReindexTimeMsec is the max time after which reindexing is triggered
	// (even if buffered requests didn't reach reindex_batch_size yet)
	SettingReindexMaxTimeMsec        = "reindex_max_time_msec"
	SettingReindexMaxTimeMsecDefault = 1000

	// SettingReindexBuffLen is the length of the reindex pipeline input
	// buffer/buffered channel (in number of reindex events)
	SettingReindexBuffLen        = "reindex_buff_len"
	SettingReindexBuffLenDefault = 100

	// SettingReindexNumWorkers is the num of workers actually issuing the reindex bulk requests
	SettingReindexNumWorkers        = "reindex_num_workers"
	SettingReindexNumWorkersDefault = 5

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
		{Key: SettingDebugLog, Value: SettingDebugLogDefault},
		{Key: SettingInventoryAddr, Value: SettingInventoryAddrDefault},
		{Key: SettingReindexBuffLen, Value: SettingReindexBuffLenDefault},
		{Key: SettingReindexMaxTimeMsec, Value: SettingReindexMaxTimeMsecDefault},
		{Key: SettingReindexBatchSize, Value: SettingReindexBatchSizeDefault},
		{Key: SettingReindexNumWorkers, Value: SettingReindexNumWorkersDefault},
	}
)
