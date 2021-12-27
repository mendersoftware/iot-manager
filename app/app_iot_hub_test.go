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

package app

import (
	"context"
	"testing"

	"github.com/mendersoftware/iot-manager/client/iothub"
	mocks_iothub "github.com/mendersoftware/iot-manager/client/iothub/mocks"
	"github.com/mendersoftware/iot-manager/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGetDeviceStateIoTHub(t *testing.T) {
	integration := &model.Integration{
		Credentials: model.Credentials{
			ConnectionString: &model.ConnectionString{HostName: "dummy"},
		},
	}
	testCases := []struct {
		Name string

		DeviceID    string
		Integration *model.Integration

		IoTHubClient              func(t *testing.T) *mocks_iothub.Client
		GetDeviceStateIoTHub      *model.DeviceState
		GetDeviceStateIoTHubError error
	}{
		{
			Name: "ok",

			DeviceID:    "1",
			Integration: integration,

			IoTHubClient: func(t *testing.T) *mocks_iothub.Client {
				hub := &mocks_iothub.Client{}

				hub.On("GetDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
				).Return(&iothub.DeviceTwin{
					Properties: iothub.TwinProperties{
						Desired: map[string]interface{}{
							"key": "value",
						},
						Reported: map[string]interface{}{
							"another-key": "another-value",
						},
					},
				}, nil)
				return hub
			},
			GetDeviceStateIoTHub: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
				Reported: map[string]interface{}{
					"another-key": "another-value",
				},
			},
		},
		{
			Name: "ko, no connection string",

			DeviceID:    "1",
			Integration: &model.Integration{},

			GetDeviceStateIoTHubError: ErrNoConnectionString,
		},
		{
			Name: "ko, error retrieving the device twin",

			DeviceID:    "1",
			Integration: integration,

			IoTHubClient: func(t *testing.T) *mocks_iothub.Client {
				hub := &mocks_iothub.Client{}

				hub.On("GetDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
				).Return(nil, errors.New("internal error"))
				return hub
			},
			GetDeviceStateIoTHubError: errors.New("failed to get the device twin: internal error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			var iotHubClient iothub.Client
			if tc.IoTHubClient != nil {
				client := tc.IoTHubClient(t)
				defer client.AssertExpectations(t)
				iotHubClient = client
			}
			app := New(nil, iotHubClient, nil)

			ctx := context.Background()
			state, err := app.GetDeviceStateIoTHub(ctx, tc.DeviceID, tc.Integration)
			if tc.GetDeviceStateIoTHubError != nil {
				assert.EqualError(t, err, tc.GetDeviceStateIoTHubError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.GetDeviceStateIoTHub, state)
			}
		})
	}
}

func TestSetDeviceStateIoTHub(t *testing.T) {
	integration := &model.Integration{
		Credentials: model.Credentials{
			ConnectionString: &model.ConnectionString{HostName: "dummy"},
		},
	}
	testCases := []struct {
		Name string

		DeviceID    string
		Integration *model.Integration
		DeviceState *model.DeviceState

		IoTHubClient              func(t *testing.T) *mocks_iothub.Client
		SetDeviceStateIoTHub      *model.DeviceState
		SetDeviceStateIoTHubError error
	}{
		{
			Name: "ok",

			DeviceID:    "1",
			Integration: integration,
			DeviceState: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
			},

			IoTHubClient: func(t *testing.T) *mocks_iothub.Client {
				hub := &mocks_iothub.Client{}

				hub.On("UpdateDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
					&iothub.DeviceTwinUpdate{
						Properties: iothub.UpdateProperties{
							Desired: map[string]interface{}{
								"key": "value",
							},
						},
					},
				).Return(nil)

				hub.On("GetDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
				).Return(&iothub.DeviceTwin{
					Properties: iothub.TwinProperties{
						Desired: map[string]interface{}{
							"key": "value",
						},
						Reported: map[string]interface{}{
							"another-key": "another-value",
						},
					},
				}, nil)

				return hub
			},
			SetDeviceStateIoTHub: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
				Reported: map[string]interface{}{
					"another-key": "another-value",
				},
			},
		},
		{
			Name: "ko, no connection string",

			DeviceID:    "1",
			Integration: &model.Integration{},

			SetDeviceStateIoTHubError: ErrNoConnectionString,
		},
		{
			Name: "ko, error setting the device twin",

			DeviceID:    "1",
			Integration: integration,
			DeviceState: &model.DeviceState{
				Desired: map[string]interface{}{
					"key": "value",
				},
			},

			IoTHubClient: func(t *testing.T) *mocks_iothub.Client {
				hub := &mocks_iothub.Client{}

				hub.On("UpdateDeviceTwin",
					contextMatcher,
					integration.Credentials.ConnectionString,
					"1",
					&iothub.DeviceTwinUpdate{
						Properties: iothub.UpdateProperties{
							Desired: map[string]interface{}{
								"key": "value",
							},
						},
					},
				).Return(errors.New("internal error"))

				return hub
			},
			SetDeviceStateIoTHubError: errors.New("failed to update the device twin: internal error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			var iotHubClient iothub.Client
			if tc.IoTHubClient != nil {
				client := tc.IoTHubClient(t)
				defer client.AssertExpectations(t)
				iotHubClient = client
			}
			app := New(nil, iotHubClient, nil)

			ctx := context.Background()
			state, err := app.SetDeviceStateIoTHub(ctx, tc.DeviceID, tc.Integration, tc.DeviceState)
			if tc.SetDeviceStateIoTHubError != nil {
				assert.EqualError(t, err, tc.SetDeviceStateIoTHubError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.SetDeviceStateIoTHub, state)
			}
		})
	}
}
