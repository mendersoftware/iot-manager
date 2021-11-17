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
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/go-lib-micro/config"
	"github.com/mendersoftware/go-lib-micro/identity"
	mstore "github.com/mendersoftware/go-lib-micro/store/v2"

	dconfig "github.com/mendersoftware/azure-iot-manager/config"
	"github.com/mendersoftware/azure-iot-manager/model"
	"github.com/mendersoftware/azure-iot-manager/store"
)

const (
	CollNameSettings = "settings"
	KeyTenantID      = "tenant_id"

	ConnectTimeoutSeconds = 10
	defaultAutomigrate    = false
)

var (
	ErrFailedToGetSettings = errors.New("Failed to get settings")
)

type Config struct {
	Automigrate *bool
}

func NewConfig() *Config {
	conf := new(Config)
	return conf.SetAutomigrate(defaultAutomigrate)
}

func (c *Config) SetAutomigrate(migrate bool) *Config {
	c.Automigrate = &migrate
	return c
}

func mergeConfig(configs []*Config) *Config {
	config := NewConfig()
	for _, c := range configs {
		if c.Automigrate != nil {
			config.SetAutomigrate(*c.Automigrate)
		}
	}
	return config
}

// SetupDataStore returns the mongo data store and optionally runs migrations
func SetupDataStore(conf *Config) (store.DataStore, error) {
	conf = mergeConfig([]*Config{conf})
	ctx := context.Background()
	dbClient, err := NewClient(ctx, config.Config)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to connect to db: %v", err))
	}
	err = doMigrations(ctx, dbClient, *conf.Automigrate)
	if err != nil {
		return nil, err
	}
	dataStore := NewDataStoreWithClient(dbClient)
	return dataStore, nil
}

func doMigrations(ctx context.Context, client *mongo.Client,
	automigrate bool) error {
	return Migrate(ctx, DbName, DbVersion, client, automigrate)
}

// NewClient returns a mongo client
func NewClient(ctx context.Context, c config.Reader) (*mongo.Client, error) {

	clientOptions := mopts.Client()
	mongoURL := c.GetString(dconfig.SettingMongo)
	if !strings.Contains(mongoURL, "://") {
		return nil, errors.Errorf("Invalid mongoURL %q: missing schema.",
			mongoURL)
	}
	clientOptions.ApplyURI(mongoURL)

	username := c.GetString(dconfig.SettingDbUsername)
	if username != "" {
		credentials := mopts.Credential{
			Username: c.GetString(dconfig.SettingDbUsername),
		}
		password := c.GetString(dconfig.SettingDbPassword)
		if password != "" {
			credentials.Password = password
			credentials.PasswordSet = true
		}
		clientOptions.SetAuth(credentials)
	}

	if c.GetBool(dconfig.SettingDbSSL) {
		tlsConfig := &tls.Config{}
		tlsConfig.InsecureSkipVerify = c.GetBool(dconfig.SettingDbSSLSkipVerify)
		clientOptions.SetTLSConfig(tlsConfig)
	}

	// Set 10s timeout
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, ConnectTimeoutSeconds*time.Second)
		defer cancel()
	}
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to mongo server")
	}

	// Validate connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "Error reaching mongo server")
	}

	return client, nil
}

// DataStoreMongo is the data storage service
type DataStoreMongo struct {
	// client holds the reference to the client used to communicate with the
	// mongodb server.
	client *mongo.Client

	*Config
}

// NewDataStoreWithClient initializes a DataStore object
func NewDataStoreWithClient(client *mongo.Client, conf ...*Config) store.DataStore {
	return &DataStoreMongo{
		client: client,
		Config: mergeConfig(conf),
	}
}

// Ping verifies the connection to the database
func (db *DataStoreMongo) Ping(ctx context.Context) error {
	res := db.client.Database(DbName).RunCommand(ctx, bson.M{"ping": 1})
	return res.Err()
}

func (db *DataStoreMongo) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err := db.client.Disconnect(ctx)
	return err
}

func (db *DataStoreMongo) SetSettings(ctx context.Context, settings model.Settings) error {
	collSettings := db.client.Database(DbName).Collection(CollNameSettings)
	o := mopts.Replace().SetUpsert(true)

	identity := identity.FromContext(ctx)
	tenantID := ""
	if identity != nil {
		tenantID = identity.Tenant
	}

	_, err := collSettings.ReplaceOne(
		ctx,
		bson.D{{Key: KeyTenantID, Value: tenantID}},
		mstore.WithTenantID(ctx, settings),
		o,
	)
	if err != nil && err != mongo.ErrNoDocuments {
		return errors.Wrapf(err, "failed to store settings %v", settings)
	}

	return err
}

func (db *DataStoreMongo) GetSettings(ctx context.Context) (model.Settings, error) {
	var settings model.Settings

	collSettings := db.client.Database(DbName).Collection(CollNameSettings)
	tenantId := ""
	id := identity.FromContext(ctx)
	if id != nil {
		tenantId = id.Tenant
	}

	if err := collSettings.FindOne(ctx,
		bson.M{KeyTenantID: tenantId},
	).Decode(&settings); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return model.Settings{}, nil
		default:
			return model.Settings{}, errors.Wrap(err, ErrFailedToGetSettings.Error())
		}
	}
	return settings, nil
}
