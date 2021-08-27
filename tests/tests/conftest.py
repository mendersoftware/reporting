# Copyright 2021 Northern.tech AS
#
#    Licensed under the Apache License, Version 2.0 (the "License");
#    you may not use this file except in compliance with the License.
#    You may obtain a copy of the License at
#
#        http://www.apache.org/licenses/LICENSE-2.0
#
#    Unless required by applicable law or agreed to in writing, software
#    distributed under the License is distributed on an "AS IS" BASIS,
#    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#    See the License for the specific language governing permissions and
#    limitations under the License.

import pytest
import os

from elasticsearch import Elasticsearch


@pytest.fixture(scope="session")
def elasticsearch():
    hosts = os.getenv("ELASTICSEARCH_URL").split(",")
    client = Elasticsearch(hosts=hosts)
    yield client
    client.close()


@pytest.fixture(scope="function")
def clean_es(elasticsearch):
    indices = elasticsearch.cat.indices(format="json")
    for idx in indices:
        elasticsearch.indices.delete(idx["index"])
    yield elasticsearch
