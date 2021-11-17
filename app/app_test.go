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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/azure-iot-manager/model"
	storeMocks "github.com/mendersoftware/azure-iot-manager/store/mocks"
)

func TestHealthCheck(t *testing.T) {
	testCases := []struct {
		Name string

		PingReturn    error
		ExpectedError error
	}{
		{
			Name:          "db Ping failed",
			PingReturn:    errors.New("failed to connect to db"),
			ExpectedError: errors.New("failed to connect to db"),
		},
		{
			Name: "db Ping successful",
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("Ping",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
			).Return(tc.PingReturn)
			app := New(store, nil, nil)

			ctx := context.Background()
			err := app.HealthCheck(ctx)
			if tc.ExpectedError != nil {
				assert.EqualError(t, err, tc.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetSettings(t *testing.T) {
	testCases := []struct {
		Name string

		GetSettingsSettings model.Settings
		GetSettingsError    error
	}{
		{
			Name: "settings exist",

			GetSettingsSettings: model.Settings{
				ConnectionString: &model.ConnectionString{
					HostName: "localhost",
					Key:      []byte("secret"),
					Name:     "foobar",
				},
			},
		},
		{
			Name: "settings do not exists",

			GetSettingsSettings: model.Settings{},
		},
		{
			Name: "settings retrieval error",

			GetSettingsError: errors.New("error getting the settings"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("GetSettings",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
			).Return(tc.GetSettingsSettings, tc.GetSettingsError)
			app := New(store, nil, nil)

			ctx := context.Background()
			settings, err := app.GetSettings(ctx)
			if tc.GetSettingsError != nil {
				assert.EqualError(t, err, tc.GetSettingsError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.GetSettingsSettings, settings)
			}
		})
	}
}

func TestSetSettings(t *testing.T) {
	testCases := []struct {
		Name string

		SetSettingsSettings model.Settings
		SetSettingsError    error
	}{
		{
			Name: "settings saved",

			SetSettingsSettings: model.Settings{
				ConnectionString: &model.ConnectionString{
					HostName: "localhost",
					Key:      []byte("secret"),
					Name:     "foobar",
				},
			},
		},
		{
			Name: "settings saving error",

			SetSettingsError: errors.New("error setting the settings"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			store := &storeMocks.DataStore{}
			store.On("SetSettings",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
				mock.AnythingOfType("model.Settings"),
			).Return(tc.SetSettingsError)
			app := New(store, nil, nil)

			ctx := context.Background()
			err := app.SetSettings(ctx, tc.SetSettingsSettings)
			if tc.SetSettingsError != nil {
				assert.EqualError(t, err, tc.SetSettingsError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
