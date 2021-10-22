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

	"github.com/mendersoftware/azure-iot-manager/model"
	"github.com/mendersoftware/azure-iot-manager/store"
)

// App interface describes app objects
//nolint:lll
//go:generate ../utils/mockgen.sh
type App interface {
	HealthCheck(ctx context.Context) error
	GetSettings(ctx context.Context) (model.Settings, error)
	SetSettings(ctx context.Context, settings model.Settings) error
}

// app is an app object
type app struct {
	Config
	store store.DataStore
}

type Config struct {
}

// NewApp initialize a new azure-iot-manager App
func New(config Config, ds store.DataStore) App {
	return &app{
		Config: config,
		store:  ds,
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
