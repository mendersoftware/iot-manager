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

package iotcore

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/aws/aws-sdk-go-v2/service/iotdataplane"

	"github.com/mendersoftware/iot-manager/crypto"
	"github.com/mendersoftware/iot-manager/model"
)

var (
	ErrDeviceNotFound    = errors.New("device not found")
	ErrDeviceIncosistent = errors.New("device is not consistent")
)

//nolint:lll
//go:generate ../../utils/mockgen.sh
type Client interface {
	GetDeviceShadow(ctx context.Context, creds model.AWSCredentials, id string) (*DeviceShadow, error)
	UpdateDeviceShadow(
		ctx context.Context,
		creds model.AWSCredentials,
		deviceID string,
		update DeviceShadowUpdate,
	) (*DeviceShadow, error)
	GetDevice(ctx context.Context, creds model.AWSCredentials, deviceID string) (*Device, error)
	UpsertDevice(ctx context.Context, creds model.AWSCredentials, deviceID string, device *Device, policy string) (*Device, error)
	DeleteDevice(ctx context.Context, creds model.AWSCredentials, deviceID string) error
}

type client struct{}

func NewClient() Client {
	return &client{}
}

func getAWSConfig(creds model.AWSCredentials) (*aws.Config, error) {
	err := creds.Validate()
	if err != nil {
		return nil, err
	}

	appCreds := credentials.NewStaticCredentialsProvider(
		*creds.AccessKeyID,
		string(*creds.SecretAccessKey),
		"",
	)
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(*creds.Region),
		config.WithCredentialsProvider(appCreds),
	)
	return &cfg, err
}

func (c *client) GetDevice(
	ctx context.Context,
	creds model.AWSCredentials,
	deviceID string,
) (*Device, error) {
	cfg, err := getAWSConfig(creds)
	if err != nil {
		return nil, err
	}
	svc := iot.NewFromConfig(*cfg)

	resp, err := svc.DescribeThing(ctx,
		&iot.DescribeThingInput{
			ThingName: aws.String(deviceID),
		})

	var device *Device
	var respListThingPrincipals *iot.ListThingPrincipalsOutput
	if err == nil {
		device = &Device{
			ID:      *resp.ThingId,
			Name:    *resp.ThingName,
			Version: resp.Version,
			Status:  StatusDisabled,
		}
		respListThingPrincipals, err = svc.ListThingPrincipals(ctx,
			&iot.ListThingPrincipalsInput{
				ThingName: aws.String(deviceID),
			})
	}

	if err == nil {
		if len(respListThingPrincipals.Principals) > 1 {
			err = ErrDeviceIncosistent
		}
	}

	if err == nil {
		for _, principal := range respListThingPrincipals.Principals {
			parts := strings.Split(principal, "/")
			certificateID := parts[len(parts)-1]

			cert, err := svc.DescribeCertificate(ctx, &iot.DescribeCertificateInput{
				CertificateId: aws.String(certificateID),
			})
			if err != nil {
				return nil, err
			}
			device.CertificateID = certificateID
			if cert.CertificateDescription.Status == types.CertificateStatusActive {
				device.Status = StatusEnabled
			}
		}
	}

	var notFoundErr *types.ResourceNotFoundException
	if errors.As(err, &notFoundErr) {
		err = ErrDeviceNotFound
	}

	return device, err
}

func policyName(deviceID string) string {
	return deviceID + "-policy"
}

func (c *client) UpsertDevice(ctx context.Context,
	creds model.AWSCredentials,
	deviceID string,
	device *Device,
	policy string,
) (*Device, error) {

	cfg, err := getAWSConfig(creds)
	if err != nil {
		return nil, err
	}
	svc := iot.NewFromConfig(*cfg)

	awsDevice, err := c.GetDevice(ctx, creds, deviceID)
	if err == nil && awsDevice != nil {
		cert, err := svc.DescribeCertificate(ctx, &iot.DescribeCertificateInput{
			CertificateId: aws.String(awsDevice.CertificateID),
		})
		if err == nil {
			newStatus := types.CertificateStatusInactive
			awsDevice.Status = StatusDisabled
			if device.Status == StatusEnabled {
				newStatus = types.CertificateStatusActive
				awsDevice.Status = StatusEnabled
			}

			if cert.CertificateDescription.Status != newStatus {
				paramsUpdateCertificate := &iot.UpdateCertificateInput{
					CertificateId: aws.String(awsDevice.CertificateID),
					NewStatus:     types.CertificateStatus(newStatus),
				}
				_, err = svc.UpdateCertificate(ctx, paramsUpdateCertificate)
			}
		}

		return awsDevice, err
	} else if err == ErrDeviceNotFound {
		err = nil
	}

	var resp *iot.CreateThingOutput
	if err == nil {
		resp, err = svc.CreateThing(ctx,
			&iot.CreateThingInput{
				ThingName: aws.String(deviceID),
			})
	}

	var respPolicy *iot.CreatePolicyOutput
	if err == nil {
		respPolicy, err = svc.CreatePolicy(ctx,
			&iot.CreatePolicyInput{
				PolicyDocument: aws.String(policy),
				PolicyName:     aws.String(policyName(deviceID)),
			})
	}

	var privKey *ecdsa.PrivateKey
	if err == nil {
		privKey, err = crypto.NewPrivateKey()
	}

	var csr []byte
	if err == nil {
		csr, err = crypto.NewCertificateSigningRequest(deviceID, privKey)
	}

	var respCert *iot.CreateCertificateFromCsrOutput
	if err == nil {
		respCert, err = svc.CreateCertificateFromCsr(ctx,
			&iot.CreateCertificateFromCsrInput{
				CertificateSigningRequest: aws.String(string(csr)),
				SetAsActive:               *aws.Bool(device.Status == StatusEnabled),
			})
	}

	if err == nil {
		_, err = svc.AttachPolicy(ctx,
			&iot.AttachPolicyInput{
				PolicyName: respPolicy.PolicyName,
				Target:     respCert.CertificateArn,
			})
	}

	if err == nil {
		_, err = svc.AttachThingPrincipal(ctx,
			&iot.AttachThingPrincipalInput{
				Principal: respCert.CertificateArn,
				ThingName: aws.String(deviceID),
			})
	}

	var deviceResp *Device
	if err == nil {
		deviceResp = &Device{
			ID:          *resp.ThingId,
			Name:        *resp.ThingName,
			Status:      device.Status,
			PrivateKey:  string(crypto.PrivateKeyToPem(privKey)),
			Certificate: *respCert.CertificatePem,
		}
	}
	return deviceResp, err
}

