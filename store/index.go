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

const (
	indexDevices         = "devices"
	indexDevicesTemplate = `{
	"index_patterns": ["devices-*"],
	"priority": 1,
	"template": {
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 1
		},
		"mappings": {
			"dynamic": "true",
			"_source": {
				"enabled": true
			},
			"properties": {
				"id": {
					"type": "keyword"
				},
				"tenantID": {
					"type": "keyword"
				},
				"name": {
					"type": "keyword"
				},
				"groupName": {
					"type": "keyword"
				},
				"status": {
					"type": "keyword"
				},
				"createdAt": {
					"type": "date"
				},
				"updatedAt": {
					"type": "date"
				}
			},
			"dynamic_templates": [
				{
					"versions": {
						"match": "*_version*",
						"mapping": {
							"type": "version"
						}
					}
				},
				{
					"inventory_strings": {
						"match": "inventory_*_str",
						"mapping": {
							"type": "keyword"
						}
					}
				},
				{
					"identity_strings": {
						"match": "identity_*_str",
						"mapping": {
							"type": "keyword"
						}
					}
				},
				{
					"custom_strings": {
						"match": "custom_*_str",
						"mapping": {
							"type": "keyword"
						}
					}
				},
				{
					"inventory_nums": {
						"match": "inventory_*_num",
						"mapping": {
							"type": "double"
						}
					}
				},
				{
					"identity_nums": {
						"match": "identity_*_num",
						"mapping": {
							"type": "double"
						}
					}
				},
				{
					"custom_nums": {
						"match": "custom_*_num",
						"mapping": {
							"type": "double"
						}
					}
				}
			]
		}
	}}`
)
