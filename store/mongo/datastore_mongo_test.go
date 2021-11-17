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

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/go-lib-micro/identity"
	mstore "github.com/mendersoftware/go-lib-micro/store/v2"

	"github.com/mendersoftware/azure-iot-manager/model"
)

func TestSetSettings(t *testing.T) {
	testCases := []struct {
		Name string

		CTX      context.Context
		Settings model.Settings

		Error error
	}{
		{
			Name: "ok",

			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: "1234567890",
			}),
			Settings: model.Settings{
				ConnectionString: &model.ConnectionString{
					HostName: "localhost",
					Key:      []byte("password123"),
					Name:     "acmeHub",
				},
			},
		}, {
			Name: "ok, no tenant context",

			CTX: context.Background(),
			Settings: model.Settings{
				ConnectionString: &model.ConnectionString{
					HostName: "localhost",
					Key:      []byte("password123"),
					Name:     "acmeHub",
				},
			},
		}, {
			Name: "error, context canceled",

			CTX: func() context.Context {
				ctx, cc := context.WithCancel(context.Background())
				cc()
				return ctx
			}(),
			Settings: model.Settings{
				ConnectionString: &model.ConnectionString{
					HostName: "localhost",
					Key:      []byte("password123"),
					Name:     "acmeHub",
				},
			},
			Error: context.Canceled,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			db.Wipe()
			mgo := db.Client()
			ds := NewDataStoreWithClient(mgo)
			err := ds.SetSettings(tc.CTX, tc.Settings)
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
					Collection(CollNameSettings).
					FindOne(tc.CTX, fltr).
					Decode(&doc)
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				field := doc.Lookup(KeyTenantID)
				actualTID, ok := field.StringValueOK()
				assert.True(t, ok, "bson document does not contain tenant_id field")
				assert.Equal(t, tenantID, actualTID)

				var settings model.Settings
				bson.Unmarshal(doc, &settings)
				assert.Equal(t, tc.Settings, settings)
			}
		})
	}
}

func TestGetSettings(t *testing.T) {
	const tenantID = "123456789012345678901234"
	testCases := []struct {
		Name string

		CTX context.Context

		Settings model.Settings
		Error    error
	}{
		{
			Name: "ok got settings",
			CTX: identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			}),
			Settings: model.Settings{
				ConnectionString: &model.ConnectionString{
					HostName: "localhost",
					Key:      []byte("password123"),
					Name:     "acmeHub",
				},
			},
		},
		{
			Name: "no settings for tenant",
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
			collAlerts := client.
				Database(DbName).
				Collection(CollNameSettings)

			ctx := identity.WithContext(context.Background(), &identity.Identity{
				Tenant: tenantID,
			})

			_, err := collAlerts.InsertMany(ctx, []interface{}{
				mstore.WithTenantID(ctx, tc.Settings),
			})
			assert.NoError(t, err)

			db := NewDataStoreWithClient(client)
			settings, err := db.GetSettings(tc.CTX)
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
				assert.Equal(t, tc.Settings, settings)
			}
		})
	}
}
