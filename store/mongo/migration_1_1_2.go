// Copyright 2024 Northern.tech AS
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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mendersoftware/go-lib-micro/mongo/migrate"

	"github.com/mendersoftware/iot-manager/model"
)

type migration_1_1_2 struct {
	client *mongo.Client
	db     string
}

// Up creates indexes for fetching event documents
func (m *migration_1_1_2) Up(from migrate.Version) error {
	ctx := context.Background()
	_, err := m.client.
		Database(m.db).
		Collection(CollNameIntegrations).
		UpdateMany(
			ctx,
			bson.M{
				KeyScope: bson.M{"$exists": false},
			},
			bson.M{
				"$set": bson.M{"$scope": model.ScopeDeviceAuth},
			},
		)
	return err
}

func (m *migration_1_1_2) Version() migrate.Version {
	return migrate.MakeVersion(1, 1, 2)
}
