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
	"net/http"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/log"

	"github.com/mendersoftware/iot-manager/client"
	"github.com/mendersoftware/iot-manager/client/devauth"
	"github.com/mendersoftware/iot-manager/client/iothub"
	"github.com/mendersoftware/iot-manager/client/workflows"
	"github.com/mendersoftware/iot-manager/model"
	"github.com/mendersoftware/iot-manager/store"
)

var (
	ErrIntegrationNotFound      = errors.New("integration not found")
	ErrIntegrationExists        = errors.New("integration already exists")
	ErrUnknownIntegration       = errors.New("unknown integration provider")
	ErrNoConnectionString       = errors.New("no connection string configured for tenant")
	ErrNoDeviceConnectionString = errors.New("device has no connection string")

	ErrDeviceAlreadyExists     = errors.New("device already exists")
	ErrDeviceNotFound          = errors.New("device not found")
	ErrDeviceStateConflict     = errors.New("conflict when updating the device state")
	ErrCannotRemoveIntegration = errors.New("cannot remove integration in use by devices")
)

const (
	confKeyPrimaryKey = "$azure.connectionString"
)

type DeviceUpdate iothub.Device

// App interface describes app objects
//nolint:lll
//go:generate ../utils/mockgen.sh
type App interface {
	HealthCheck(context.Context) error
	GetDeviceIntegrations(context.Context, string) ([]model.Integration, error)
	GetIntegrations(context.Context) ([]model.Integration, error)
	GetIntegrationById(context.Context, uuid.UUID) (*model.Integration, error)
	CreateIntegration(context.Context, model.Integration) (*model.Integration, error)
	SetDeviceStatus(context.Context, string, model.Status) error
	SetIntegrationCredentials(context.Context, uuid.UUID, model.Credentials) error
	RemoveIntegration(context.Context, uuid.UUID) error
	GetDevice(context.Context, string) (*model.Device, error)
	GetDeviceStateIntegration(context.Context, string, uuid.UUID) (*model.DeviceState, error)
	SetDeviceStateIntegration(context.Context, string, uuid.UUID, *model.DeviceState) (*model.DeviceState, error)
	GetDeviceStateIoTHub(context.Context, string, *model.Integration) (*model.DeviceState, error)
	SetDeviceStateIoTHub(context.Context, string, *model.Integration, *model.DeviceState) (*model.DeviceState, error)
	ProvisionDevice(context.Context, string) error
	DecommissionDevice(context.Context, string) error

	SyncDevices(context.Context, int, bool) error
}

// app is an app object
type app struct {
	store   store.DataStore
	hub     iothub.Client
	wf      workflows.Client
	devauth devauth.Client
}

// NewApp initialize a new iot-manager App
func New(ds store.DataStore, hub iothub.Client, wf workflows.Client, da devauth.Client) App {
	return &app{
		store:   ds,
		hub:     hub,
		wf:      wf,
		devauth: da,
	}
}

// HealthCheck performs a health check and returns an error if it fails
func (a *app) HealthCheck(ctx context.Context) error {
	return a.store.Ping(ctx)
}

func (a *app) GetIntegrations(ctx context.Context) ([]model.Integration, error) {
	return a.store.GetIntegrations(ctx, model.IntegrationFilter{})
}

func (a *app) GetIntegrationById(ctx context.Context, id uuid.UUID) (*model.Integration, error) {
	integration, err := a.store.GetIntegrationById(ctx, id)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case store.ErrObjectNotFound:
			return nil, ErrIntegrationNotFound
		default:
			return nil, err
		}
	}
	return integration, err
}

func (a *app) CreateIntegration(
	ctx context.Context,
	integration model.Integration,
) (*model.Integration, error) {
	result, err := a.store.CreateIntegration(ctx, integration)
	if err == store.ErrObjectExists {
		return nil, ErrIntegrationExists
	}
	return result, err
}

func (a *app) SetIntegrationCredentials(
	ctx context.Context,
	integrationID uuid.UUID,
	credentials model.Credentials,
) error {
	err := a.store.SetIntegrationCredentials(ctx, integrationID, credentials)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case store.ErrObjectNotFound:
			return ErrIntegrationNotFound
		default:
			return err
		}
	}
	return err
}

func (a *app) RemoveIntegration(
	ctx context.Context,
	integrationID uuid.UUID,
) error {
	// check if there are any devices with given integration enabled
	devicesExist, err := a.store.DoDevicesExistByIntegrationID(ctx, integrationID)
	if err != nil {
		return err
	}
	if devicesExist {
		return ErrCannotRemoveIntegration
	}
	err = a.store.RemoveIntegration(ctx, integrationID)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case store.ErrObjectNotFound:
			return ErrIntegrationNotFound
		default:
			return err
		}
	}
	return err
}

