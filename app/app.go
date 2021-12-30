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

	"github.com/google/uuid"
	"github.com/pkg/errors"

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

	ErrDeviceAlreadyExists = errors.New("device already exists")
	ErrDeviceNotFound      = errors.New("device not found")
	ErrDeviceStateConflict = errors.New("conflict when updating the device state")
)

const (
	confKeyPrimaryKey = "$azure.primaryKey"
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
	GetDevice(context.Context, string) (*model.Device, error)
	GetDeviceStateIntegration(context.Context, string, uuid.UUID) (*model.DeviceState, error)
	SetDeviceStateIntegration(context.Context, string, uuid.UUID, *model.DeviceState) (*model.DeviceState, error)
	GetDeviceStateIoTHub(context.Context, string, *model.Integration) (*model.DeviceState, error)
	SetDeviceStateIoTHub(context.Context, string, *model.Integration, *model.DeviceState) (*model.DeviceState, error)
	ProvisionDevice(context.Context, string) error
	DeleteIOTHubDevice(context.Context, string) error
}

// app is an app object
type app struct {
	store store.DataStore
	hub   iothub.Client
	wf    workflows.Client
}

// NewApp initialize a new iot-manager App
func New(ds store.DataStore, hub iothub.Client, wf workflows.Client) App {
	return &app{
		store: ds,
		hub:   hub,
		wf:    wf,
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
		return errors.Wrap(err, "failed to retrieve settings")
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
		return errors.Wrap(err, "failed to retrieve integration")
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

func (a *app) DeleteIOTHubDevice(ctx context.Context, deviceID string) error {
	integrations, err := a.GetDeviceIntegrations(ctx, deviceID)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve settings")
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
