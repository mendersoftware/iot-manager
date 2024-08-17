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
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendersoftware/go-lib-micro/mongo/migrate"

	"github.com/mendersoftware/iot-manager/model"
)

func TestMigration_1_1_2(t *testing.T) {
	ctx := context.Background()
	client := db.Client()
	documents := []interface{}{
		model.Integration{
			ID:          uuid.New(),
			Provider:    "p1",
			Description: "some 1 dev",
		},
		[]interface{}{
			model.Integration{
				ID:          uuid.New(),
				Provider:    "p2",
				Description: "some 2 dev",
			},
		},
	}
	_, err := client.Database(DbName).
		Collection(CollNameIntegrations).
		InsertMany(
			ctx,
			documents,
		)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	m := &migration_1_1_2{
		client: client,
		db:     DbName,
	}
	from := migrate.MakeVersion(0, 0, 0)

	err = m.Up(from)
	require.NoError(t, err)

	cursor, err := client.Database(DbName).
		Collection(CollNameIntegrations).
		Find(ctx, bson.M{KeyScope: model.ScopeDeviceAuth})
	assert.NoError(t, err)
	assert.Equal(t, cursor.RemainingBatchLength(), len(documents))
}
