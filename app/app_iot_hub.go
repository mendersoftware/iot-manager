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

	"github.com/pkg/errors"

	"github.com/mendersoftware/iot-manager/client/iothub"
	"github.com/mendersoftware/iot-manager/model"
)

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
	update := &iothub.DeviceTwinUpdate{
		Properties: iothub.UpdateProperties{
			Desired: state.Desired,
		},
	}
	err := a.hub.UpdateDeviceTwin(ctx, cs, deviceID, update)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update the device twin")
	}
	return a.GetDeviceStateIoTHub(ctx, deviceID, integration)
}
