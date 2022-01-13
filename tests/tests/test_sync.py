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

import logging
import uuid

from os import path

import docker
import pytest

from client import ManagementAPIClient
from management_api import models


class TestSync:
    connection_string = "HostName=mock.azure-devices.net:8443;SharedAccessKeyName=TestSyncEnterprise;SharedAccessKey=c2VjcmV0"
    tenant_devices = {
        "TestSync01": [
            {
                "id": "44a66a7e-4b09-4665-b343-0c5d3ad9a2ed",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "cea63fbb-0b0b-4a92-bd47-8d310a220a16",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            {
                "id": "8ba0bc48-a669-462e-873f-f7de2f0e04c3",
                "deviceauth": "no auth",
                "hub": "disabled",
            },
            {
                "id": "895a1ee2-1be0-4e78-889a-7d753f53dbd5",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "a74ea639-8fae-4b17-ac3e-f9759dbdbfd9",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            # Remainder are inconsistent devices
            {
                # Inconsistent status
                "id": "1424a387-3431-425b-9f44-1c8eba21812d",
                "deviceauth": "accepted",
                "hub": "disabled",
            },
            {
                # Inconsistent status
                "id": "966095ec-6bdd-4a76-8498-3c0dffdb9ee2",
                "deviceauth": "rejected",
                "hub": "enabled",
            },
            # 3x Devices stored locally but not mirrored everywhere
            {"id": "9b2083e4-83b6-41d3-b089-4d2ec137620b"},
            {"id": "93406e21-8e3f-4435-9786-a294a70298ee", "deviceauth": "accepted"},
            {"id": "7abb6133-ad97-44ba-a159-674242ee565e", "hub": "disabled"},
        ],
        "TestSync02": [
            {
                "id": "1e657abe-ad58-4d20-af7a-3a3449a405e7",
                "deviceauth": "no auth",
                "hub": "disabled",
            },
            {
                "id": "43396b5b-ff5c-45f2-ab91-150957f037c9",
                "deviceauth": "no auth",
                "hub": "disabled",
            },
            {
                "id": "5a629188-f221-49b1-aece-4fc336597834",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "653e15bc-3797-4eb0-b71e-46383096f512",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "917ae5d9-b7ae-4ae4-970e-7b2801ba84c3",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "9d3360c0-147e-4018-828f-4331a0e2a8ef",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            {
                "id": "a43eefc2-2370-468c-b614-710e16a8379f",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "a7050cae-be5c-4ea7-87c9-d4b62253f21b",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "aba6808c-ca58-40d5-97b7-fba13b68a42d",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            {
                "id": "b216b66d-c09c-4cb5-9f50-68ea45cdd6f4",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            {
                "id": "b5be0df7-ddb6-449f-b456-9ecf5dc54de5",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            {
                "id": "b7221134-6058-4f5e-9d51-ccc7d58d648b",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
            {
                "id": "b9c488af-4c79-4b1f-b1bb-5626ace246de",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "c650152a-e7d3-44b0-9ca7-10ec78caaff3",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "c65d71e9-cfaf-4d5a-9155-b3b043901b8c",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "cdfa43e7-eb2e-4fb0-81e0-93d7d8521cbf",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "cf1baca8-49f0-4dc5-91bd-c742292ea497",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "e1e686ae-925e-4d5b-a12a-ade8ff11b855",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "eb8a74a6-dc6b-4930-95d2-5d1db5da1be0",
                "deviceauth": "accepted",
                "hub": "enabled",
            },
            {
                "id": "fd06ea0c-0ff6-4751-b7a9-44c5ce40e2fe",
                "deviceauth": "rejected",
                "hub": "disabled",
            },
        ],
    }

    def test_sync(self, clean_mongo, cli_iot_manager):
        mgo = clean_mongo
        dc = docker.from_env()
        for tenant_id, devices in self.tenant_devices.items():
            conn_str = f"HostName=mock.azure-devices.net:8443;SharedAccessKeyName={tenant_id};SharedAccessKey=c2VjcmV0"
            client = ManagementAPIClient(tenant_id)
            _, code, hdr = client.register_integration(
                models.Integration(
                    provider="iot-hub",
                    credentials=models.Credentials(
                        type="sas", connection_string=conn_str
                    ),
                ),
                _return_http_data_only=False,
            )
            assert code == 201
            assert "Location" in hdr
            location_basename = path.basename(hdr.get("Location"))
            assert location_basename != ""
            integration_id = uuid.UUID(location_basename)

            dev_docs = [
                {
                    "_id": dev["id"],
                    "tenant_id": tenant_id,
                    "integration_ids": [integration_id],
                }
                for dev in devices
            ]
            mgo.iot_manager.devices.insert_many(dev_docs)

            # ... to be continued (FIXME)

        code, output = cli_iot_manager.sync_devices(batch_size=10)
        assert code == 0, output.decode("ascii")
        print(output.decode("ascii"))

        # Requests:
        # Tenant01:
        # - devauth: GET /devices
        ## Trim devices not in deviceauth
        # - hub: DELETE devices/7abb6133-ad97-44ba-a159-674242ee565e
        # - hub: DELETE devices/9b2083e4-83b6-41d3-b089-4d2ec137620b
        ## Get device twins
        # - hub: GET /devices/query
        ## Update inconsistent device statuses
        # - hub: GET /devices/1424a387-3431-425b-9f44-1c8eba21812d
        # - hub: PUT /devices/1424a387-3431-425b-9f44-1c8eba21812d
        # - hub: GET /devices/966095ec-6bdd-4a76-8498-3c0dffdb9ee2
        # - hub: PUT /devices/966095ec-6bdd-4a76-8498-3c0dffdb9ee2
        ## Provision device:
        # - hub: PUT /deivces/93406e21-8e3f-4435-9786-a294a70298ee
        # Tenant02:
        # - devauth: GET /devices
        # - hub: /devices/query
        # - devauth: GET /devices
        # - hub: /devices/query
