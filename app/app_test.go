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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/iot-manager/client/iothub"
	miothub "github.com/mendersoftware/iot-manager/client/iothub/mocks"
	mworkflows "github.com/mendersoftware/iot-manager/client/workflows/mocks"
	"github.com/mendersoftware/iot-manager/model"
	"github.com/mendersoftware/iot-manager/store"
	storeMocks "github.com/mendersoftware/iot-manager/store/mocks"
)

var (
	contextMatcher  = mock.MatchedBy(func(ctx context.Context) bool { return true })
	validConnString = &model.ConnectionString{
		HostName: "localhost:8080",
		Key:      []byte("not-so-secret-key"),
		Name:     "foobar",
	}
)

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

func TestGetIntegrations(t *testing.T) {
	t.Parallel()
	integrationID := uuid.New()
	type testCase struct {
		Name     string
		Store    func(t *testing.T, self *testCase) *storeMocks.DataStore
		Expected []model.Integration

		// Credentials model.Credentials
		Error error
	}

	testCases := []testCase{
		{
			Name: "ok",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("GetIntegrations", contextMatcher, mock.AnythingOfType("model.IntegrationFilter")).
					Return([]model.Integration{{
						ID:          integrationID,
						Provider:    model.ProviderIoTHub,
						Credentials: model.Credentials{},
					}}, nil)
				return ds
			},
			Expected: []model.Integration{{
				ID:          integrationID,
				Provider:    model.ProviderIoTHub,
				Credentials: model.Credentials{},
			}},
		},
		{
			Name: "error: object not found",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("GetIntegrations", contextMatcher, mock.AnythingOfType("model.IntegrationFilter")).
					Return(nil, errors.New("store error: error retrieving integrations collection results"))
				return ds
			},
			Error: errors.New("store error: error retrieving integrations collection results"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			app := New(tc.Store(t, &tc), nil, nil)

			ctx := context.Background()
			res, err := app.GetIntegrations(ctx)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.Equal(t, tc.Expected, res)
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetIntegrationByID(t *testing.T) {
	t.Parallel()

	integrationID := uuid.New()
	type testCase struct {
		Name string

		ID    uuid.UUID
		Store func(t *testing.T, self *testCase) *storeMocks.DataStore

		Integration *model.Integration
		Error       error
	}
	testCases := []testCase{{
		Name: "ok",

		ID: integrationID,
		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			ds := new(storeMocks.DataStore)
			ds.On("GetIntegrationById", contextMatcher, self.ID).
				Return(self.Integration, nil)
			return ds
		},

		Integration: &model.Integration{
			ID:       integrationID,
			Provider: model.ProviderIoTHub,
			Credentials: model.Credentials{
				Type:             model.CredentialTypeSAS,
				ConnectionString: validConnString,
			},
		},
	}, {
		Name: "error/not found",

		ID: integrationID,
		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			ds := new(storeMocks.DataStore)
			ds.On("GetIntegrationById", contextMatcher, self.ID).
				Return(nil, store.ErrObjectNotFound)
			return ds
		},

		Error: ErrIntegrationNotFound,
	}, {
		Name: "error/internal",

		ID: integrationID,
		Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
			ds := new(storeMocks.DataStore)
			ds.On("GetIntegrationById", contextMatcher, self.ID).
				Return(nil, errors.New("internal error"))
			return ds
		},

		Error: errors.New("internal error"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			store := tc.Store(t, &tc)
			app := New(store, nil, nil)
			integration, err := app.GetIntegrationById(context.Background(), tc.ID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Integration, integration)
			}
		})
	}
}

func TestCreateIntegration(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name                  string
		Store                 func(t *testing.T, self *testCase) *storeMocks.DataStore
		CreateIntegrationData model.Integration
		Error                 error
	}

	testCases := []testCase{
		{
			Name: "integration created",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("CreateIntegration", contextMatcher, mock.AnythingOfType("model.Integration")).
					Return(&self.CreateIntegrationData, nil)
				return ds
			},
			CreateIntegrationData: model.Integration{
				Provider: model.ProviderIoTHub,
				Credentials: model.Credentials{
					Type: "connection_string",
					ConnectionString: &model.ConnectionString{
						HostName: "localhost",
						Key:      []byte("secret"),
						Name:     "foobar",
					},
				},
			},
		},
		{
			Name: "create integration error",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("CreateIntegration", contextMatcher, mock.AnythingOfType("model.Integration")).
					Return(nil, errors.New("error creating the integration"))
				return ds
			},
			Error: errors.New("error creating the integration"),
		},
		{
			Name: "error: integration already exists",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("CreateIntegration", contextMatcher, mock.AnythingOfType("model.Integration")).
					Return(nil, store.ErrObjectExists)
				return ds
			},
			Error: ErrIntegrationExists,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			app := New(tc.Store(t, &tc), nil, nil)

			ctx := context.Background()
			_, err := app.CreateIntegration(ctx, tc.CreateIntegrationData)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetIntegrationCredentials(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("integration"))
	type testCase struct {
		Name        string
		Store       func(t *testing.T, self *testCase) *storeMocks.DataStore
		Credentials model.Credentials
		Error       error
	}

	testCases := []testCase{
		{
			Name: "ok",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("SetIntegrationCredentials", contextMatcher, integrationID, mock.AnythingOfType("model.Credentials")).
					Return(nil)
				return ds
			},
			Credentials: model.Credentials{},
		},
		{
			Name: "error: object not found",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("SetIntegrationCredentials", contextMatcher, integrationID, mock.AnythingOfType("model.Credentials")).
					Return(store.ErrObjectNotFound)
				return ds
			},
			Error: ErrIntegrationNotFound,
		},
		{
			Name: "error: unexpected error",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("SetIntegrationCredentials", contextMatcher, integrationID, mock.AnythingOfType("model.Credentials")).
					Return(errors.New("unexpected error"))
				return ds
			},
			Error: errors.New("unexpected error"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			app := New(tc.Store(t, &tc), nil, nil)

			ctx := context.Background()
			err := app.SetIntegrationCredentials(ctx, integrationID, tc.Credentials)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRemoveIntegration(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("integration"))
	type testCase struct {
		Name  string
		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
		Error error
	}

	testCases := []testCase{
		{
			Name: "ok",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("DoDevicesExistByIntegrationID", contextMatcher, integrationID).
					Return(false, nil)
				ds.On("RemoveIntegration", contextMatcher, integrationID).
					Return(nil)
				return ds
			},
		},
		{
			Name: "error: get devices by integration ID issue",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("DoDevicesExistByIntegrationID", contextMatcher, integrationID).
					Return(false, errors.New("some error: error retrieving integration collection results"))
				return ds
			},
			Error: errors.New("some error: error retrieving integration collection results"),
		},
		{
			Name: "error: devices with given integration ID exist",
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("DoDevicesExistByIntegrationID", contextMatcher, integrationID).
					Return(true, nil)
				return ds
			},
			Error: ErrCannotRemoveIntegration,
		},
		{
			Name: "error: integration not found",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("DoDevicesExistByIntegrationID", contextMatcher, integrationID).
					Return(false, nil)
				ds.On("RemoveIntegration", contextMatcher, integrationID).
					Return(store.ErrObjectNotFound)
				return ds
			},
			Error: ErrIntegrationNotFound,
		},
		{
			Name: "error: unexpected error",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				ds := new(storeMocks.DataStore)
				ds.On("DoDevicesExistByIntegrationID", contextMatcher, integrationID).
					Return(false, nil)
				ds.On("RemoveIntegration", contextMatcher, integrationID).
					Return(errors.New("unexpected mongo error"))
				return ds
			},
			Error: errors.New("unexpected mongo error"),
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			app := New(tc.Store(t, &tc), nil, nil)

			ctx := context.Background()
			err := app.RemoveIntegration(ctx, integrationID)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProvisionDevice(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	connString := &model.ConnectionString{
		HostName: "localhost",
		Key:      []byte("secret"),
		Name:     "foobar",
	}
	type testCase struct {
		Name     string
		DeviceID string

		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
		Hub   func(t *testing.T, self *testCase) *miothub.Client
		Wf    func(t *testing.T, self *testCase) *mworkflows.Client

		Error error
	}
	testCases := []testCase{
		{
			Name:     "ok",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{}).
					Return([]model.Integration{
						{
							ID:       integrationID,
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type:             "connection_string",
								ConnectionString: connString,
							},
						},
					}, nil)
				store.On("UpsertDeviceIntegrations", contextMatcher, self.DeviceID, []uuid.UUID{integrationID}).
					Return(&model.Device{
						ID:             self.DeviceID,
						IntegrationIDs: []uuid.UUID{integrationID},
					}, nil)
				return store
			},
			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				hub.On("UpsertDevice", contextMatcher, connString, self.DeviceID).
					Return(&iothub.Device{
						DeviceID: self.DeviceID,
						Auth: &iothub.Auth{
							Type: iothub.AuthTypeSymmetric,
							SymmetricKey: &iothub.SymmetricKey{
								Primary:   iothub.Key("key1"),
								Secondary: iothub.Key("key2"),
							},
						},
					}, nil).
					On("UpdateDeviceTwin", contextMatcher, connString, self.DeviceID,
						&iothub.DeviceTwinUpdate{
							Tags: map[string]interface{}{"mender": true},
						}).
					Return(nil)
				return hub
			},
			Wf: func(t *testing.T, self *testCase) *mworkflows.Client {
				wf := new(mworkflows.Client)
				primKey := &model.ConnectionString{
					Key:      []byte("key1"),
					DeviceID: self.DeviceID,
					HostName: connString.HostName,
				}
				wf.On("ProvisionExternalDevice",
					contextMatcher,
					self.DeviceID,
					map[string]string{
						confKeyPrimaryKey: primKey.String(),
					}).Return(nil)
				return wf
			},
		},
		{
			Name:     "error/device does not have a connection string",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{}).
					Return([]model.Integration{
						{
							ID:       integrationID,
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type:             "connection_string",
								ConnectionString: connString,
							},
						},
					}, nil)
				return store
			},
			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				hub.On("UpsertDevice", contextMatcher, connString, self.DeviceID).
					Return(&iothub.Device{
						DeviceID: self.DeviceID,
						Auth: &iothub.Auth{
							Type: iothub.AuthTypeNone,
						},
					}, nil)
				return hub
			},
			Wf: func(t *testing.T, self *testCase) *mworkflows.Client {
				return new(mworkflows.Client)
			},

			Error: ErrNoDeviceConnectionString,
		},
		{
			Name:     "error/hub failure",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{}).
					Return([]model.Integration{
						{
							ID:       integrationID,
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type:             "connection_string",
								ConnectionString: connString,
							},
						},
					}, nil)
				return store
			},
			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				hub.On("UpsertDevice", contextMatcher, connString, self.DeviceID).
					Return(nil, errors.New("internal error"))
				return hub
			},
			Wf: func(t *testing.T, self *testCase) *mworkflows.Client {
				return new(mworkflows.Client)
			},

			Error: errors.New("failed to update iothub devices: internal error"),
		},
		{
			Name: "error/getting integrations",

			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{}).
					Return([]model.Integration{}, errors.New("wut?"))
				return store
			},
			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				return new(miothub.Client)
			},
			Wf: func(t *testing.T, self *testCase) *mworkflows.Client {
				return new(mworkflows.Client)
			},

			Error: errors.New("failed to retrieve integrations: wut?"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := tc.Store(t, &tc)
			hub := tc.Hub(t, &tc)
			wf := tc.Wf(t, &tc)
			defer ds.AssertExpectations(t)
			defer hub.AssertExpectations(t)
			defer wf.AssertExpectations(t)

			app := New(ds, hub, wf)
			err := app.ProvisionDevice(ctx, tc.DeviceID)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteIOTHubDevice(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	connString := &model.ConnectionString{
		HostName: "localhost",
		Key:      []byte("secret"),
		Name:     "foobar",
	}
	type testCase struct {
		Name     string
		DeviceID string

		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
		Hub   func(t *testing.T, self *testCase) *miothub.Client

		Error error
	}
	testCases := []testCase{
		{
			Name:     "error: device not found in db in GetDeviceIntegrations",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				return hub
			},
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				mockedStore := new(storeMocks.DataStore)
				mockedStore.On("GetDevice", contextMatcher, self.DeviceID).
					Return(nil, store.ErrObjectNotFound)
				return mockedStore
			},
			Error: ErrDeviceNotFound,
		},
		{
			Name:     "ok",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				hub.On("DeleteDevice", contextMatcher, connString, self.DeviceID).
					Return(nil)
				return hub
			},
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetDevice", contextMatcher, self.DeviceID).
					Return(&model.Device{
						ID:             self.DeviceID,
						IntegrationIDs: []uuid.UUID{integrationID},
					},
						nil)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{IDs: []uuid.UUID{integrationID}}).
					Return([]model.Integration{
						{
							ID:       integrationID,
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type:             model.CredentialTypeSAS,
								ConnectionString: connString,
							},
						},
					}, nil)
				store.On("DeleteDevice", contextMatcher, self.DeviceID).
					Return(nil)
				return store
			},
		},
		{
			Name:     "error: no connection string",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				return hub
			},
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetDevice", contextMatcher, self.DeviceID).
					Return(&model.Device{
						ID:             self.DeviceID,
						IntegrationIDs: []uuid.UUID{integrationID},
					},
						nil)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{IDs: []uuid.UUID{integrationID}}).
					Return([]model.Integration{
						{
							ID:          integrationID,
							Provider:    model.ProviderIoTHub,
							Credentials: model.Credentials{},
						},
					}, nil)
				return store
			},
			Error: ErrNoConnectionString,
		},
		{
			Name:     "error: device not found",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				hub.On("DeleteDevice", contextMatcher, connString, self.DeviceID).
					Return(nil)
				return hub
			},
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				mockedStore := new(storeMocks.DataStore)
				mockedStore.On("GetDevice", contextMatcher, self.DeviceID).
					Return(&model.Device{
						ID:             self.DeviceID,
						IntegrationIDs: []uuid.UUID{integrationID},
					},
						nil)
				mockedStore.On("GetIntegrations", contextMatcher, model.IntegrationFilter{IDs: []uuid.UUID{integrationID}}).
					Return([]model.Integration{
						{
							ID:       integrationID,
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type:             model.CredentialTypeSAS,
								ConnectionString: connString,
							},
						},
					}, nil)
				mockedStore.On("DeleteDevice", contextMatcher, self.DeviceID).
					Return(store.ErrObjectNotFound)
				return mockedStore
			},
			Error: ErrDeviceNotFound,
		},
		{
			Name:     "error: hub failure",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				hub.On("DeleteDevice", contextMatcher, connString, self.DeviceID).
					Return(errors.New("failed to delete IoT Hub device: store: unexpected error"))
				return hub
			},
			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetDevice", contextMatcher, self.DeviceID).
					Return(&model.Device{
						ID:             self.DeviceID,
						IntegrationIDs: []uuid.UUID{integrationID},
					},
						nil)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{IDs: []uuid.UUID{integrationID}}).
					Return([]model.Integration{
						{
							ID:       integrationID,
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type:             model.CredentialTypeSAS,
								ConnectionString: connString,
							},
						},
					}, nil)
				return store
			},
			Error: errors.New("failed to delete IoT Hub device: store: unexpected error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := tc.Store(t, &tc)
			hub := tc.Hub(t, &tc)
			defer ds.AssertExpectations(t)
			defer hub.AssertExpectations(t)

			app := New(ds, hub, nil)
			err := app.DeleteIOTHubDevice(ctx, tc.DeviceID)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetDeviceStatus(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	connString := &model.ConnectionString{
		HostName: "localhost",
		Key:      []byte("secret"),
		Name:     "foobar",
	}
	type testCase struct {
		Name string

		ConnStr  *model.ConnectionString
		DeviceID string
		Status   model.Status

		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
		Hub   func(t *testing.T, self *testCase) *miothub.Client

		Error error
	}
	testCases := []testCase{
		{
			Name: "ok",

			Status:   model.StatusAccepted,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetDevice", contextMatcher, self.DeviceID).
					Return(&model.Device{
						ID:             self.DeviceID,
						IntegrationIDs: []uuid.UUID{integrationID},
					},
						nil)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{IDs: []uuid.UUID{integrationID}}).
					Return([]model.Integration{
						{
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type:             "connection_string",
								ConnectionString: connString,
							},
						},
					}, nil)
				return store
			},
			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				dev := &iothub.Device{
					DeviceID: "foobar",
					Status:   iothub.StatusDisabled,
				}
				hub.On("GetDevice", contextMatcher, connString, self.DeviceID).
					Return(dev, nil).
					On("UpsertDevice", contextMatcher, connString, self.DeviceID,
						mock.MatchedBy(func(dev *iothub.Device) bool {
							return dev.Status == iothub.StatusEnabled
						})).
					Return(dev, nil)
				return hub
			},
		},
		{
			Name: "ok, device already has matching status",

			Status:   model.StatusDecommissioned,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetDevice", contextMatcher, self.DeviceID).
					Return(&model.Device{
						ID:             self.DeviceID,
						IntegrationIDs: []uuid.UUID{integrationID},
					},
						nil)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{IDs: []uuid.UUID{integrationID}}).
					Return([]model.Integration{
						{
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type:             "connection_string",
								ConnectionString: connString,
							},
						},
					}, nil)
				return store
			},
			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				dev := &iothub.Device{
					DeviceID: self.DeviceID,
					Status:   iothub.StatusDisabled,
				}
				hub.On("GetDevice", contextMatcher, connString, self.DeviceID).
					Return(dev, nil)
				return hub
			},
		},
		{
			Name: "error/hub fail to update device",

			Status:   model.StatusAccepted,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetDevice", contextMatcher, self.DeviceID).
					Return(&model.Device{
						ID:             self.DeviceID,
						IntegrationIDs: []uuid.UUID{integrationID},
					},
						nil)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{IDs: []uuid.UUID{integrationID}}).
					Return([]model.Integration{
						{
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type:             "connection_string",
								ConnectionString: connString,
							},
						},
					}, nil)
				return store
			},
			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				dev := &iothub.Device{
					DeviceID: self.DeviceID,
					Status:   iothub.StatusDisabled,
				}
				hub.On("GetDevice", contextMatcher, connString, self.DeviceID).
					Return(dev, nil).
					On("UpsertDevice", contextMatcher, connString, self.DeviceID,
						mock.MatchedBy(func(dev *iothub.Device) bool {
							return dev.Status == iothub.StatusEnabled
						})).
					Return(nil, errors.New("failed to update IoT Hub device: hub: unexpected error"))
				return hub
			},
			Error: errors.New("failed to update IoT Hub device: hub: unexpected error"),
		},
		{
			Name: "ok",

			Status:   model.StatusAccepted,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetDevice", contextMatcher, self.DeviceID).
					Return(&model.Device{
						ID:             self.DeviceID,
						IntegrationIDs: []uuid.UUID{integrationID},
					},
						nil)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{IDs: []uuid.UUID{integrationID}}).
					Return([]model.Integration{
						{
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type:             "connection_string",
								ConnectionString: connString,
							},
						},
					}, nil)
				return store
			},
			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				hub.On("GetDevice", contextMatcher, connString, self.DeviceID).
					Return(nil, errors.New("failed to retrieve device from IoT Hub: hub: unexpected error"))
				return hub
			},
			Error: errors.New("failed to retrieve device from IoT Hub: hub: unexpected error"),
		},
		{
			Name: "error/no connection string",

			Status:   model.StatusAccepted,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetDevice", contextMatcher, self.DeviceID).
					Return(&model.Device{
						ID:             self.DeviceID,
						IntegrationIDs: []uuid.UUID{integrationID},
					},
						nil)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{IDs: []uuid.UUID{integrationID}}).
					Return([]model.Integration{
						{
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type: model.CredentialTypeSAS,
							},
						},
					}, nil)
				return store
			},
			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				return hub
			},
			Error: ErrNoConnectionString,
		},
		{
			Name: "error/getting settings",

			Status:   model.StatusPreauthorized,
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetDevice", contextMatcher, self.DeviceID).
					Return(&model.Device{
						ID:             self.DeviceID,
						IntegrationIDs: []uuid.UUID{integrationID},
					},
						nil)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{IDs: []uuid.UUID{integrationID}}).
					Return(nil, errors.New("failed to retrieve device integrations: unexpected error"))
				return store
			},
			Hub: func(t *testing.T, self *testCase) *miothub.Client {
				hub := new(miothub.Client)
				return hub
			},
			Error: errors.New("failed to retrieve device integrations: unexpected error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := tc.Store(t, &tc)
			hub := tc.Hub(t, &tc)
			defer ds.AssertExpectations(t)
			defer hub.AssertExpectations(t)

			app := New(ds, hub, nil)
			err := app.SetDeviceStatus(ctx, tc.DeviceID, tc.Status)

			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDevice(t *testing.T) {
	testCases := []struct {
		Name string

		DeviceID       string
		GetDevice      *model.Device
		GetDeviceError error
		Error          error
	}{
		{
			Name: "ok",

			DeviceID: "1",
			GetDevice: &model.Device{
				ID: "1",
			},
		},
		{
			Name: "ok, device doesn't exist",

			DeviceID:       "1",
			GetDeviceError: store.ErrObjectNotFound,
			Error:          ErrDeviceNotFound,
		},
		{
			Name: "ko, device retrieval error",

			DeviceID:       "1",
			GetDeviceError: errors.New("error getting the device"),
			Error:          errors.New("error getting the device"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("GetDevice",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
				tc.DeviceID,
			).Return(tc.GetDevice, tc.GetDeviceError)
			app := New(store, nil, nil)

			ctx := context.Background()
			device, err := app.GetDevice(ctx, tc.DeviceID)
			if tc.Error != nil {
				assert.EqualError(t, err, tc.Error.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.GetDevice, device)
			}
		})
	}
}

func TestGetDeviceStateIntegration(t *testing.T) {
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	testCases := []struct {
		Name string

		DeviceID      string
		IntegrationID uuid.UUID

		GetDeviceByIntegrationID       *model.Device
		GetDeviceByIntegrationIDError  error
		GetIntegration                 *model.Integration
		GetIntegrationError            error
		GetDeviceStateIntegration      *model.DeviceState
		GetDeviceStateIntegrationError error
	}{
		{
			Name: "ko, device not found",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationIDError:  store.ErrObjectNotFound,
			GetDeviceStateIntegrationError: ErrIntegrationNotFound,
		},
		{
			Name: "ko, failed retrieving the device not found",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationIDError:  errors.New("internal error"),
			GetDeviceStateIntegrationError: errors.New("failed to retrieve the device: internal error"),
		},
		{
			Name: "ko, integration not found",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID:       &model.Device{},
			GetDeviceStateIntegrationError: ErrIntegrationNotFound,
		},
		{
			Name: "ko, failed retrieving the integration",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID:       &model.Device{},
			GetIntegrationError:            errors.New("internal error"),
			GetDeviceStateIntegrationError: errors.New("failed to retrieve the integration: internal error"),
		},
		{
			Name: "ko, unknown integration",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID: &model.Device{},
			GetIntegration: &model.Integration{
				Provider: model.Provider("super-secret-provider"),
			},
			GetDeviceStateIntegrationError: ErrUnknownIntegration,
		},
		{
			Name: "ko, no connection string",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID: &model.Device{},
			GetIntegration: &model.Integration{
				Provider: model.ProviderIoTHub,
			},
			GetDeviceStateIntegrationError: ErrNoConnectionString,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("GetDeviceByIntegrationID",
				contextMatcher,
				tc.DeviceID,
				tc.IntegrationID,
			).Return(
				tc.GetDeviceByIntegrationID,
				tc.GetDeviceByIntegrationIDError,
			)
			if tc.GetDeviceByIntegrationIDError == nil {
				store.On("GetIntegrationById",
					contextMatcher,
					tc.IntegrationID,
				).Return(
					tc.GetIntegration,
					tc.GetIntegrationError,
				)
			}
			app := New(store, nil, nil)

			ctx := context.Background()
			state, err := app.GetDeviceStateIntegration(ctx, tc.DeviceID, tc.IntegrationID)
			if tc.GetDeviceStateIntegrationError != nil {
				assert.EqualError(t, err, tc.GetDeviceStateIntegrationError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.GetDeviceStateIntegration, state)
			}
		})
	}
}

func TestSetDeviceStateIntegration(t *testing.T) {
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	testCases := []struct {
		Name string

		DeviceID      string
		IntegrationID uuid.UUID

		GetDeviceByIntegrationID       *model.Device
		GetDeviceByIntegrationIDError  error
		GetIntegration                 *model.Integration
		GetIntegrationError            error
		SetDeviceStateIntegration      *model.DeviceState
		SetDeviceStateIntegrationError error
	}{
		{
			Name: "ko, device not found",

			DeviceID:      "1",
			IntegrationID: integrationID,

			SetDeviceStateIntegrationError: ErrIntegrationNotFound,
		},
		{
			Name: "ko, failed retrieving the device not found",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationIDError:  errors.New("internal error"),
			SetDeviceStateIntegrationError: errors.New("failed to retrieve the device: internal error"),
		},
		{
			Name: "ko, integration not found",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID:       &model.Device{},
			SetDeviceStateIntegrationError: ErrIntegrationNotFound,
		},
		{
			Name: "ko, failed retrieving the integration",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID:       &model.Device{},
			GetIntegrationError:            errors.New("internal error"),
			SetDeviceStateIntegrationError: errors.New("failed to retrieve the integration: internal error"),
		},
		{
			Name: "ko, unknown integration",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID: &model.Device{},
			GetIntegration: &model.Integration{
				Provider: model.Provider("super-secret-provider"),
			},
			SetDeviceStateIntegrationError: ErrUnknownIntegration,
		},
		{
			Name: "ko, no connection string",

			DeviceID:      "1",
			IntegrationID: integrationID,

			GetDeviceByIntegrationID: &model.Device{},
			GetIntegration: &model.Integration{
				Provider: model.ProviderIoTHub,
			},
			SetDeviceStateIntegrationError: ErrNoConnectionString,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("GetDeviceByIntegrationID",
				contextMatcher,
				tc.DeviceID,
				tc.IntegrationID,
			).Return(
				tc.GetDeviceByIntegrationID,
				tc.GetDeviceByIntegrationIDError,
			)
			if tc.GetDeviceByIntegrationID != nil && tc.GetDeviceByIntegrationIDError == nil {
				store.On("GetIntegrationById",
					contextMatcher,
					tc.IntegrationID,
				).Return(
					tc.GetIntegration,
					tc.GetIntegrationError,
				)
			}
			app := New(store, nil, nil)

			ctx := context.Background()
			state := &model.DeviceState{}
			state, err := app.SetDeviceStateIntegration(ctx, tc.DeviceID, tc.IntegrationID, state)
			if tc.SetDeviceStateIntegrationError != nil {
				assert.EqualError(t, err, tc.SetDeviceStateIntegrationError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.SetDeviceStateIntegration, state)
			}
		})
	}
}
