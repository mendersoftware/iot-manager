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
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iot"

	"github.com/mendersoftware/iot-manager/crypto"
)

var (
	ErrDeviceNotFound    = errors.New("device not found")
	ErrDeviceIncosistent = errors.New("device is not consistent")
)

//nolint:lll
//go:generate ../../utils/mockgen.sh
type Client interface {
	GetDevice(ctx context.Context, sess *session.Session, deviceID string) (*Device, error)
	UpsertDevice(ctx context.Context, sess *session.Session, deviceID string, device *Device, policy string) (*Device, error)
	DeleteDevice(ctx context.Context, sess *session.Session, deviceID string) error
}

type client struct{}

func NewClient() Client {
	return &client{}
}

func (c *client) GetDevice(
	ctx context.Context,
	sess *session.Session,
	deviceID string,
) (*Device, error) {
	svc := iot.New(sess)

	resp, err := svc.DescribeThingWithContext(ctx,
		&iot.DescribeThingInput{
			ThingName: aws.String(deviceID),
		})

	var device *Device
	var respListThingPrincipals *iot.ListThingPrincipalsOutput
	if err == nil {
		device = &Device{
			ID:      *resp.ThingId,
			Name:    *resp.ThingName,
			Version: *resp.Version,
			Status:  StatusDisabled,
		}
		respListThingPrincipals, err = svc.ListThingPrincipalsWithContext(ctx,
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
			parts := strings.Split(*principal, "/")
			certificateID := parts[len(parts)-1]

			cert, err := svc.DescribeCertificateWithContext(ctx, &iot.DescribeCertificateInput{
				CertificateId: aws.String(certificateID),
			})
			if err != nil {
				return nil, err
			}
			device.CertificateID = certificateID
			if *cert.CertificateDescription.Status == iot.CertificateStatusActive {
				device.Status = StatusEnabled
			}
		}
	}

	if _, ok := err.(*iot.ResourceNotFoundException); ok {
		err = ErrDeviceNotFound
	}

	return device, err
}

func policyName(deviceID string) string {
	return deviceID + "-policy"
}

func (c *client) UpsertDevice(ctx context.Context,
	sess *session.Session,
	deviceID string,
	device *Device,
	policy string,
) (*Device, error) {
	svc := iot.New(sess)

	awsDevice, err := c.GetDevice(ctx, sess, deviceID)
	if err == nil && awsDevice != nil {
		cert, err := svc.DescribeCertificateWithContext(ctx, &iot.DescribeCertificateInput{
			CertificateId: aws.String(awsDevice.CertificateID),
		})
		if err == nil {
			newStatus := iot.CertificateStatusInactive
			awsDevice.Status = StatusDisabled
			if device.Status == StatusEnabled {
				newStatus = iot.CertificateStatusActive
				awsDevice.Status = StatusEnabled
			}

			if *cert.CertificateDescription.Status != newStatus {
				paramsUpdateCertificate := &iot.UpdateCertificateInput{
					CertificateId: aws.String(awsDevice.CertificateID),
					NewStatus:     aws.String(newStatus),
				}
				_, err = svc.UpdateCertificateWithContext(ctx, paramsUpdateCertificate)
			}
		}

		return awsDevice, err
	} else if err == ErrDeviceNotFound {
		err = nil
	}

	var resp *iot.CreateThingOutput
	if err == nil {
		resp, err = svc.CreateThingWithContext(ctx,
			&iot.CreateThingInput{
				ThingName: aws.String(deviceID),
			})
	}

	var respPolicy *iot.CreatePolicyOutput
	if err == nil {
		respPolicy, err = svc.CreatePolicyWithContext(ctx,
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
		respCert, err = svc.CreateCertificateFromCsrWithContext(ctx,
			&iot.CreateCertificateFromCsrInput{
				CertificateSigningRequest: aws.String(string(csr)),
				SetAsActive:               aws.Bool(device.Status == StatusEnabled),
			})
	}

	if err == nil {
		_, err = svc.AttachPolicyWithContext(ctx,
			&iot.AttachPolicyInput{
				PolicyName: respPolicy.PolicyName,
				Target:     respCert.CertificateArn,
			})
	}

	if err == nil {
		_, err = svc.AttachThingPrincipalWithContext(ctx,
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

func (c *client) DeleteDevice(ctx context.Context, sess *session.Session, deviceID string) error {
	svc := iot.New(sess)

	respDescribe, err := svc.DescribeThingWithContext(ctx,
		&iot.DescribeThingInput{
			ThingName: aws.String(deviceID),
		})

	var respListThingPrincipals *iot.ListThingPrincipalsOutput
	if err == nil {
		respListThingPrincipals, err = svc.ListThingPrincipalsWithContext(ctx,
			&iot.ListThingPrincipalsInput{
				ThingName: aws.String(deviceID),
			})
	}

	if err == nil {
		for _, principal := range respListThingPrincipals.Principals {
			_, err := svc.DetachThingPrincipalWithContext(ctx,
				&iot.DetachThingPrincipalInput{
					Principal: aws.String(*principal),
					ThingName: aws.String(deviceID),
				})
			var certificateID string
			if err == nil {
				parts := strings.SplitAfter(*principal, "/")
				certificateID = parts[len(parts)-1]

				_, err = svc.UpdateCertificateWithContext(ctx,
					&iot.UpdateCertificateInput{
						CertificateId: aws.String(certificateID),
						NewStatus:     aws.String(iot.CertificateStatusInactive),
					})
			}
			if err == nil {
				_, err = svc.DeleteCertificateWithContext(ctx,
					&iot.DeleteCertificateInput{
						CertificateId: aws.String(certificateID),
						ForceDelete:   aws.Bool(true),
					})
			}
			if err != nil {
				break
			}
		}
	}

	if err == nil {
		_, err = svc.DeleteThingWithContext(ctx,
			&iot.DeleteThingInput{
				ThingName:       aws.String(deviceID),
				ExpectedVersion: aws.Int64(*respDescribe.Version),
			})
	}

	if err != nil {
		if _, ok := err.(*iot.ResourceNotFoundException); ok {
			err = ErrDeviceNotFound
		}
		return err
	}

	if err == nil {
		_, err = svc.DeletePolicyWithContext(ctx,
			&iot.DeletePolicyInput{
				PolicyName: aws.String(policyName(deviceID)),
			})
	}

	if _, ok := err.(*iot.ResourceNotFoundException); ok {
		err = ErrDeviceNotFound
	}

	return err
}
