# Copyright 2022 Northern.tech AS
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
import json
import re
import time

from datetime import datetime, timedelta
from typing import Union

import internal_api
import utils

from elasticsearch import exceptions


class TestInternalHealth:
    def test_internal_alive(self):
        client = internal_api.InternalAPIClient()
        r = client.check_liveliness_with_http_info(_preload_content=False)
        assert r.status == 204


test_set = [
    internal_api.models.InternalDevice(
        id="463e12dd-1adb-4f62-965e-b0a9ba2c93ff",
        tenant_id="123456789012345678901234",
        name="bagelBone",
        attributes=[
            internal_api.models.Attribute(
                name="string", value="Lorem ipsum dolor sit amet", scope="inventory"
            ),
            internal_api.models.Attribute(
                name="number", value=2 ** 47, scope="inventory"
            ),
        ],
    ),
    internal_api.models.InternalDevice(
        id="d8b04e01-690d-41ce-8c6d-ab079a04d488",
        tenant_id="123456789012345678901234",
        name="blueberryPi",
        attributes=[
            internal_api.models.Attribute(
                name="string", value="consectetur adipiscing elit", scope="inventory",
            ),
            internal_api.models.Attribute(
                name="number", value=420.69, scope="inventory"
            ),
        ],
    ),
    internal_api.models.InternalDevice(
        id="ad707aab-916b-4ec9-a43f-0031b2bcf9ad",
        tenant_id="123456789012345678901234",
        name="birch32",
        attributes=[
            internal_api.models.Attribute(
                name="string", value="sed do eiusmod tempor", scope="inventory",
            ),
        ],
    ),
    internal_api.models.InternalDevice(
        id="85388603-5852-437f-89c4-7549502893d5",
        tenant_id="123456789012345678901234",
        name="birch32",
        attributes=[
            internal_api.models.Attribute(
                name="string", value="incididunt ut labore", scope="inventory",
            ),
            internal_api.models.Attribute(name="number", value=0.0, scope="inventory"),
        ],
    ),
    internal_api.models.InternalDevice(
        id="98efdb94-26c2-42eb-828d-12a5d6eb698c",
        tenant_id="098765432109876543210987",
        name="shepherdsPi",
        attributes=[
            internal_api.models.Attribute(
                name="string", value="sample text", scope="inventory",
            ),
            internal_api.models.Attribute(name="number", value=1234, scope="inventory"),
        ],
    ),
]


@pytest.fixture(scope="class")
def setup_test_context(elasticsearch):
    # clean up any indices from previous tests
    indices = elasticsearch.cat.indices(format="json")
    for idx in indices:
        if not idx["index"].startswith("."):
            elasticsearch.delete_by_query(
                index=[idx["index"]], body={"query": {"match_all": {}}}
            )

    for dev in test_set:
        utils.index_device(dev)

    time.sleep(5)