func (c *client) DeleteDevice(
	ctx context.Context,
	creds model.AWSCredentials,
	deviceID string,
) error {
	cfg, err := getAWSConfig(creds)
	if err != nil {
		return err
	}
	svc := iot.NewFromConfig(*cfg)

	respDescribe, err := svc.DescribeThing(ctx,
		&iot.DescribeThingInput{
			ThingName: aws.String(deviceID),
		})

	var respListThingPrincipals *iot.ListThingPrincipalsOutput
	if err == nil {
		respListThingPrincipals, err = svc.ListThingPrincipals(ctx,
			&iot.ListThingPrincipalsInput{
				ThingName: aws.String(deviceID),
			})
	}

	if err == nil {
		for _, principal := range respListThingPrincipals.Principals {
			_, err := svc.DetachThingPrincipal(ctx,
				&iot.DetachThingPrincipalInput{
					Principal: aws.String(principal),
					ThingName: aws.String(deviceID),
				})
			var certificateID string
			if err == nil {
				parts := strings.SplitAfter(principal, "/")
				certificateID = parts[len(parts)-1]

				_, err = svc.UpdateCertificate(ctx,
					&iot.UpdateCertificateInput{
						CertificateId: aws.String(certificateID),
						NewStatus:     types.CertificateStatusInactive,
					})
			}
			if err == nil {
				_, err = svc.DeleteCertificate(ctx,
					&iot.DeleteCertificateInput{
						CertificateId: aws.String(certificateID),
						ForceDelete:   *aws.Bool(true),
					})
			}
			if err != nil {
				break
			}
		}
	}

	if err == nil {
		_, err = svc.DeleteThing(ctx,
			&iot.DeleteThingInput{
				ThingName:       aws.String(deviceID),
				ExpectedVersion: aws.Int64(respDescribe.Version),
			})
	}

	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			err = ErrDeviceNotFound
		}
		return err
	}

	if err == nil {
		_, err = svc.DeletePolicy(ctx,
			&iot.DeletePolicyInput{
				PolicyName: aws.String(policyName(deviceID)),
			})
	}

	var notFoundErr *types.ResourceNotFoundException
	if errors.As(err, &notFoundErr) {
		err = ErrDeviceNotFound
	}

	return err
}

func (c *client) GetDeviceShadow(
	ctx context.Context,
	creds model.AWSCredentials,
	deviceID string,
) (*DeviceShadow, error) {
	cfg, err := getAWSConfig(creds)
	if err != nil {
		return nil, err
	}
	svc := iotdataplane.NewFromConfig(*cfg)
	shadow, err := svc.GetThingShadow(
		ctx,
		&iotdataplane.GetThingShadowInput{
			ThingName: aws.String(deviceID),
		},
	)
	if err != nil {
		var httpResponseErr *awshttp.ResponseError
		if errors.As(err, &httpResponseErr) {
			if httpResponseErr.HTTPStatusCode() == http.StatusNotFound {
				_, errDevice := c.GetDevice(ctx, creds, deviceID)
				if errDevice == ErrDeviceNotFound {
					err = ErrDeviceNotFound
				} else {
					return &DeviceShadow{
						Payload: model.DeviceState{
							Desired:  map[string]interface{}{},
							Reported: make(map[string]interface{}),
						},
					}, nil
				}
			}
		}
		return nil, err
	}
	var devShadow DeviceShadow
	err = json.Unmarshal(shadow.Payload, &devShadow)
	if err != nil {
		return nil, err
	}
	return &devShadow, nil

}

func (c *client) UpdateDeviceShadow(
	ctx context.Context,
	creds model.AWSCredentials,
	deviceID string,
	update DeviceShadowUpdate,
) (*DeviceShadow, error) {
	cfg, err := getAWSConfig(creds)
	if err != nil {
		return nil, err
	}
	svc := iotdataplane.NewFromConfig(*cfg)
	payloadUpdate, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}
	updated, err := svc.UpdateThingShadow(
		ctx,
		&iotdataplane.UpdateThingShadowInput{
			Payload:   payloadUpdate,
			ThingName: aws.String(deviceID),
		},
	)
	if err != nil {
		var httpResponseErr *awshttp.ResponseError
		if errors.As(err, &httpResponseErr) {
			if httpResponseErr.HTTPStatusCode() == http.StatusNotFound {
				err = ErrDeviceNotFound
			}
		}
		return nil, err
	}
	var shadow DeviceShadow
	err = json.Unmarshal(updated.Payload, &shadow)
	if err != nil {
		return nil, err
	}
	return &shadow, nil

}
