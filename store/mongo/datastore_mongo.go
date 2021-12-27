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

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopts "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mendersoftware/go-lib-micro/config"
	"github.com/mendersoftware/go-lib-micro/identity"
	mstore "github.com/mendersoftware/go-lib-micro/store/v2"

	dconfig "github.com/mendersoftware/iot-manager/config"
	"github.com/mendersoftware/iot-manager/model"
	"github.com/mendersoftware/iot-manager/store"
)

const (
	CollNameIntegrations = "integrations"
	CollNameDevices      = "devices"

	KeyID             = "_id"
	KeyProvider       = "provider"
	KeyTenantID       = "tenant_id"
	KeyIntegrationIDs = "integration_ids"

	ConnectTimeoutSeconds = 10
	defaultAutomigrate    = false
)

var (
	ErrFailedToGetIntegrations = errors.New("failed to get integrations")
	ErrFailedToGetDevice       = errors.New("failed to get device")
	ErrFailedToGetSettings     = errors.New("failed to get settings")
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

func (db *DataStoreMongo) GetIntegrations(ctx context.Context) ([]model.Integration, error) {
	// TODO: Should I add filter with limit?
	var (
		err     error
		results = []model.Integration{}
	)

	collIntegrations := db.client.
		Database(DbName).
		Collection(CollNameIntegrations)
	findOpts := mopts.Find().
		SetSort(bson.D{{
			Key:   KeyProvider,
			Value: 1,
		}, {
			Key:   KeyID,
			Value: 1,
		}})
	tenantId := ""
	id := identity.FromContext(ctx)
	if id != nil {
		tenantId = id.Tenant
	}

	cur, err := collIntegrations.Find(ctx,
		bson.D{{Key: KeyTenantID, Value: tenantId}},
		findOpts,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error executing integrations collection request")
	}
	if err = cur.All(ctx, &results); err != nil {
		return nil, errors.Wrap(err, "error retrieving integrations collection results")
	}

	return results, nil
}

func (db *DataStoreMongo) GetIntegrationById(
	ctx context.Context,
	integrationId uuid.UUID,
) (*model.Integration, error) {
	var integration = new(model.Integration)

	collIntegrations := db.client.Database(DbName).Collection(CollNameIntegrations)
	tenantId := ""
	id := identity.FromContext(ctx)
	if id != nil {
		tenantId = id.Tenant
	}

	if err := collIntegrations.FindOne(ctx,
		bson.M{KeyTenantID: tenantId},
	).Decode(&integration); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, store.ErrObjectNotFound
		default:
			return nil, errors.Wrap(err, ErrFailedToGetIntegrations.Error())
		}
	}
	return integration, nil
}

func (db *DataStoreMongo) CreateIntegration(
	ctx context.Context,
	integration model.Integration,
) error {
	var tenantID string
	if id := identity.FromContext(ctx); id != nil {
		tenantID = id.Tenant
	}
	collIntegrations := db.client.
		Database(DbName).
		Collection(CollNameIntegrations)

	// Force a single integration per tenant by utilizing unique '_id' index
	integration.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(tenantID))

	_, err := collIntegrations.
		InsertOne(ctx, mstore.WithTenantID(ctx, integration))
	if err != nil {
		if isDuplicateKeyError(err) {
			return store.ErrObjectExists
		}
		return errors.Wrapf(err, "failed to store integration %v", integration)
	}

	return err
}

func (db *DataStoreMongo) GetDevice(ctx context.Context, deviceID string) (*model.Device, error) {
	return db.GetDeviceByIntegrationID(ctx, deviceID, uuid.Nil)
}

func (db *DataStoreMongo) GetDeviceByIntegrationID(
	ctx context.Context,
	deviceID string,
	integrationID uuid.UUID,
) (*model.Device, error) {
	var device *model.Device

	collDevices := db.client.Database(DbName).Collection(CollNameDevices)
	tenantId := ""
	id := identity.FromContext(ctx)
	if id != nil {
		tenantId = id.Tenant
	}

	filter := bson.M{KeyID: deviceID, KeyTenantID: tenantId}
	if integrationID != uuid.Nil {
		filter[KeyIntegrationIDs] = integrationID
	}
	if err := collDevices.FindOne(ctx,
		filter,
	).Decode(&device); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, nil
		default:
			return nil, errors.Wrap(err, ErrFailedToGetDevice.Error())
		}
	}
	return device, nil
}
