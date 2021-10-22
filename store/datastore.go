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

	"github.com/mendersoftware/azure-iot-manager/model"
)

// DataStore interface for DataStore services
//nolint:lll
//go:generate ../utils/mockgen.sh
type DataStore interface {
	Ping(ctx context.Context) error
	Close() error

	SetSettings(ctx context.Context, settings model.Settings) error
	GetSettings(ctx context.Context) (model.Settings, error)
}

var (
	ErrSerialization  = errors.New("store: failed to serialize object")
	ErrObjectNotFound = errors.New("store: object not found")
)
