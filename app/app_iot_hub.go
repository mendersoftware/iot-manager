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

	"github.com/pkg/errors"

	"github.com/mendersoftware/iot-manager/client"
	"github.com/mendersoftware/iot-manager/client/iothub"
	"github.com/mendersoftware/iot-manager/crypto"
	"github.com/mendersoftware/iot-manager/model"
)

func (a *app) provisionIoTHubDevice(
	ctx context.Context,
	deviceID string,
	integration model.Integration,
) error {
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
		Key:      crypto.String(dev.Auth.SymmetricKey.Primary),
		DeviceID: dev.DeviceID,
		HostName: cs.HostName,
	}

	err = a.wf.ProvisionExternalDevice(ctx, dev.DeviceID, map[string]string{
		confKeyPrimaryKey: primKey.String(),
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
}

func (a *app) GetDeviceStateIoTHub(
	ctx context.Context,
	deviceID string,
	integration *model.Integration,
) (*model.DeviceState, error) {
	cs := integration.Credentials.ConnectionString
	if cs == nil {
		return nil, ErrNoConnectionString
	}
	twin, err := a.hub.GetDeviceTwin(ctx, cs, deviceID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the device twin")
	}
	return &model.DeviceState{
		Desired:  twin.Properties.Desired,
		Reported: twin.Properties.Reported,
	}, nil
}

func (a *app) SetDeviceStateIoTHub(
	ctx context.Context,
	deviceID string,
	integration *model.Integration,
	state *model.DeviceState,
) (*model.DeviceState, error) {
	cs := integration.Credentials.ConnectionString
	if cs == nil {
		return nil, ErrNoConnectionString
	}
	twin, err := a.hub.GetDeviceTwin(ctx, cs, deviceID)
	if err == nil {
		update := &iothub.DeviceTwinUpdate{
			Tags: twin.Tags,
			Properties: iothub.UpdateProperties{
				Desired: state.Desired,
			},
			ETag:    twin.ETag,
			Replace: true,
		}
		err = a.hub.UpdateDeviceTwin(ctx, cs, deviceID, update)
	}
	if errHTTP, ok := err.(client.HTTPError); ok && errHTTP.Code == http.StatusPreconditionFailed {
		return nil, ErrDeviceStateConflict
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to update the device twin")
	}
	return a.GetDeviceStateIoTHub(ctx, deviceID, integration)
}
