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

package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/iot-manager/client/iotcore"
	coreMocks "github.com/mendersoftware/iot-manager/client/iotcore/mocks"
	wfMocks "github.com/mendersoftware/iot-manager/client/workflows/mocks"
	"github.com/mendersoftware/iot-manager/crypto"
	"github.com/mendersoftware/iot-manager/model"
	"github.com/mendersoftware/iot-manager/store"
	storeMocks "github.com/mendersoftware/iot-manager/store/mocks"
)

var (
	awsAccessKeyID     = "dummy"
	awsSecretAccessKey = crypto.String("dummy")
	awsEndpoint        = "random-id.iot.us-east-1.amazonaws.com"
	awsPolicyDocument  = "{}"
)

func TestProvisionDeviceIoTCore(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	type testCase struct {
		Name     string
		DeviceID string

		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
		Core  func(t *testing.T, self *testCase) *coreMocks.Client
		Wf    func(t *testing.T, self *testCase) *wfMocks.Client

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
							Provider: model.ProviderIoTCore,
							Credentials: model.Credentials{
								Type:                 model.CredentialTypeAWS,
								AccessKeyID:          &awsAccessKeyID,
								SecretAccessKey:      &awsSecretAccessKey,
								EndpointURL:          &awsEndpoint,
								DevicePolicyDocument: &awsPolicyDocument,
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
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("UpsertDevice",
					contextMatcher,
					mock.AnythingOfType("*session.Session"),
					self.DeviceID,
					&iotcore.Device{
						Status: iotcore.StatusDisabled,
					},
					awsPolicyDocument).
					Return(&iotcore.Device{
						ID:          self.DeviceID,
						PrivateKey:  "private_key",
						Certificate: "certificate",
					}, nil)
				return core
			},
			Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
				wf := new(wfMocks.Client)
				wf.On("ProvisionExternalDevice",
					contextMatcher,
					self.DeviceID,
					map[string]string{
						confKeyAWSPrivateKey:  "private_key",
						confKeyAWSCertificate: "certificate",
					}).Return(nil)
				return wf
			},
		},
		{
			Name:     "error, no credentials",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{}).
					Return([]model.Integration{
						{
							ID:       integrationID,
							Provider: model.ProviderIoTCore,
							Credentials: model.Credentials{
								Type: model.CredentialTypeAWS,
							},
						},
					}, nil)
				return store
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				return core
			},
			Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
				return new(wfMocks.Client)
			},

			Error: ErrNoCredentials,
		},
		{
			Name:     "error, failure",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{}).
					Return([]model.Integration{
						{
							ID:       integrationID,
							Provider: model.ProviderIoTCore,
							Credentials: model.Credentials{
								Type:                 model.CredentialTypeAWS,
								AccessKeyID:          &awsAccessKeyID,
								SecretAccessKey:      &awsSecretAccessKey,
								EndpointURL:          &awsEndpoint,
								DevicePolicyDocument: &awsPolicyDocument,
							},
						},
					}, nil)
				return store
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("UpsertDevice",
					contextMatcher,
					mock.AnythingOfType("*session.Session"),
					self.DeviceID,
					&iotcore.Device{
						Status: iotcore.StatusDisabled,
					},
					awsPolicyDocument).
					Return(nil, errors.New("internal error"))
				return core
			},
			Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
				return new(wfMocks.Client)
			},

			Error: errors.New("failed to update iotcore devices: internal error"),
		},
		{
			Name:     "ok",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{}).
					Return([]model.Integration{
						{
							ID:       integrationID,
							Provider: model.ProviderIoTCore,
							Credentials: model.Credentials{
								Type:                 model.CredentialTypeAWS,
								AccessKeyID:          &awsAccessKeyID,
								SecretAccessKey:      &awsSecretAccessKey,
								EndpointURL:          &awsEndpoint,
								DevicePolicyDocument: &awsPolicyDocument,
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
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("UpsertDevice",
					contextMatcher,
					mock.AnythingOfType("*session.Session"),
					self.DeviceID,
					&iotcore.Device{
						Status: iotcore.StatusDisabled,
					},
					awsPolicyDocument).
					Return(&iotcore.Device{
						ID:          self.DeviceID,
						PrivateKey:  "private_key",
						Certificate: "certificate",
					}, nil)
				return core
			},
			Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
				wf := new(wfMocks.Client)
				wf.On("ProvisionExternalDevice",
					contextMatcher,
					self.DeviceID,
					map[string]string{
						confKeyAWSPrivateKey:  "private_key",
						confKeyAWSCertificate: "certificate",
					}).Return(nil)
				return wf
			},
		},
		{
			Name:     "error, deviceconfig",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Store: func(t *testing.T, self *testCase) *storeMocks.DataStore {
				store := new(storeMocks.DataStore)
				store.On("GetIntegrations", contextMatcher, model.IntegrationFilter{}).
					Return([]model.Integration{
						{
							ID:       integrationID,
							Provider: model.ProviderIoTCore,
							Credentials: model.Credentials{
								Type:                 model.CredentialTypeAWS,
								AccessKeyID:          &awsAccessKeyID,
								SecretAccessKey:      &awsSecretAccessKey,
								EndpointURL:          &awsEndpoint,
								DevicePolicyDocument: &awsPolicyDocument,
							},
						},
					}, nil)
				return store
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("UpsertDevice",
					contextMatcher,
					mock.AnythingOfType("*session.Session"),
					self.DeviceID,
					&iotcore.Device{
						Status: iotcore.StatusDisabled,
					},
					awsPolicyDocument).
					Return(&iotcore.Device{
						ID:          self.DeviceID,
						PrivateKey:  "private_key",
						Certificate: "certificate",
					}, nil)
				return core
			},
			Wf: func(t *testing.T, self *testCase) *wfMocks.Client {
				wf := new(wfMocks.Client)
				wf.On("ProvisionExternalDevice",
					contextMatcher,
					self.DeviceID,
					map[string]string{
						confKeyAWSPrivateKey:  "private_key",
						confKeyAWSCertificate: "certificate",
					}).Return(errors.New("internal error"))
				return wf
			},
			Error: errors.New("failed to submit iotcore credentials to deviceconfig: internal error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			ds := tc.Store(t, &tc)
			defer ds.AssertExpectations(t)

			wf := tc.Wf(t, &tc)
			defer wf.AssertExpectations(t)

			app := New(ds, wf, nil)

			core := tc.Core(t, &tc)
			defer core.AssertExpectations(t)
			app = app.WithIoTCore(core)

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

func TestDecommissionDeviceIoTCore(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	type testCase struct {
		Name     string
		DeviceID string

		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
		Core  func(t *testing.T, self *testCase) *coreMocks.Client

		Error error
	}
	testCases := []testCase{
		{
			Name:     "ok, iot core",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("DeleteDevice", contextMatcher, mock.AnythingOfType("*session.Session"), self.DeviceID).
					Return(nil)
				return core
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
							Provider: model.ProviderIoTCore,
							Credentials: model.Credentials{
								Type:                 model.CredentialTypeAWS,
								AccessKeyID:          &awsAccessKeyID,
								SecretAccessKey:      &awsSecretAccessKey,
								EndpointURL:          &awsEndpoint,
								DevicePolicyDocument: &awsPolicyDocument,
							},
						},
					}, nil)
				store.On("DeleteDevice", contextMatcher, self.DeviceID).
					Return(nil)
				return store
			},
		},
		{
			Name:     "error, no credentials",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				return core
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
							Provider: model.ProviderIoTCore,
							Credentials: model.Credentials{
								Type: model.CredentialTypeAWS,
							},
						},
					}, nil)
				return store
			},
			Error: ErrNoCredentials,
		},
		{
			Name:     "error, device not found",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("DeleteDevice", contextMatcher, mock.AnythingOfType("*session.Session"), self.DeviceID).
					Return(nil)
				return core
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
							Provider: model.ProviderIoTCore,
							Credentials: model.Credentials{
								Type:                 model.CredentialTypeAWS,
								AccessKeyID:          &awsAccessKeyID,
								SecretAccessKey:      &awsSecretAccessKey,
								EndpointURL:          &awsEndpoint,
								DevicePolicyDocument: &awsPolicyDocument,
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
			Name:     "error, failure",
			DeviceID: "68ac6f41-c2e7-429f-a4bd-852fac9a5045",

			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("DeleteDevice", contextMatcher, mock.AnythingOfType("*session.Session"), self.DeviceID).
					Return(errors.New("failed to delete IoT Core device: store: unexpected error"))
				return core
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
							Provider: model.ProviderIoTCore,
							Credentials: model.Credentials{
								Type:                 model.CredentialTypeAWS,
								AccessKeyID:          &awsAccessKeyID,
								SecretAccessKey:      &awsSecretAccessKey,
								EndpointURL:          &awsEndpoint,
								DevicePolicyDocument: &awsPolicyDocument,
							},
						},
					}, nil)
				return store
			},
			Error: errors.New("failed to delete IoT Core device: store: unexpected error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := tc.Store(t, &tc)
			defer ds.AssertExpectations(t)

			app := New(ds, nil, nil)

			core := tc.Core(t, &tc)
			defer core.AssertExpectations(t)
			app = app.WithIoTCore(core)

			err := app.DecommissionDevice(ctx, tc.DeviceID)

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

func TestSetDeviceStatusIoTCore(t *testing.T) {
	t.Parallel()
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	type testCase struct {
		Name string

		ConnStr  *model.ConnectionString
		DeviceID string
		Status   model.Status

		Store func(t *testing.T, self *testCase) *storeMocks.DataStore
		Core  func(t *testing.T, self *testCase) *coreMocks.Client

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
							ID:       integrationID,
							Provider: model.ProviderIoTCore,
							Credentials: model.Credentials{
								Type:                 model.CredentialTypeAWS,
								AccessKeyID:          &awsAccessKeyID,
								SecretAccessKey:      &awsSecretAccessKey,
								EndpointURL:          &awsEndpoint,
								DevicePolicyDocument: &awsPolicyDocument,
							},
						},
					}, nil)
				return store
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				dev := &iotcore.Device{
					ID:     "foobar",
					Status: iotcore.StatusDisabled,
				}
				core.On("UpsertDevice", contextMatcher, mock.AnythingOfType("*session.Session"), self.DeviceID,
					mock.MatchedBy(func(dev *iotcore.Device) bool {
						return dev.Status == iotcore.StatusEnabled
					}), awsPolicyDocument).
					Return(dev, nil)
				return core
			},
		},
		{
			Name: "error, missing credentials",

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
							ID:       integrationID,
							Provider: model.ProviderIoTCore,
							Credentials: model.Credentials{
								Type: model.CredentialTypeAWS,
							},
						},
					}, nil)
				return store
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				return core
			},
			Error: ErrNoCredentials,
		},
		{
			Name: "error, fail to update device",

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
							ID:       integrationID,
							Provider: model.ProviderIoTCore,
							Credentials: model.Credentials{
								Type:                 model.CredentialTypeAWS,
								AccessKeyID:          &awsAccessKeyID,
								SecretAccessKey:      &awsSecretAccessKey,
								EndpointURL:          &awsEndpoint,
								DevicePolicyDocument: &awsPolicyDocument,
							},
						},
					}, nil)
				return store
			},
			Core: func(t *testing.T, self *testCase) *coreMocks.Client {
				core := new(coreMocks.Client)
				core.On("UpsertDevice", contextMatcher, mock.AnythingOfType("*session.Session"), self.DeviceID,
					mock.MatchedBy(func(dev *iotcore.Device) bool {
						return dev.Status == iotcore.StatusEnabled
					}), awsPolicyDocument).
					Return(nil, errors.New("failed to update IoT Hub device: hub: unexpected error"))
				return core
			},
			Error: errors.New("failed to update IoT Hub device: hub: unexpected error"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ds := tc.Store(t, &tc)
			defer ds.AssertExpectations(t)

			app := New(ds, nil, nil)

			core := tc.Core(t, &tc)
			defer core.AssertExpectations(t)
			app = app.WithIoTCore(core)

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
