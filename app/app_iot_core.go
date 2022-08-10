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

	"github.com/pkg/errors"

	"github.com/mendersoftware/iot-manager/client/iotcore"
	"github.com/mendersoftware/iot-manager/model"
)

func assertAWSIntegration(integration model.Integration) error {
	if err := integration.Validate(); err != nil {
		return ErrNoCredentials
	} else if integration.Credentials.Type != model.CredentialTypeAWS {
		return ErrNoCredentials
	}
	return nil
}

func (a *app) provisionIoTCoreDevice(
	ctx context.Context,
	deviceID string,
	integration model.Integration,
	device *iotcore.Device,
) error {
	if err := assertAWSIntegration(integration); err != nil {
		return err
	}

	dev, err := a.iotcoreClient.UpsertDevice(ctx,
		*integration.Credentials.AWSCredentials,
		deviceID,
		device,
		*integration.Credentials.AWSCredentials.DevicePolicyDocument)
	if err != nil {
		return errors.Wrap(err, "failed to update iotcore devices")
	}

	err = a.deployConfiguration(ctx, deviceID, dev)
	return err
}

func (a *app) setDeviceStatusIoTCore(ctx context.Context, deviceID string, status model.Status,
	integration model.Integration) error {
	if err := assertAWSIntegration(integration); err != nil {
		return err
	}
	_, err := a.iotcoreClient.UpsertDevice(
		ctx,
		*integration.Credentials.AWSCredentials,
		deviceID,
		&iotcore.Device{
			Status: iotcore.NewStatusFromMenderStatus(status),
		},
		*integration.Credentials.AWSCredentials.DevicePolicyDocument,
	)
	return err

}

func (a *app) deployConfiguration(ctx context.Context, deviceID string, dev *iotcore.Device) error {
	if dev.Certificate != "" && dev.PrivateKey != "" {
		err := a.wf.ProvisionExternalDevice(ctx, deviceID, map[string]string{
			confKeyAWSCertificate: dev.Certificate,
			confKeyAWSPrivateKey:  dev.PrivateKey,
		})
		if err != nil {
			return errors.Wrap(err, "failed to submit iotcore credentials to deviceconfig")
		}
	}
	return nil
}

func (a *app) decommissionIoTCoreDevice(ctx context.Context, deviceID string,
	integration model.Integration) error {
	if err := assertAWSIntegration(integration); err != nil {
		return err
	}
	err := a.iotcoreClient.DeleteDevice(ctx, *integration.Credentials.AWSCredentials, deviceID)
	if err != nil {
		return errors.Wrap(err, "failed to delete IoT Core device")
	}
	return nil
}

func (a *app) syncIoCoreDevices(
	ctx context.Context,
	deviceIDs []string,
	integration model.Integration,
	failEarly bool,
) error {
	if err := assertAWSIntegration(integration); err != nil {
		return err
	}
	return nil
}

func (a *app) GetDeviceStateIoTCore(
	ctx context.Context,
	deviceID string,
	integration *model.Integration,
) (*model.DeviceState, error) {
	if err := assertAWSIntegration(*integration); err != nil {
		return nil, err
	}
	shadow, err := a.iotcoreClient.GetDeviceShadow(
		ctx,
		*integration.Credentials.AWSCredentials,
		deviceID,
	)
	if err != nil {
		if err == iotcore.ErrDeviceNotFound {
			return nil, nil
		} else {
			return nil, errors.Wrap(err, "failed to get the device shadow")
		}
	}
	return &shadow.Payload, nil
}

func (a *app) SetDeviceStateIoTCore(
	ctx context.Context,
	deviceID string,
	integration *model.Integration,
	state *model.DeviceState,
) (*model.DeviceState, error) {
	if state == nil {
		return nil, nil
	}
	if err := assertAWSIntegration(*integration); err != nil {
		return nil, err
	}
	shadow, err := a.iotcoreClient.UpdateDeviceShadow(
		ctx,
		*integration.Credentials.AWSCredentials,
		deviceID,
		iotcore.DeviceShadowUpdate{
			State: iotcore.DesiredState{
				Desired: state.Desired,
			},
		},
	)
	if err != nil {
		if err == iotcore.ErrDeviceNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &shadow.Payload, nil
}
