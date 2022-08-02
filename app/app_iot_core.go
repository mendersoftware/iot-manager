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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"

	"github.com/mendersoftware/iot-manager/client/iotcore"
	"github.com/mendersoftware/iot-manager/model"
)

func getRegionFromEndpoint(endpointURL string) string {
	// expected endpoint: https://random-id.iot.us-east-1.amazonaws.com
	if strings.Contains(endpointURL, ".amazonaws.com") {
		parts := strings.Split(endpointURL, ".")
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	return ""
}

func getIoTCoreSession(integration model.Integration) (*session.Session, error) {
	accessKeyID := integration.Credentials.AccessKeyID
	secretAccessKey := integration.Credentials.SecretAccessKey
	endpointURL := integration.Credentials.EndpointURL
	if accessKeyID == nil || secretAccessKey == nil || endpointURL == nil {
		return nil, ErrNoCredentials
	}

	region := getRegionFromEndpoint(*endpointURL)
	return session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(*accessKeyID, string(*secretAccessKey), ""),
	})
}

func (a *app) provisionIoTCoreDevice(
	ctx context.Context,
	deviceID string,
	integration model.Integration,
	device *iotcore.Device,
) error {
	sess, err := getIoTCoreSession(integration)
	if err != nil {
		return err
	}

	dev, err := a.iotcoreClient.UpsertDevice(ctx, sess, deviceID, device,
		*integration.Credentials.DevicePolicyDocument)
	if err != nil {
		return errors.Wrap(err, "failed to update iotcore devices")
	}

	if dev.Certificate != "" && dev.PrivateKey != "" {
		err = a.wf.ProvisionExternalDevice(ctx, dev.ID, map[string]string{
			confKeyAWSCertificate: dev.Certificate,
			confKeyAWSPrivateKey:  dev.PrivateKey,
		})
		if err != nil {
			return errors.Wrap(err, "failed to submit iotcore credentials to deviceconfig")
		}
	}

	return nil
}

func (a *app) setDeviceStatusIoTCore(ctx context.Context, deviceID string, status model.Status,
	integration model.Integration) error {
	sess, err := getIoTCoreSession(integration)
	if err != nil {
		return err
	}
	_, err = a.iotcoreClient.UpsertDevice(
		ctx,
		sess,
		deviceID,
		&iotcore.Device{
			Status: iotcore.NewStatusFromMenderStatus(status),
		},
		*integration.Credentials.DevicePolicyDocument,
	)
	return err
}

func (a *app) decommissionIoTCoreDevice(ctx context.Context, deviceID string,
	integration model.Integration) error {
	sess, err := getIoTCoreSession(integration)
	if err != nil {
		return err
	}
	err = a.iotcoreClient.DeleteDevice(ctx, sess, deviceID)
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
	return nil
}
