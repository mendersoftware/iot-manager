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

package mongo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mendersoftware/go-lib-micro/identity"
	mstore "github.com/mendersoftware/go-lib-micro/store/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/iot-manager/model"
	"github.com/mendersoftware/iot-manager/store"
)

func TestCreateIntegration(t *testing.T) {
	testCases := []struct {
		Name string

		CTX         context.Context
		Integration model.Integration

		Error error
	}{{
		Name: "ok",

		CTX: identity.WithContext(context.Background(), &identity.Identity{
			Tenant: "1234567890",
		}),
		Integration: model.Integration{
			ID:       uuid.NewSHA1(uuid.NameSpaceOID, []byte("1234567890")),
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
	}, {
		Name: "ok, no tenant context",

		CTX: context.Background(),
		Integration: model.Integration{
			ID:       uuid.NewSHA1(uuid.NameSpaceOID, []byte("")),
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
	}, {
		Name: "error, context canceled",

		CTX: func() context.Context {
			ctx, cc := context.WithCancel(context.Background())
			cc()
			return ctx
		}(),
		Integration: model.Integration{
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
		Error: context.Canceled,
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			db.Wipe()
			mgo := db.Client()
			ds := NewDataStoreWithClient(mgo)
			err := ds.CreateIntegration(tc.CTX, tc.Integration)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				idty := identity.FromContext(tc.CTX)
				var tenantID string
				if idty != nil {
					tenantID = idty.Tenant
				}
				fltr := bson.D{{
					Key: "tenant_id", Value: tenantID,
				}}

				var doc bson.Raw
				err := mgo.Database(DbName).
					Collection(CollNameIntegrations).
					FindOne(tc.CTX, fltr).
					Decode(&doc)
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				field := doc.Lookup(KeyTenantID)
				actualTID, ok := field.StringValueOK()
				assert.True(t, ok, "bson document does not contain tenant_id field")
				assert.Equal(t, tenantID, actualTID)

				var integration model.Integration
				bson.Unmarshal(doc, &integration)
				assert.Equal(t, tc.Integration, integration)
			}
		})
	}
}

func TestGetIntegrations(t *testing.T) {
	const tenantID = "123456789012345678901234"
	testCases := []struct {
		Name string

		CTX context.Context

		Integrations []model.Integration
		Error        error
	}{
		{
			Name: "ok got integration",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			Integrations: []model.Integration{
				{
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
		},
		{
			Name: "no integrations for tenant",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: "111111111111111111111111",
			}),
			Integrations: []model.Integration{},
		},
		{
			Name: "error, context deadline exceeded",
			CTX: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			Error: context.Canceled,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			db.Wipe()
			client := db.Client()
			collIntegrations := client.
				Database(DbName).
				Collection(CollNameIntegrations)

			ctx := identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			})

			for _, integration := range tc.Integrations {
				_, err := collIntegrations.InsertOne(ctx,
					mstore.WithTenantID(ctx, integration),
				)
				require.NoError(t, err)
			}

			db := NewDataStoreWithClient(client)
			integrations, err := db.GetIntegrations(tc.CTX)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"error did not match expected expression",
					)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Integrations, integrations)
			}
		})
	}
}

func TestGetDevice(t *testing.T) {
	const deviceID = "1"
	const tenantID = "123456789012345678901234"
	testCases := []struct {
		Name string

		CTX context.Context

		Device *model.Device
		Error  error
	}{
		{
			Name: "ok",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			Device: &model.Device{
				ID: deviceID,
			},
		},
		{
			Name: "not found",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: "111111111111111111111111",
			}),
		},
		{
			Name: "error, context deadline exceeded",
			CTX: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			Error: context.Canceled,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			db.Wipe()
			client := db.Client()
			collDevices := client.
				Database(DbName).
				Collection(CollNameDevices)

			ctx := identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			})

			if tc.Device != nil {
				_, err := collDevices.InsertMany(ctx, []interface{}{
					mstore.WithTenantID(ctx, tc.Device),
				})
				assert.NoError(t, err)
			}

			db := NewDataStoreWithClient(client)
			device, err := db.GetDevice(tc.CTX, deviceID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"error did not match expected expression",
					)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Device, device)
			}
		})
	}
}

func TestGetDeviceByIntegrationID(t *testing.T) {
	const deviceID = "1"
	const tenantID = "123456789012345678901234"
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	testCases := []struct {
		Name string

		CTX context.Context

		Device   *model.Device
		NotFound bool
	}{
		{
			Name: "ok",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			Device: &model.Device{
				ID:             deviceID,
				IntegrationIDs: []uuid.UUID{integrationID},
			},
		},
		{
			Name: "not found",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: "111111111111111111111111",
			}),
			Device: &model.Device{
				ID:             deviceID,
				IntegrationIDs: []uuid.UUID{},
			},
			NotFound: true,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			db.Wipe()
			client := db.Client()
			collDevices := client.
				Database(DbName).
				Collection(CollNameDevices)

			ctx := identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			})

			if tc.Device != nil {
				_, err := collDevices.InsertMany(ctx, []interface{}{
					mstore.WithTenantID(ctx, tc.Device),
				})
				assert.NoError(t, err)
			}

			db := NewDataStoreWithClient(client)
			device, err := db.GetDeviceByIntegrationID(tc.CTX, deviceID, integrationID)
			assert.NoError(t, err)
			if tc.NotFound {
				assert.Nil(t, device)
			} else {
				assert.Equal(t, tc.Device, device)
			}
		})
	}
}

func TestGetIntegrationById(t *testing.T) {
	const tenantID = "123456789012345678901234"
	integrationID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("digest"))
	testCases := []struct {
		Name string

		CTX context.Context

		Integration *model.Integration
		Error       error
	}{
		{
			Name: "ok",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			Integration: &model.Integration{
				ID: integrationID,
			},
		},
		{
			Name: "not found",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: "111111111111111111111111",
			}),
			Error: store.ErrObjectNotFound,
		},
		{
			Name: "error, context deadline exceeded",
			CTX: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			Error: context.Canceled,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			db.Wipe()
			client := db.Client()
			collIntegrations := client.
				Database(DbName).
				Collection(CollNameIntegrations)

			ctx := identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			})

			if tc.Integration != nil {
				_, err := collIntegrations.InsertMany(ctx, []interface{}{
					mstore.WithTenantID(ctx, tc.Integration),
				})
				assert.NoError(t, err)
			}

			db := NewDataStoreWithClient(client)
			integration, err := db.GetIntegrationById(tc.CTX, integrationID)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t,
						tc.Error.Error(),
						err.Error(),
						"error did not match expected expression",
					)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Integration, integration)
			}
		})
	}
}
