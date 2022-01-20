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

import socket

import docker
import requests

from management_api.apis import ManagementAPIClient as GenManagementAPIClient
from management_api import Configuration, ApiClient
from utils import generate_jwt


class ManagementAPIClient(GenManagementAPIClient):
    def __init__(self, tenant_id, subject="tester"):
        jwt = generate_jwt(tenant_id, subject, is_user=True)
        config = Configuration(
            host="http://mender-iot-manager:8080/api/management/v1/iot-manager",
            access_token=jwt,
        )
        client = ApiClient(configuration=config)
        super().__init__(api_client=client)


class CliIoTManager:
    def __init__(self):
        self.docker = docker.from_env()
        _self = self.docker.containers.list(filters={"id": socket.gethostname()})[0]

        project = _self.labels.get("com.docker.compose.project")
        self.iot_manager = self.docker.containers.list(
            filters={
                "label": [
                    f"com.docker.compose.project={project}",
                    "com.docker.compose.service=mender-iot-manager",
                ]
            },
            limit=1,
        )[0]

    def sync_devices(self, fail_early=False, batch_size=None, **kwargs):
        cmd = ["/usr/bin/iot-manager", "sync-devices"]
        if batch_size:
            cmd.append("--batch-size")
            cmd.append(str(batch_size))

        if fail_early:
            cmd.append("--fail-early")

        return self.iot_manager.exec_run(cmd, **kwargs)


class MMockAPIClient:
    def __init__(self, mmock_url: str):
        self.mmock_url = mmock_url.removesuffix("/")

    def reset(self):
        requests.get(self.mmock_url + "/api/request/reset")

    @property
    def unmatched(self) -> list[dict]:
        rsp = requests.get(self.mmock_url + "/api/request/unmatched")
        return rsp.json()

    @property
    def matched(self) -> list[dict]:
        rsp = requests.get(self.mmock_url + "/api/request/matched")
        return rsp.json()
