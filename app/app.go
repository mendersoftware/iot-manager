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
	"net/http"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/iot-manager/client"
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
	ErrDeviceAlreadyExists      = errors.New("device already exists")
)

const (
	confKeyPrimaryKey   = "$azure.primaryKey"
	confKeySecondaryKey = "$azure.secondaryKey"
)

type DeviceUpdate iothub.Device

type Status iothub.Status

const (
	StatusEnabled  = Status(iothub.StatusEnabled)
	StatusDisabled = Status(iothub.StatusDisabled)
)

// App interface describes app objects
//nolint:lll
//go:generate ../utils/mockgen.sh
type App interface {
	HealthCheck(context.Context) error
	GetDeviceIntegrations(context.Context, string) ([]model.Integration, error)
	GetIntegrations(context.Context) ([]model.Integration, error)
	GetIntegrationById(context.Context, uuid.UUID) (*model.Integration, error)
	CreateIntegration(context.Context, model.Integration) error
	SetDeviceStatus(context.Context, string, Status) error
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
	return a.store.GetIntegrations(ctx)
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

func (a *app) CreateIntegration(ctx context.Context, integration model.Integration) error {
	err := a.store.CreateIntegration(ctx, integration)
	if err == store.ErrObjectExists {
		return ErrIntegrationExists
	}
	return err
}

func (a *app) GetDeviceIntegrations(
	ctx context.Context,
	deviceID string,
) ([]model.Integration, error) {
	// TODO: stub only, needs to be implemented
	return []model.Integration{}, nil
}

func (a *app) SetDeviceStatus(ctx context.Context, deviceID string, status Status) error {
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
			dev, err := a.hub.GetDevice(ctx, cs, deviceID)
			if err != nil {
				return errors.Wrap(err, "failed to retrieve device from IoT Hub")
			} else if dev.Status == iothub.Status(status) {
				// We're done...
				return nil
			}

			dev.Status = iothub.Status(status)
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

	for _, integration := range integrations {
		// NEXT Create device document.
		// TODO Filter only by integrations that apply to device.
		switch integration.Provider {
		case model.ProviderIoTHub:
			cs := integration.Credentials.ConnectionString
			if cs == nil {
				return ErrNoConnectionString
			}

			dev, err := a.hub.UpsertDevice(ctx, cs, deviceID)
			if err != nil {
				if htErr, ok := err.(client.HTTPError); ok {
					switch htErr.Code {
					case http.StatusUnauthorized:
						return ErrNoConnectionString
					case http.StatusConflict:
						return ErrDeviceAlreadyExists
					}
				}
				return errors.Wrap(err, "failed to update iothub devices")
			}
			if dev.Auth == nil || dev.Auth.SymmetricKey == nil {
				return ErrNoDeviceConnectionString
			}
			primKey := &model.ConnectionString{
				Key:      dev.Auth.SymmetricKey.Primary,
				DeviceID: dev.DeviceID,
				HostName: cs.HostName,
			}
			secKey := &model.ConnectionString{
				Key:      dev.Auth.SymmetricKey.Secondary,
				DeviceID: dev.DeviceID,
				HostName: cs.HostName,
			}

			err = a.wf.ProvisionExternalDevice(ctx, dev.DeviceID, map[string]string{
				confKeyPrimaryKey:   primKey.String(),
				confKeySecondaryKey: secKey.String(),
			})
			if err != nil {
				return errors.Wrap(err, "failed to submit iothub authn to deviceconfig")
			}
			err = a.hub.UpdateDeviceTwin(ctx, cs, dev.DeviceID, &iothub.DeviceTwinUpdate{
				Tags: map[string]interface{}{
					"mender": true,
				},
			})
			return errors.Wrap(err, "failed to tag provisioned iothub device")
		default:
			continue
		}
	}
	return nil
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
	return nil
}

func (a *app) GetDevice(ctx context.Context, deviceID string) (*model.Device, error) {
	return a.store.GetDevice(ctx, deviceID)
}

func (a *app) GetDeviceStateIntegration(
	ctx context.Context,
	deviceID string,
	integrationID uuid.UUID,
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
