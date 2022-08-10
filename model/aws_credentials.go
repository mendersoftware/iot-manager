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

package model

import (
	"encoding/json"
	"errors"

	"github.com/mendersoftware/iot-manager/crypto"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type AWSCredentials struct {
	AccessKeyID          *string        `json:"access_key_id,omitempty" bson:"access_key_id,omitempty"`
	SecretAccessKey      *crypto.String `json:"secret_access_key,omitempty" bson:"secret_access_key,omitempty"`
	Region               *string        `json:"region,omitempty" bson:"region,omitempty"`
	DevicePolicyDocument *string        `json:"device_policy_document,omitempty" bson:"device_policy_arn,omitempty"`
}

func (c AWSCredentials) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.AccessKeyID, validation.Required),
		validation.Field(&c.SecretAccessKey, validation.Required),
		validation.Field(&c.Region, validation.Required),
		validation.Field(&c.DevicePolicyDocument, validation.Required, validation.By(validatePolicy)),
	)
}

func validatePolicy(value interface{}) error {
	err := errors.New("value is not a string")
	if c, ok := value.(*string); ok {
		policy := struct {
			Id        string        `json:"Id"`
			Version   string        `json:"Version"`
			Statement []interface{} `json:"Statement"`
		}{}
		err = json.Unmarshal([]byte(*c), &policy)
		if err != nil || policy.Statement == nil {
			err = errors.New("not an AWS IAM policy document")
		} else {
			err = nil
		}
	}
	return err
}
