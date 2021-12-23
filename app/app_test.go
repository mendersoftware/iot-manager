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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/iot-manager/model"
	storeMocks "github.com/mendersoftware/iot-manager/store/mocks"
)

var contextMatcher = mock.MatchedBy(func(ctx context.Context) bool { return true })

func TestHealthCheck(t *testing.T) {
	testCases := []struct {
		Name string

		PingReturn    error
		ExpectedError error
	}{
		{
			Name:          "db Ping failed",
			PingReturn:    errors.New("failed to connect to db"),
			ExpectedError: errors.New("failed to connect to db"),
		},
		{
			Name: "db Ping successful",
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("Ping",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
			).Return(tc.PingReturn)
			app := New(store, nil, nil)

			ctx := context.Background()
			err := app.HealthCheck(ctx)
			if tc.ExpectedError != nil {
				assert.EqualError(t, err, tc.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// func TestGetIntegrations(t *testing.T) {
// 	testCases := []struct {
// 		Name string

// 		GetIntegrationsData  model.Integration
// 		GetIntegrationsError error
// 	}{
// 		{
// 			Name: "integration exists",

// 			GetIntegrationsData: model.Integration{
// 				Provider: model.AzureIoTHub,
// 				Credentials: model.Credentials{
// 					Type: "connection_string",
// 					Creds: &model.ConnectionString{
// 						HostName: "localhost",
// 						Key:      []byte("secret"),
// 						Name:     "foobar",
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Name: "integration do not exist",

// 			GetIntegrationsData: model.Integration{},
// 		},
// 		{
// 			Name: "integration retrieval error",

// 			GetIntegrationsError: errors.New("error getting the integration"),
// 		},
// 	}
// 	for i := range testCases {
// 		tc := testCases[i]
// 		t.Run(tc.Name, func(t *testing.T) {
// 			store := &storeMocks.DataStore{}
// 			store.On("GetIntegrations",
// 				mock.MatchedBy(func(ctx context.Context) bool {
// 					return true
// 				}),
// 			).Return(tc.GetIntegrationsData, tc.GetIntegrationsError)
// 			app := New(store, nil, nil)

// 			ctx := context.Background()
// 			settings, err := app.GetIntegrations(ctx)
// 			if tc.GetIntegrationsError != nil {
// 				assert.EqualError(t, err, tc.GetIntegrationsError.Error())
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tc.GetIntegrationsData, settings)
// 			}
// 		})
// 	}
// }

func TestCreateIntegration(t *testing.T) {
	testCases := []struct {
		Name string

		CreateIntegrationData model.Integration
		SetSettingsError      error
	}{
		{
			Name: "integration created",

			CreateIntegrationData: model.Integration{
				Provider: model.AzureIoTHub,
				Credentials: model.Credentials{
					Type: "connection_string",
					Creds: &model.ConnectionString{
						HostName: "localhost",
						Key:      []byte("secret"),
						Name:     "foobar",
					},
				},
			},
		},
		{
			Name: "create integration error",

			SetSettingsError: errors.New("error creating the integration"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("CreateIntegration",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
				mock.AnythingOfType("model.Integration"),
			).Return(tc.SetSettingsError)
			app := New(store, nil, nil)

			ctx := context.Background()
			err := app.CreateIntegration(ctx, tc.CreateIntegrationData)
			if tc.SetSettingsError != nil {
				assert.EqualError(t, err, tc.SetSettingsError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// func TestProvisionDevice(t *testing.T) {
// 	t.Parallel()
// 	type testCase struct {
// 		Name string

// 		ConnStr  *model.ConnectionString
// 		DeviceID string

// 		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
// 		Hub   func(t *testing.T, self *testCase) *miothub.Client
// 		Wf    func(t *testing.T, self *testCase) *mworkflows.Client

// 		Error error
// 	}
// 	testCases := []testCase{{
// 		Name: "ok",

// 		ConnStr: &model.ConnectionString{
// 			HostName: "localhost",
// 			Key:      []byte("super secret"),
// 			Name:     "my favorite string",
// 		},
// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{
// 					{
// 						Provider: model.AzureIoTHub,
// 						Credentials: model.Credentials{
// 							Type: "connection_string",
// 							Creds: &model.ConnectionString{
// 								HostName: "localhost",
// 								Key:      []byte("secret"),
// 								Name:     "foobar",
// 							},
// 						},
// 					},
// 				}, nil)
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			hub := new(miothub.Client)
// 			hub.On("UpsertDevice", contextMatcher, self.ConnStr, self.DeviceID).
// 				Return(&iothub.Device{
// 					DeviceID: self.DeviceID,
// 					Auth: &iothub.Auth{
// 						Type: iothub.AuthTypeSymmetric,
// 						SymmetricKey: &iothub.SymmetricKey{
// 							Primary:   iothub.Key("key1"),
// 							Secondary: iothub.Key("key2"),
// 						},
// 					},
// 				}, nil).
// 				On("UpdateDeviceTwin", contextMatcher, self.ConnStr, self.DeviceID,
// 					&iothub.DeviceTwinUpdate{
// 						Tags: map[string]interface{}{"mender": true},
// 					}).
// 				Return(nil)
// 			return hub
// 		},
// 		Wf: func(t *testing.T, self *testCase) *mworkflows.Client {
// 			wf := new(mworkflows.Client)
// 			primKey := &model.ConnectionString{
// 				Key:      []byte("key1"),
// 				DeviceID: self.DeviceID,
// 				HostName: self.ConnStr.HostName,
// 			}
// 			secKey := &model.ConnectionString{
// 				Key:      []byte("key2"),
// 				DeviceID: self.DeviceID,
// 				HostName: self.ConnStr.HostName,
// 			}
// 			wf.On("ProvisionExternalDevice",
// 				contextMatcher,
// 				self.DeviceID,
// 				map[string]string{
// 					confKeyPrimaryKey:   primKey.String(),
// 					confKeySecondaryKey: secKey.String(),
// 				}).Return(nil)
// 			return wf
// 		},
// 	}, {
// 		Name: "error/device does not have a connection string",

// 		ConnStr: &model.ConnectionString{
// 			HostName: "localhost",
// 			Key:      []byte("super secret"),
// 			Name:     "my favorite string",
// 		},
// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{
// 					{
// 						Provider: model.AzureIoTHub,
// 						Credentials: model.Credentials{
// 							Type: "connection_string",
// 							Creds: &model.ConnectionString{
// 								HostName: "localhost",
// 								Key:      []byte("secret"),
// 								Name:     "foobar",
// 							},
// 						},
// 					},
// 				}, nil)
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			hub := new(miothub.Client)
// 			hub.On("UpsertDevice", contextMatcher, self.ConnStr, self.DeviceID).
// 				Return(&iothub.Device{
// 					DeviceID: self.DeviceID,
// 					Auth: &iothub.Auth{
// 						Type: iothub.AuthTypeNone,
// 					},
// 				}, nil)
// 			return hub
// 		},
// 		Wf: func(t *testing.T, self *testCase) *mworkflows.Client {
// 			return new(mworkflows.Client)
// 		},

// 		Error: ErrNoDeviceConnectionString,
// 	}, {
// 		Name: "error/hub failure",

// 		ConnStr: &model.ConnectionString{
// 			HostName: "localhost",
// 			Key:      []byte("super secret"),
// 			Name:     "my favorite string",
// 		},
// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{
// 					{
// 						Provider: model.AzureIoTHub,
// 						Credentials: model.Credentials{
// 							Type: "connection_string",
// 							Creds: &model.ConnectionString{
// 								HostName: "localhost",
// 								Key:      []byte("secret"),
// 								Name:     "foobar",
// 							},
// 						},
// 					},
// 				}, nil)
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			hub := new(miothub.Client)
// 			hub.On("UpsertDevice", contextMatcher, self.ConnStr, self.DeviceID).
// 				Return(nil, errors.New("internal error"))
// 			return hub
// 		},
// 		Wf: func(t *testing.T, self *testCase) *mworkflows.Client {
// 			return new(mworkflows.Client)
// 		},

// 		Error: errors.New("failed to update iothub devices: internal error"),
// 	}, {
// 		Name: "error/no connection string",

// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{
// 					{
// 						Provider: model.AzureIoTHub,
// 						Credentials: model.Credentials{
// 							Type:  "connection_string",
// 							Creds: &model.ConnectionString{},
// 						},
// 					},
// 				}, nil)
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			return new(miothub.Client)
// 		},
// 		Wf: func(t *testing.T, self *testCase) *mworkflows.Client {
// 			return new(mworkflows.Client)
// 		},

// 		Error: ErrNoConnectionString,
// 	}, {
// 		Name: "error/getting settings",

// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{}, errors.New("wut?"))
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			return new(miothub.Client)
// 		},
// 		Wf: func(t *testing.T, self *testCase) *mworkflows.Client {
// 			return new(mworkflows.Client)
// 		},

// 		Error: errors.New("failed to retrieve settings: wut?"),
// 	}}
// 	for i := range testCases {
// 		tc := testCases[i]
// 		t.Run(tc.Name, func(t *testing.T) {
// 			t.Parallel()
// 			ctx := context.Background()
// 			ds := tc.Store(t, &tc)
// 			hub := tc.Hub(t, &tc)
// 			wf := tc.Wf(t, &tc)
// 			defer ds.AssertExpectations(t)
// 			defer hub.AssertExpectations(t)
// 			defer wf.AssertExpectations(t)

// 			app := New(ds, hub, wf)
// 			err := app.ProvisionDevice(ctx, tc.DeviceID)

// 			if tc.Error != nil {
// 				if assert.Error(t, err) {
// 					assert.Regexp(t, tc.Error.Error(), err.Error())
// 				}
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestDecommissionDevice(t *testing.T) {
// 	t.Parallel()
// 	type testCase struct {
// 		Name string

// 		ConnStr  *model.ConnectionString
// 		DeviceID string

// 		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
// 		Hub   func(t *testing.T, self *testCase) *miothub.Client

// 		Error error
// 	}
// 	testCases := []testCase{{
// 		Name: "ok",

// 		ConnStr: &model.ConnectionString{
// 			HostName: "localhost",
// 			Key:      []byte("super secret"),
// 			Name:     "my favorite string",
// 		},
// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{
// 					{
// 						Provider: model.AzureIoTHub,
// 						Credentials: model.Credentials{
// 							Type: "connection_string",
// 							Creds: &model.ConnectionString{
// 								HostName: "localhost",
// 								Key:      []byte("secret"),
// 								Name:     "foobar",
// 							},
// 						},
// 					},
// 				}, nil)
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			hub := new(miothub.Client)
// 			hub.On("DeleteDevice", contextMatcher, self.ConnStr, self.DeviceID).
// 				Return(nil)
// 			return hub
// 		},
// 	}, {
// 		Name: "error/hub failure",

// 		ConnStr: &model.ConnectionString{
// 			HostName: "localhost",
// 			Key:      []byte("super secret"),
// 			Name:     "my favorite string",
// 		},
// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{
// 					{
// 						Provider: model.AzureIoTHub,
// 						Credentials: model.Credentials{
// 							Type: "connection_string",
// 							Creds: &model.ConnectionString{
// 								HostName: "localhost",
// 								Key:      []byte("secret"),
// 								Name:     "foobar",
// 							},
// 						},
// 					},
// 				}, nil)
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			hub := new(miothub.Client)
// 			hub.On("DeleteDevice", contextMatcher, self.ConnStr, self.DeviceID).
// 				Return(errors.New("internal error"))
// 			return hub
// 		},

// 		Error: errors.New("failed to delete IoT Hub device: internal error"),
// 	}, {
// 		Name: "error/no connection string",

// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{
// 					{
// 						Provider: model.AzureIoTHub,
// 						Credentials: model.Credentials{
// 							Type:  "connection_string",
// 							Creds: &model.ConnectionString{},
// 						},
// 					},
// 				}, nil)
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			return new(miothub.Client)
// 		},

// 		Error: ErrNoConnectionString,
// 	}, {
// 		Name: "error/getting integrations",

// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{}, errors.New("wut?"))
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			return new(miothub.Client)
// 		},

// 		Error: errors.New("failed to retrieve settings: wut?"),
// 	}}
// 	for i := range testCases {
// 		tc := testCases[i]
// 		t.Run(tc.Name, func(t *testing.T) {
// 			t.Parallel()
// 			ctx := context.Background()
// 			ds := tc.Store(t, &tc)
// 			hub := tc.Hub(t, &tc)
// 			defer ds.AssertExpectations(t)
// 			defer hub.AssertExpectations(t)

// 			app := New(ds, hub, nil)
// 			err := app.DeleteIOTHubDevice(ctx, tc.DeviceID)

// 			if tc.Error != nil {
// 				if assert.Error(t, err) {
// 					assert.Regexp(t, tc.Error.Error(), err.Error())
// 				}
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestSetDeviceStatus(t *testing.T) {
// 	t.Parallel()
// 	type testCase struct {
// 		Name string

// 		ConnStr  *model.ConnectionString
// 		DeviceID string
// 		Status   Status

// 		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
// 		Hub   func(t *testing.T, self *testCase) *miothub.Client

// 		Error error
// 	}
// 	testCases := []testCase{{
// 		Name: "ok",

// 		Status: StatusEnabled,
// 		ConnStr: &model.ConnectionString{
// 			HostName: "localhost",
// 			Key:      []byte("super secret"),
// 			Name:     "my favorite string",
// 		},
// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{
// 					{
// 						Provider: model.AzureIoTHub,
// 						Credentials: model.Credentials{
// 							Type: "connection_string",
// 							Creds: &model.ConnectionString{
// 								HostName: "localhost",
// 								Key:      []byte("secret"),
// 								Name:     "foobar",
// 							},
// 						},
// 					},
// 				}, nil)
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			hub := new(miothub.Client)
// 			dev := &iothub.Device{
// 				DeviceID: "foobar",
// 				Status:   iothub.StatusDisabled,
// 			}
// 			hub.On("GetDevice", contextMatcher, self.ConnStr, self.DeviceID).
// 				Return(dev, nil).
// 				On("UpsertDevice", contextMatcher, self.ConnStr, self.DeviceID,
// 					mock.MatchedBy(func(dev *iothub.Device) bool {
// 						return dev.Status == iothub.StatusEnabled
// 					})).
// 				Return(dev, nil)
// 			return hub
// 		},
// 	}, {
// 		Name: "ok, device already has matching status",

// 		Status: StatusEnabled,
// 		ConnStr: &model.ConnectionString{
// 			HostName: "localhost",
// 			Key:      []byte("super secret"),
// 			Name:     "my favorite string",
// 		},
// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{
// 					{
// 						Provider: model.AzureIoTHub,
// 						Credentials: model.Credentials{
// 							Type: "connection_string",
// 							Creds: &model.ConnectionString{
// 								HostName: "localhost",
// 								Key:      []byte("secret"),
// 								Name:     "foobar",
// 							},
// 						},
// 					},
// 				}, nil)
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			hub := new(miothub.Client)
// 			dev := &iothub.Device{
// 				DeviceID: "foobar",
// 				Status:   iothub.StatusEnabled,
// 			}
// 			hub.On("GetDevice", contextMatcher, self.ConnStr, self.DeviceID).
// 				Return(dev, nil)
// 			return hub
// 		},
// 	}, {
// 		Name: "error/hub fail to update device",

// 		Status: StatusDisabled,
// 		ConnStr: &model.ConnectionString{
// 			HostName: "localhost",
// 			Key:      []byte("super secret"),
// 			Name:     "my favorite string",
// 		},
// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{
// 					{
// 						Provider: model.AzureIoTHub,
// 						Credentials: model.Credentials{
// 							Type: "connection_string",
// 							Creds: &model.ConnectionString{
// 								HostName: "localhost",
// 								Key:      []byte("secret"),
// 								Name:     "foobar",
// 							},
// 						},
// 					},
// 				}, nil)
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			hub := new(miothub.Client)
// 			dev := &iothub.Device{
// 				DeviceID: "foobar",
// 				Status:   iothub.StatusEnabled,
// 			}
// 			hub.On("GetDevice", contextMatcher, self.ConnStr, self.DeviceID).
// 				Return(dev, nil).
// 				On("UpsertDevice", contextMatcher, self.ConnStr, self.DeviceID,
// 					mock.MatchedBy(func(dev *iothub.Device) bool {
// 						return dev.Status == iothub.StatusDisabled
// 					})).
// 				Return(nil, errors.New("internal error"))
// 			return hub
// 		},

// 		Error: errors.New("failed to update IoT Hub device: internal error"),
// 	}, {
// 		Name: "error/hub fail to get device",

// 		Status: StatusDisabled,
// 		ConnStr: &model.ConnectionString{
// 			HostName: "localhost",
// 			Key:      []byte("super secret"),
// 			Name:     "my favorite string",
// 		},
// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{
// 					{
// 						Provider: model.AzureIoTHub,
// 						Credentials: model.Credentials{
// 							Type: "connection_string",
// 							Creds: &model.ConnectionString{
// 								HostName: "localhost",
// 								Key:      []byte("secret"),
// 								Name:     "foobar",
// 							},
// 						},
// 					},
// 				}, nil)
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			hub := new(miothub.Client)
// 			hub.On("GetDevice", contextMatcher, self.ConnStr, self.DeviceID).
// 				Return(nil, errors.New("internal error"))
// 			return hub
// 		},

// 		Error: errors.New("failed to retrieve device from IoT Hub: internal error"),
// 	}, {
// 		Name: "error/no connection string",

// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Status: StatusDisabled,
// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{
// 					{
// 						Provider: model.AzureIoTHub,
// 						Credentials: model.Credentials{
// 							Type:  "connection_string",
// 							Creds: &model.ConnectionString{},
// 						},
// 					},
// 				}, nil)
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			return new(miothub.Client)
// 		},

// 		Error: ErrNoConnectionString,
// 	}, {
// 		Name: "error/getting settings",

// 		DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

// 		Status: StatusDisabled,
// 		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
// 			store := new(storeMocks.DataStore)
// 			store.On("GetDeviceIntegrations", contextMatcher).
// 				Return([]model.Integration{}, errors.New("wut?"))
// 			return store
// 		},
// 		Hub: func(t *testing.T, self *testCase) *miothub.Client {
// 			return new(miothub.Client)
// 		},

// 		Error: errors.New("failed to retrieve settings: wut?"),
// 	}}
// 	for i := range testCases {
// 		tc := testCases[i]
// 		t.Run(tc.Name, func(t *testing.T) {
// 			t.Parallel()
// 			ctx := context.Background()
// 			ds := tc.Store(t, &tc)
// 			hub := tc.Hub(t, &tc)
// 			defer ds.AssertExpectations(t)
// 			defer hub.AssertExpectations(t)

// 			app := New(ds, hub, nil)
// 			err := app.SetDeviceStatus(ctx, tc.DeviceID, tc.Status)

// 			if tc.Error != nil {
// 				if assert.Error(t, err) {
// 					assert.Regexp(t, tc.Error.Error(), err.Error())
// 				}
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }
