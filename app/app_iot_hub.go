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

	"github.com/mendersoftware/go-lib-micro/log"

	"github.com/mendersoftware/iot-manager/client"
	"github.com/mendersoftware/iot-manager/client/iothub"
	"github.com/mendersoftware/iot-manager/crypto"
	"github.com/mendersoftware/iot-manager/model"
)

func (a *app) provisionIoTHubDevice(
	ctx context.Context,
	deviceID string,
	integration model.Integration,
	deviceUpdate ...*iothub.Device,
) error {
	cs := integration.Credentials.ConnectionString
	if cs == nil {
		return ErrNoConnectionString
	}

	dev, err := a.hub.UpsertDevice(ctx, cs, deviceID, deviceUpdate...)
	if err != nil {
		if htErr, ok := err.(client.HTTPError); ok {
			switch htErr.Code() {
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

func (a *app) syncIoTHubDevices(
	ctx context.Context,
	deviceIDs []string,
	integration model.Integration,
	failEarly bool,
) error {
	l := log.FromContext(ctx)
	cs := integration.Credentials.ConnectionString

	// Get device authentication
	devAuths, err := a.devauth.GetDevices(ctx, deviceIDs)
	if err != nil {
		return errors.Wrap(err, "app: failed to lookup device authentication")
	}

	statuses := make(map[string]iothub.Status, len(deviceIDs))
	for _, auth := range devAuths {
		statuses[auth.ID] = iothub.NewStatusFromMenderStatus(auth.Status)
	}
	// Find devices that shouldn't exist
	var (
		i int
		j int = len(deviceIDs)
	)
	for i < j {
		id := deviceIDs[i]
		if _, ok := statuses[id]; !ok {
			l.Warnf("Device '%s' does not have an auth set: deleting device", id)
			err := a.DecommissionDevice(ctx, id)
			if err != nil && err != ErrDeviceNotFound {
				err = errors.Wrap(err, "app: failed to decommission device")
				if failEarly {
					return err
				}
				l.Error(err)
			}
			// swap(deviceIDs[i], deviceIDs[j])
			j--
			tmp := deviceIDs[i]
			deviceIDs[i] = deviceIDs[j]
			deviceIDs[j] = tmp
		} else {
			i++
		}
	}

	// Fetch IoT Hub device twins
	hubDevs, err := a.hub.GetDeviceTwins(ctx, cs, deviceIDs[:j])
	if err != nil {
		return errors.Wrap(err, "app: failed to get devices from IoT Hub")
	}

	// Set of device IDs in iot hub
	devicesInHub := make(map[string]struct{}, len(hubDevs))

	// Check if devices (statuses) are in sync
	for _, twin := range hubDevs {
		devicesInHub[twin.DeviceID] = struct{}{}
		if stat, ok := statuses[twin.DeviceID]; ok {
			if stat == twin.Status {
				continue
			}
			l.Warnf("Device '%s' status does not match Mender auth status, updating status",
				twin.DeviceID)
			// Update the device's status
			// NOTE need to fetch device identity first
			dev, err := a.hub.GetDevice(ctx, cs, twin.DeviceID)
			if err != nil {
				err = errors.Wrap(err, "failed to retrieve IoT Hub device identity")
				if failEarly {
					return err
				}
				l.Error(err)
				continue
			}
			dev.Status = stat
			_, err = a.hub.UpsertDevice(ctx, cs, twin.DeviceID, dev)
			if err != nil {
				err = errors.Wrap(err, "failed to update IoT Hub device identity")
				if failEarly {
					return err
				}
				l.Error(err)
			}
		}
	}

	// Find devices not present in IoT Hub
	for id, status := range statuses {
		if _, ok := devicesInHub[id]; !ok {
			l.Warnf("Found device not existing in IoT Hub '%s': provisioning device", id)
			// Device inconsistency
			// Device exist in Mender but not in IoT Hub
			err := a.provisionIoTHubDevice(ctx, id, integration, &iothub.Device{
				DeviceID: id,
				Status:   status,
			})
			if err != nil {
				if failEarly {
					return err
				}
				l.Error(err)
				continue
			}
		}
	}
	return nil
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
	if errHTTP, ok := err.(client.HTTPError); ok &&
		errHTTP.Code() == http.StatusPreconditionFailed {
		return nil, ErrDeviceStateConflict
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to update the device twin")
	}
	return a.GetDeviceStateIoTHub(ctx, deviceID, integration)
}
