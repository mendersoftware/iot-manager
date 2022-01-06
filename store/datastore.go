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

package store

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/mendersoftware/iot-manager/model"
)

// DataStore interface for DataStore services
//nolint:lll
//go:generate ../utils/mockgen.sh
type DataStore interface {
	Ping(ctx context.Context) error
	Close() error

	GetIntegrations(context.Context, model.IntegrationFilter) ([]model.Integration, error)
	GetIntegrationById(context.Context, uuid.UUID) (*model.Integration, error)
	CreateIntegration(context.Context, model.Integration) (*model.Integration, error)
	GetDevice(ctx context.Context, deviceID string) (*model.Device, error)
	GetDeviceByIntegrationID(
		ctx context.Context,
		deviceID string,
		integrationID uuid.UUID,
	) (*model.Device, error)
	DoDevicesExistByIntegrationID(context.Context, uuid.UUID) (bool, error)
	// RemoveDevicesFromIntegration integration with given id from all devices
	// belonging to the tenant within the context.
	UpsertDeviceIntegrations(
		ctx context.Context,
		deviceID string,
		integrationIDs []uuid.UUID,
	) (newDevice *model.Device, err error)
	DeleteDevice(ctx context.Context, deviceID string) error
	SetIntegrationCredentials(context.Context, uuid.UUID, model.Credentials) error
	RemoveIntegration(context.Context, uuid.UUID) error
}

var (
	ErrSerialization  = errors.New("store: failed to serialize object")
	ErrObjectNotFound = errors.New("store: object not found")

	ErrObjectExists = errors.New("store: the object already exists")
)
