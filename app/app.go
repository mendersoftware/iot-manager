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

	"github.com/mendersoftware/azure-iot-manager/client/iothub"
	"github.com/mendersoftware/azure-iot-manager/client/workflows"
	"github.com/mendersoftware/azure-iot-manager/model"
	"github.com/mendersoftware/azure-iot-manager/store"

	"github.com/pkg/errors"
)

var (
	ErrNoConnectionString = errors.New("no connection string configured for tenant")

	ErrNoDeviceConnectionString = errors.New("device has no connection string")
)

const (
	confKeyPrimaryKey   = "$azure.primaryKey"
	confKeySecondaryKey = "$azure.secondaryKey"
)

type DeviceUpdate iothub.Device

// App interface describes app objects
//nolint:lll
//go:generate ../utils/mockgen.sh
type App interface {
	HealthCheck(context.Context) error
	GetSettings(context.Context) (model.Settings, error)
	SetSettings(context.Context, model.Settings) error
	ProvisionDevice(context.Context, string) error
	DeleteIOTHubDevice(context.Context, string) error
}

// app is an app object
type app struct {
	store store.DataStore
	hub   iothub.Client
	wf    workflows.Client
}

// NewApp initialize a new azure-iot-manager App
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

func (a *app) GetSettings(ctx context.Context) (model.Settings, error) {
	return a.store.GetSettings(ctx)
}

func (a *app) SetSettings(ctx context.Context, settings model.Settings) error {
	return a.store.SetSettings(ctx, settings)
}

func (a *app) ProvisionDevice(
	ctx context.Context,
	deviceID string,
) error {
	settings, err := a.GetSettings(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve settings")
	}
	cs := settings.ConnectionString
	if cs == nil {
		return ErrNoConnectionString
	}

	dev, err := a.hub.UpsertDevice(ctx, cs, deviceID)
	if err != nil {
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

	err = a.wf.SubmitDeviceConfiguration(ctx, dev.DeviceID, map[string]string{
		confKeyPrimaryKey:   primKey.String(),
		confKeySecondaryKey: secKey.String(),
	})
	return errors.Wrap(err, "failed to submit iothub authn to deviceconfig")
}

func (a *app) DeleteIOTHubDevice(ctx context.Context, deviceID string) error {
	settings, err := a.GetSettings(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve settings")
	}
	cs := settings.ConnectionString
	if cs == nil {
		return ErrNoConnectionString
	}
	err = a.hub.DeleteDevice(ctx, cs, deviceID)
	if err != nil {
		return errors.Wrap(err, "failed to delete IoT Hub device")
	}
	return nil
}