func (a *app) GetDeviceIntegrations(
	ctx context.Context,
	deviceID string,
) ([]model.Integration, error) {
	device, err := a.store.GetDevice(ctx, deviceID)
	if err != nil {
		if err == store.ErrObjectNotFound {
			return nil, ErrDeviceNotFound
		}
		return nil, errors.Wrap(err, "app: failed to get device integrations")
	}
	if len(device.IntegrationIDs) > 0 {
		integrations, err := a.store.GetIntegrations(ctx,
			model.IntegrationFilter{IDs: device.IntegrationIDs},
		)
		return integrations, errors.Wrap(err, "app: failed to get device integrations")
	}
	return []model.Integration{}, nil
}

func (a *app) SetDeviceStatus(ctx context.Context, deviceID string, status model.Status) error {
	integrations, err := a.GetDeviceIntegrations(ctx, deviceID)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve device integrations")
	}

	for _, integration := range integrations {
		switch integration.Provider {
		case model.ProviderIoTHub:
			cs := integration.Credentials.ConnectionString
			if cs == nil {
				return ErrNoConnectionString
			}
			azureStatus := iothub.NewStatusFromMenderStatus(status)
			dev, err := a.hub.GetDevice(ctx, cs, deviceID)
			if err != nil {
				return errors.Wrap(err, "failed to retrieve device from IoT Hub")
			} else if dev.Status == azureStatus {
				// We're done...
				return nil
			}

			dev.Status = azureStatus
			_, err = a.hub.UpsertDevice(ctx, cs, deviceID, dev)
			if err != nil {
				return errors.Wrap(err, "failed to update IoT Hub device")
			}
		default:
			continue
		}
	}
	return nil
}

func (a *app) ProvisionDevice(
	ctx context.Context,
	deviceID string,
) error {
	integrations, err := a.GetIntegrations(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve integrations")
	}
	integrationIDs := make([]uuid.UUID, 0, len(integrations))
	for _, integration := range integrations {
		switch integration.Provider {
		case model.ProviderIoTHub:
			err = a.provisionIoTHubDevice(ctx, deviceID, integration)
		default:
			continue
		}
		if err != nil {
			return err
		}
		integrationIDs = append(integrationIDs, integration.ID)
	}
	_, err = a.store.UpsertDeviceIntegrations(ctx, deviceID, integrationIDs)
	return err
}

func (a *app) syncBatch(
	ctx context.Context,
	devices []model.Device,
	integCache map[uuid.UUID]*model.Integration,
	failEarly bool,
) error {
	var err error
	l := log.FromContext(ctx)

	deviceMap := make(map[uuid.UUID][]string, len(integCache))
	for _, dev := range devices {
		for _, id := range dev.IntegrationIDs {
			deviceMap[id] = append(deviceMap[id], dev.ID)
		}
	}

	for integID, deviceIDs := range deviceMap {
		integration, ok := integCache[integID]
		if !ok {
			// (Data race) Try again to fetch the integration
			integration, err = a.store.GetIntegrationById(ctx, integID)
			if err != nil {
				if err == store.ErrObjectNotFound {
					integCache[integID] = nil
					continue
				}
				err = errors.Wrap(err, "failed to retrieve device integration")
				if failEarly {
					return err
				}
				l.Errorf("failed to get device integration: %s", err)
				continue
			} else {
				integCache[integID] = integration
			}
		}
		if integration == nil {
			// Should not occur, but is not impossible since mongo client
			// caches batches of results.
			_, err := a.store.RemoveDevicesFromIntegration(ctx, integID)
			if err != nil {
				err = errors.Wrap(err, "failed to remove integration from devices")
				if failEarly {
					return err
				}
				l.Error(err)
			}
			continue
		}

		switch integration.Provider {
		case model.ProviderIoTHub:
			err := a.syncIoTHubDevices(ctx, deviceIDs, *integration, failEarly)
			if err != nil {
				if failEarly {
					return err
				}
				l.Error(err)
			}
		default:
			// Invalid integration
			// FIXME(alf) what to do?
		}
	}

	return nil
}

func (a app) syncCacheIntegrations(ctx context.Context) (map[uuid.UUID]*model.Integration, error) {
	// NOTE At the time of writing this, we don't allow more than one
	//      integration per tenant so this const doesn't matter.
	// TODO Will we need a more sophisticated cache data structure?
	const MaxIntegrationsToCache = 20
	// Cache integrations for the given tenant
	integrations, err := a.store.GetIntegrations(
		ctx, model.IntegrationFilter{Limit: MaxIntegrationsToCache},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get integrations for tenant")
	}
	integCache := make(map[uuid.UUID]*model.Integration, len(integrations))
	for i := range integrations {
		integCache[integrations[i].ID] = &integrations[i]
	}
	return integCache, nil
}

