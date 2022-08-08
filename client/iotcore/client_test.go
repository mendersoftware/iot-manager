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
	"flag"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/iot-manager/model"
)

func init() {
	model.SetTrustedHostnames([]string{"*.iot.*.amazonaws.com", "localhost"})
}

var (
	accessKeyID     string
	secretAccessKey string
	endpointURL     string
)

const testPolicy = `{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"iot:Publish",
				"iot:Receive"
			],
			"Resource": [
				"arn:aws:iot:us-east-1:304194462000:topic/sdk/test/Python"
			]
		},
		{
			"Effect": "Allow",
			"Action": [
				"iot:Subscribe"
			],
			"Resource": [
				"arn:aws:iot:us-east-1:304194462000:topicfilter/sdk/test/Python"
			]
		},
		{
			"Effect": "Allow",
			"Action": [
				"iot:Connect"
			],
			"Resource": [
				"arn:aws:iot:us-east-1:304194462000:client/basicPubSub"
			]
		}
	]
}`

func init() {
	flag.StringVar(&accessKeyID,
		"test.aws-access-key-id",
		"",
		"AWS Access Key ID (overwrite with env var TEST_AWS_ACCESS_KEY_ID).",
	)
	if val, ok := os.LookupEnv("TEST_AWS_ACCESS_KEY_ID"); ok && val != "" {
		accessKeyID = val
	}
	flag.StringVar(&secretAccessKey,
		"test.aws-secret-access-key",
		"",
		"AWS Secret Access Key (overwrite with env var TEST_AWS_SECRET_ACCESS_KEY).",
	)
	if val, ok := os.LookupEnv("TEST_AWS_SECRET_ACCESS_KEY"); ok && val != "" {
		secretAccessKey = val
	}
	flag.StringVar(&endpointURL,
		"test.aws-endpoint-url",
		"",
		"AWS IoT Core Endpoint URL (overwrite with env var TEST_AWS_ENDPOINT_URL).",
	)
	if val, ok := os.LookupEnv("TEST_AWS_ENDPOINT_URL"); ok && val != "" {
		endpointURL = val
	}

	testing.Init()
}

func validAWSSettings(t *testing.T) bool {
	if accessKeyID == "" || secretAccessKey == "" || endpointURL == "" {
		t.Skip("AWS settings not provided or invalid")
		return false
	}
	return true
}

func TestGetDevice(t *testing.T) {
	if !validAWSSettings(t) {
		return
	}

	appCreds := aws.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""))
	cfg, _ := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(*aws.String("us-east-1")),
		config.WithCredentialsProvider(appCreds),
	)

	ctx := context.Background()
	deviceID := uuid.NewString()

	client := NewClient()
	_, err := client.UpsertDevice(ctx, &cfg, deviceID, &Device{}, testPolicy)
	assert.NoError(t, err)

	device, err := client.GetDevice(ctx, &cfg, deviceID)
	assert.NoError(t, err)
	assert.NotNil(t, device)

	assert.Equal(t, deviceID, device.Name)

	_, err = client.GetDevice(ctx, &cfg, "dummy")
	assert.EqualError(t, err, ErrDeviceNotFound.Error())

	err = client.DeleteDevice(ctx, &cfg, device.Name)
	assert.NoError(t, err)

	device, err = client.GetDevice(ctx, &cfg, deviceID)
	assert.EqualError(t, err, ErrDeviceNotFound.Error())
	assert.Nil(t, device)
}

func TestDeleteDevice(t *testing.T) {
	if !validAWSSettings(t) {
		return
	}

	appCreds := aws.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""))
	cfg, _ := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(*aws.String("us-east-1")),
		config.WithCredentialsProvider(appCreds),
	)

	ctx := context.Background()
	deviceID := uuid.NewString()

	client := NewClient()

	device, err := client.UpsertDevice(ctx, &cfg, deviceID, &Device{}, testPolicy)
	assert.NoError(t, err)

	err = client.DeleteDevice(ctx, &cfg, device.Name)
	assert.NoError(t, err)

	err = client.DeleteDevice(ctx, &cfg, device.Name)
	assert.EqualError(t, err, ErrDeviceNotFound.Error())

	device, err = client.GetDevice(ctx, &cfg, deviceID)
	assert.EqualError(t, err, ErrDeviceNotFound.Error())
	assert.Nil(t, device)
}

func TestUpsertDevice(t *testing.T) {
	if !validAWSSettings(t) {
		return
	}

	appCreds := aws.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""))
	cfg, _ := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(*aws.String("us-east-1")),
		config.WithCredentialsProvider(appCreds),
	)

	ctx := context.Background()
	deviceID := uuid.NewString()

	client := NewClient()
	device, err := client.UpsertDevice(ctx, &cfg, deviceID, &Device{
		Status: StatusDisabled,
	}, testPolicy)
	assert.NoError(t, err)
	assert.Equal(t, StatusDisabled, device.Status)

	assert.NotEmpty(t, device.ID)
	assert.NotEmpty(t, device.PrivateKey)
	assert.NotEmpty(t, device.Certificate)

	device.Status = StatusEnabled
	device, err = client.UpsertDevice(ctx, &cfg, deviceID, device, testPolicy)
	assert.NoError(t, err)
	assert.Equal(t, StatusEnabled, device.Status)

	err = client.DeleteDevice(ctx, &cfg, deviceID)
	assert.NoError(t, err)
}
