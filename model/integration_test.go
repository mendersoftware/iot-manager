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

package model

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func str2ptr(s string) *string {
	return &s
}

func TestIntegrationValidate(t *testing.T) {
	cs, _ := ParseConnectionString(
		"HostName=mender-test-hub.azure-devices.net;DeviceId=7b478313-de33-4735-bf00-0ebc31851faf;" +
			"SharedAccessKey=YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE=")

	testCases := map[string]struct {
		integration *Integration
		err         error
	}{
		"ok, Azure IoT Hub": {
			integration: &Integration{
				Provider: ProviderIoTHub,
				Credentials: Credentials{
					Type:             CredentialTypeSAS,
					ConnectionString: cs,
				},
			},
		},
		"ko, Azure IoT Hub": {
			integration: &Integration{
				Provider: ProviderIoTHub,
			},
			err: errors.New("credentials: (type: cannot be blank.)."),
		},
		"ok, AWS IoT Core": {
			integration: &Integration{
				Provider: ProviderIoTCore,
				Credentials: Credentials{
					Type:            CredentialTypeAWS,
					AccessKeyID:     str2ptr("x"),
					SecretAccessKey: str2ptr("x"),
					EndpointURL:     str2ptr("https://abcdefg123456.iot.us-east-1.amazonaws.com"),
					DevicePolicyARN: str2ptr("x"),
				},
			},
		},
		"ko, AWS IoT Core wrong URL": {
			integration: &Integration{
				Provider: ProviderIoTCore,
				Credentials: Credentials{
					Type:            CredentialTypeAWS,
					AccessKeyID:     str2ptr("x"),
					SecretAccessKey: str2ptr("x"),
					EndpointURL:     str2ptr("https://mender.io"),
					DevicePolicyARN: str2ptr("x"),
				},
			},
			err: errors.New("credentials: (endpoint_url: hostname does not refer to a trusted domain.)."),
		},
		"ko, AWS IoT Core": {
			integration: &Integration{
				Provider: ProviderIoTCore,
				Credentials: Credentials{
					Type: CredentialTypeAWS,
				},
			},
			err: errors.New("credentials: (access_key_id: cannot be blank; device_policy_arn: cannot be blank; endpoint_url: cannot be blank; secret_access_key: cannot be blank.)."),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.integration.Validate()
			if tc.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