func (a *app) SyncDevices(
	ctx context.Context,
	batchSize int,
	failEarly bool,
) error {
	type DeviceWithTenantID struct {
		model.Device `bson:",inline"`
		TenantID     string `bson:"tenant_id"`
	}
	iter, err := a.store.GetAllDevices(ctx)
	if err != nil {
		return err
	}
	defer iter.Close(ctx)

	var (
		deviceBatch        = make([]model.Device, 0, batchSize)
		tenantID    string = ""
		integCache  map[uuid.UUID]*model.Integration
	)
	tCtx := identity.WithContext(ctx, &identity.Identity{
		Tenant: tenantID,
	})
	integCache, err = a.syncCacheIntegrations(tCtx)
	if err != nil {
		return err
	}
	for iter.Next(ctx) {
		dev := DeviceWithTenantID{}
		err := iter.Decode(&dev)
		if err != nil {
			return err
		}
		if len(deviceBatch) == cap(deviceBatch) ||
			(tenantID != dev.TenantID && len(deviceBatch) > 0) {
			err := a.syncBatch(tCtx, deviceBatch, integCache, failEarly)
			if err != nil {
				return err
			}
			deviceBatch = deviceBatch[:0]
		}
		if tenantID != dev.TenantID {
			tenantID = dev.TenantID
			tCtx = identity.WithContext(ctx, &identity.Identity{
				Tenant: tenantID,
			})

			integCache, err = a.syncCacheIntegrations(tCtx)
			if err != nil {
				return err
			}

		}
		deviceBatch = append(deviceBatch, dev.Device)
	}
	if len(deviceBatch) > 0 {
		err := a.syncBatch(tCtx, deviceBatch, integCache, failEarly)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *app) DecommissionDevice(ctx context.Context, deviceID string) error {
	integrations, err := a.GetDeviceIntegrations(ctx, deviceID)
	if err != nil {
		return err
	}

	for _, integration := range integrations {
		switch integration.Provider {
		case model.ProviderIoTHub:
			cs := integration.Credentials.ConnectionString
			if cs == nil {
				return ErrNoConnectionString
			}
			err = a.hub.DeleteDevice(ctx, cs, deviceID)
			if err != nil {
				if htErr, ok := err.(client.HTTPError); ok &&
					htErr.Code() == http.StatusNotFound {
					continue
				}
				return errors.Wrap(err, "failed to delete IoT Hub device")
			}
		default:
			continue
		}
	}
	err = a.store.DeleteDevice(ctx, deviceID)
	if err == store.ErrObjectNotFound {
		return ErrDeviceNotFound
	}
	return err
}

func (a *app) GetDevice(ctx context.Context, deviceID string) (*model.Device, error) {
	device, err := a.store.GetDevice(ctx, deviceID)
	if err == store.ErrObjectNotFound {
		return nil, ErrDeviceNotFound
	}
	return device, err
}

func (a *app) GetDeviceStateIntegration(
	ctx context.Context,
	deviceID string,
	integrationID uuid.UUID,
) (*model.DeviceState, error) {
	_, err := a.store.GetDeviceByIntegrationID(ctx, deviceID, integrationID)
	if err != nil {
		if err == store.ErrObjectNotFound {
			return nil, ErrIntegrationNotFound
		}
		return nil, errors.Wrap(err, "failed to retrieve the device")
	}
	integration, err := a.store.GetIntegrationById(ctx, integrationID)
	if integration == nil && (err == nil || err == store.ErrObjectNotFound) {
		return nil, ErrIntegrationNotFound
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the integration")
	}
	switch integration.Provider {
	case model.ProviderIoTHub:
		return a.GetDeviceStateIoTHub(ctx, deviceID, integration)
	default:
		return nil, ErrUnknownIntegration
	}
}

func (a *app) SetDeviceStateIntegration(
	ctx context.Context,
	deviceID string,
	integrationID uuid.UUID,
	state *model.DeviceState,
) (*model.DeviceState, error) {
	device, err := a.store.GetDeviceByIntegrationID(ctx, deviceID, integrationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the device")
	} else if device == nil {
		return nil, ErrIntegrationNotFound
	}
	integration, err := a.store.GetIntegrationById(ctx, integrationID)
	if integration == nil && (err == nil || err == store.ErrObjectNotFound) {
		return nil, ErrIntegrationNotFound
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the integration")
	}
	switch integration.Provider {
	case model.ProviderIoTHub:
		return a.SetDeviceStateIoTHub(ctx, deviceID, integration, state)
	default:
		return nil, ErrUnknownIntegration
	}
}