class TestInternalSearch:
    class _TestCase:
        def __init__(
            self,
            tenant_id: str,
            search_terms: internal_api.models.SearchTerms,
            http_code: int,
            result: Union[list[internal_api.models.DeviceInventory], str],
        ):
            self.tenant_id = tenant_id
            self.search_terms = search_terms
            self.http_code = http_code
            self.result = result

    @pytest.mark.parametrize(
        argnames="test_case",
        argvalues=[
            _TestCase(
                tenant_id="123456789012345678901234",
                search_terms=internal_api.models.SearchTerms(
                    filters=[
                        internal_api.models.FilterTerm(
                            attribute="string",
                            value="Lorem ipsum dolor sit amet",
                            type="$eq",
                            scope="inventory",
                        )
                    ],
                ),
                http_code=200,
                result=[
                    internal_api.models.DeviceInventory(
                        id="463e12dd-1adb-4f62-965e-b0a9ba2c93ff",
                        attributes=[
                            internal_api.models.Attribute(
                                name="number", value=2 ** 47, scope="inventory"
                            ),
                            internal_api.models.Attribute(
                                name="string",
                                value="Lorem ipsum dolor sit amet",
                                scope="inventory",
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                tenant_id="123456789012345678901234",
                search_terms=internal_api.models.SearchTerms(
                    filters=[
                        internal_api.models.FilterTerm(
                            attribute="number",
                            value=2 ** 32,
                            type="$gt",
                            scope="inventory",
                        )
                    ],
                ),
                http_code=200,
                result=[
                    internal_api.models.DeviceInventory(
                        id="463e12dd-1adb-4f62-965e-b0a9ba2c93ff",
                        attributes=[
                            internal_api.models.Attribute(
                                name="number", value=2 ** 47, scope="inventory"
                            ),
                            internal_api.models.Attribute(
                                name="string",
                                value="Lorem ipsum dolor sit amet",
                                scope="inventory",
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                tenant_id="123456789012345678901234",
                search_terms=internal_api.models.SearchTerms(
                    filters=[
                        internal_api.models.FilterTerm(
                            attribute="string",
                            value=[
                                "Lorem ipsum dolor sit amet",
                                "consectetur adipiscing elit",
                            ],
                            type="$in",
                            scope="inventory",
                        )
                    ],
                    sort=[
                        internal_api.models.SortTerm(
                            attribute="number", scope="inventory", order="asc"
                        )
                    ],
                ),
                http_code=200,
                result=[
                    internal_api.models.InternalDevice(
                        id="d8b04e01-690d-41ce-8c6d-ab079a04d488",
                        attributes=[
                            internal_api.models.Attribute(
                                name="number", value=420.69, scope="inventory"
                            ),
                            internal_api.models.Attribute(
                                name="string",
                                value="consectetur adipiscing elit",
                                scope="inventory",
                            ),
                        ],
                    ),
                    internal_api.models.DeviceInventory(
                        id="463e12dd-1adb-4f62-965e-b0a9ba2c93ff",
                        attributes=[
                            internal_api.models.Attribute(
                                name="number", value=2 ** 47, scope="inventory"
                            ),
                            internal_api.models.Attribute(
                                name="string",
                                value="Lorem ipsum dolor sit amet",
                                scope="inventory",
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                tenant_id="123456789012345678901234",
                search_terms=internal_api.models.SearchTerms(
                    filters=[
                        internal_api.models.FilterTerm(
                            attribute="number",
                            value=2 ** 32,
                            type="$lt",
                            scope="inventory",
                        )
                    ],
                    sort=[
                        internal_api.models.SortTerm(
                            attribute="number", scope="inventory", order="asc"
                        )
                    ],
                ),
                http_code=200,
                result=[
                    internal_api.models.InternalDevice(
                        id="85388603-5852-437f-89c4-7549502893d5",
                        attributes=[
                            internal_api.models.Attribute(
                                name="string",
                                value="incididunt ut labore",
                                scope="inventory",
                            ),
                            internal_api.models.Attribute(
                                name="number", value=0.0, scope="inventory"
                            ),
                        ],
                    ),
                    internal_api.models.InternalDevice(
                        id="d8b04e01-690d-41ce-8c6d-ab079a04d488",
                        attributes=[
                            internal_api.models.Attribute(
                                name="string",
                                value="consectetur adipiscing elit",
                                scope="inventory",
                            ),
                            internal_api.models.Attribute(
                                name="number", value=420.69, scope="inventory"
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                tenant_id="123456789012345678901234",
                search_terms=internal_api.models.SearchTerms(
                    filters=[
                        internal_api.models.FilterTerm(
                            attribute="string",
                            value="consectetur adipiscing elit",
                            type="$ne",
                            scope="inventory",
                        )
                    ],
                    sort=[
                        internal_api.models.SortTerm(
                            attribute="string", scope="inventory", order="asc"
                        )
                    ],
                ),
                http_code=200,
                result=[
                    internal_api.models.InternalDevice(
                        id="463e12dd-1adb-4f62-965e-b0a9ba2c93ff",
                        attributes=[
                            internal_api.models.Attribute(
                                name="string",
                                value="Lorem ipsum dolor sit amet",
                                scope="inventory",
                            ),
                            internal_api.models.Attribute(
                                name="number", value=2 ** 47, scope="inventory"
                            ),
                        ],
                    ),
                    internal_api.models.InternalDevice(
                        id="85388603-5852-437f-89c4-7549502893d5",
                        attributes=[
                            internal_api.models.Attribute(
                                name="string",
                                value="incididunt ut labore",
                                scope="inventory",
                            ),
                            internal_api.models.Attribute(
                                name="number", value=0.0, scope="inventory"
                            ),
                        ],
                    ),
                    internal_api.models.InternalDevice(
                        id="ad707aab-916b-4ec9-a43f-0031b2bcf9ad",
                        attributes=[
                            internal_api.models.Attribute(
                                name="string",
                                value="sed do eiusmod tempor",
                                scope="inventory",
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                tenant_id="123456789012345678901234",
                search_terms=internal_api.models.SearchTerms(
                    filters=[
                        internal_api.models.FilterTerm(
                            attribute="number",
                            value=False,
                            type="$exists",
                            scope="inventory",
                        )
                    ],
                ),
                http_code=200,
                result=[
                    internal_api.models.InternalDevice(
                        id="ad707aab-916b-4ec9-a43f-0031b2bcf9ad",
                        attributes=[
                            internal_api.models.Attribute(
                                name="string",
                                value="sed do eiusmod tempor",
                                scope="inventory",
                            ),
                        ],
                    ),
                ],
            ),
            _TestCase(
                tenant_id="123456789012345678901234",
                search_terms=internal_api.models.SearchTerms(
                    filters=[
                        internal_api.models.FilterTerm(
                            attribute="string",
                            value="(Lorem|consectetur).*",
                            type="$regex",
                            scope="inventory",
                        )
                    ],
                    sort=[
                        internal_api.models.SortTerm(
                            attribute="string", scope="inventory", order="asc"
                        )
                    ],
                ),
                http_code=200,
                result=[
                    internal_api.models.InternalDevice(
                        id="463e12dd-1adb-4f62-965e-b0a9ba2c93ff",
                        attributes=[
                            internal_api.models.Attribute(
                                name="string",
                                value="Lorem ipsum dolor sit amet",
                                scope="inventory",
                            ),
                            internal_api.models.Attribute(
                                name="number", value=2 ** 47, scope="inventory"
                            ),
                        ],
                    ),
                    internal_api.models.InternalDevice(
                        id="d8b04e01-690d-41ce-8c6d-ab079a04d488",
                        attributes=[
                            internal_api.models.Attribute(
                                name="string",
                                value="consectetur adipiscing elit",
                                scope="inventory",
                            ),
                            internal_api.models.Attribute(
                                name="number", value=420.69, scope="inventory"
                            ),
                        ],
                    ),
                ],
            ),
        ],
        ids=[
            "ok, $eq",
            "ok, $gt",
            "ok, $in + sort",
            "ok, $lt + sort",
            "ok, $ne + sort",
            "ok, $exists",
            "ok, $regex + sort",
        ],
    )
    def test_internal_search(self, test_case, setup_test_context):
        client = internal_api.InternalAPIClient()

        try:
            body, status, headers = client.device_search_with_http_info(
                test_case.tenant_id, search_terms=test_case.search_terms
            )
        except internal_api.ApiException as r:
            body = r.body
            status = r.status
            headers = r.headers
        assert status == test_case.http_code
        if isinstance(test_case.result, str):
            assert isinstance(body, bytes)
            re.match(test_case.result, r.body.decode())
        elif len(test_case.result) > 0:
            assert isinstance(body, type(test_case.result))

            expected_ids = [dev.id for dev in test_case.result]
            actual_ids = [dev.id for dev in body]
            assert expected_ids == actual_ids
            for i, expected in enumerate(test_case.result):
                actual = body[i]
                for attr in expected.attributes:
                    assert attr in actual.attributes
