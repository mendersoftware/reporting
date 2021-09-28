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

import hmac
import json
import os
import pytest
import re
import uuid

from base64 import b64encode
from datetime import datetime, timedelta
from typing import Union
from urllib.parse import urljoin

import requests

from elasticsearch import Elasticsearch
from internal_api.models import InternalDevice, Attribute

type_decoder = {
    str: "str",
    int: "num",
    float: "num",
    bool: "bool",
}


def index_device(es: Elasticsearch, device: InternalDevice):
    doc = device.to_dict()
    if device.attributes is not None:
        for attr in device.attributes:
            typ = type_decoder[type(attr.value)]
            doc[f"{attr.scope}_{attr.name}_{typ}"] = [attr.value]
    try:
        doc.pop("attributes")
    except KeyError:
        pass
    doc["tenantID"] = device.tenant_id
    es.index(f"devices-{device.tenant_id}", doc, refresh="wait_for", id=device.id)


def attributes_to_document(attrs: list[Attribute]) -> dict[str, object]:
    doc = {}
    if attrs is not None:
        for attr in attrs:
            typ = type_decoder[type(attr.value)]
            doc[f"{attr.scope}_{attr.name}_{typ}"] = [attr.value]
    return doc


def generate_jwt(tenant_id: str = "", subject: str = "", is_user: bool = True) -> str:
    if len(subject) == 0:
        subject = str(uuid.uuid4())

    hdr = {
        "alg": "HS256",
        "typ": "JWT",
    }
    hdr64 = (
        b64encode(json.dumps(hdr).encode(), altchars=b"-_").decode("ascii").rstrip("=")
    )

    claims = {
        "sub": subject,
        "exp": (datetime.utcnow() + timedelta(hours=1)).isoformat("T"),
        "mender.device": not is_user,
        "mender.tenant": tenant_id,
    }
    if is_user:
        claims["mender.user"] = True
    else:
        claims["mender.device"] = True

    claims64 = (
        b64encode(json.dumps(claims).encode(), altchars=b"-_")
        .decode("ascii")
        .rstrip("=")
    )

    jwt = hdr64 + "." + claims64
    sign = hmac.new(b"secretJWTkey", msg=jwt.encode(), digestmod="sha256")
    sign64 = b64encode(sign.digest(), altchars=b"-_").decode("ascii").rstrip("=")
    return jwt + "." + sign64


_MMOCK_URL = os.getenv("MMOCK_CONTROL_URL")


class MockAPI:
    def __init__(
        self,
        method: str,
        path: str,
        rsp_code: int = 200,
        rsp_hdrs: Union[dict[str, str], None] = None,
        rsp_body: Union[str, None] = None,
        req_qparams: Union[dict[str, str]] = None,
        req_host: Union[str, None] = None,
        req_hdrs: Union[dict[str, str]] = None,
        req_body: Union[str, None] = None,
    ):
        self.method = method
        self.path = path
        self.rsp_code = rsp_code
        self.rsp_hdrs = rsp_hdrs
        self.rsp_body = rsp_body
        self.req_host = req_host
        self.req_qparams = req_qparams
        self.req_hdrs = req_hdrs
        self.req_body = req_body
        # API URL is on the form: "<method>_<*path*>.json where *path* is
        # path with all illegal path characters replaced by '_'.
        self._api_url = urljoin(
            _MMOCK_URL,
            f"/api/mapping/%s_%s.json" % (method, re.sub("[^0-9A-Za-z-_]", "_", path)),
        )

    @property
    def _request(self):
        js = {
            "request": {"method": self.method, "path": self.path},
            "response": {"statusCode": self.rsp_code},
        }
        if self.rsp_hdrs is not None:
            js["response"]["headers"] = {
                key: [value] for key, value in self.rsp_hdrs.items()
            }
        if self.rsp_body is not None:
            js["response"]["body"] = self.rsp_body

        if self.req_host is not None:
            js["request"]["host"] = self.req_host
        if self.req_qparams is not None:
            js["request"]["queryStringParameters"] = {
                key: [value] for key, value in self.req_qparams.items()
            }
        if self.req_hdrs is not None:
            js["request"]["headers"] = {key: [value] for key, value in self.req_hdrs}
        if self.req_body is not None:
            js["request"]["body"] = self.req_body

        return js

    def __enter__(self):
        url = urljoin(_MMOCK_URL, self._api_url)
        rsp = requests.post(url, json=self._request)
        if rsp.status_code == 409:
            rsp = requests.put(url, json=self._request)
        if rsp.status_code >= 300:
            raise ValueError(
                "mmock server returned an unexpected status code: %d" % rsp.status_code
            )

    def __exit__(self, exception_type, exception_value, exception_traceback):
        url = urljoin(_MMOCK_URL, self._api_url)
        rsp = requests.delete(url)
        if rsp.status_code >= 300:
            raise ValueError(
                "mmock server returned an unexpected status code: %d" % rsp.status_code
            )
