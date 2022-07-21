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
	"net/url"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Integration struct {
	ID          uuid.UUID   `json:"id" bson:"_id"`
	Provider    Provider    `json:"provider" bson:"provider"`
	Credentials Credentials `json:"credentials" bson:"credentials"`
}

func (itg Integration) Validate() error {
	return validation.ValidateStruct(&itg,
		validation.Field(&itg.ID),
		validation.Field(&itg.Provider, validation.Required),
		validation.Field(&itg.Credentials),
	)
}

type CredentialType string

const (
	CredentialTypeAWS CredentialType = "aws"
	CredentialTypeSAS CredentialType = "sas"
)

var credentialTypeRule = validation.In(
	CredentialTypeAWS,
	CredentialTypeSAS,
)

func (typ CredentialType) Validate() error {
	return credentialTypeRule.Validate(typ)
}

type Credentials struct {
	Type CredentialType `json:"type" bson:"type"`

	// AWS Iot Core
	AccessKeyID     *string `json:"access_key_id,omitempty" bson:"access_key_id,omitempty"`
	SecretAccessKey *string `json:"secret_access_key,omitempty" bson:"secret_access_key,omitempty"`
	EndpointURL     *string `json:"endpoint_url,omitempty" bson:"endpoint_url,omitempty"`
	DevicePolicyARN *string `json:"device_policy_arn,omitempty" bson:"device_policy_arn,omitempty"`

	// Azure IoT Hub
	//nolint:lll
	ConnectionString *ConnectionString `json:"connection_string,omitempty" bson:"connection_string,omitempty"`
}

func validateHostname(value interface{}) error {
	c, ok := value.(*string)
	if !ok {
		return errors.New("value is not a string")
	}
	parsed, err := url.Parse(*c)
	if err != nil {
		return err
	}
	return trustedHostnames.Validate(parsed.Hostname())
}

func (s Credentials) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Type, validation.Required),
		validation.Field(&s.ConnectionString,
			validation.When(s.Type == CredentialTypeSAS, validation.Required)),
		validation.Field(&s.AccessKeyID,
			validation.When(s.Type == CredentialTypeAWS, validation.Required)),
		validation.Field(&s.SecretAccessKey,
			validation.When(s.Type == CredentialTypeAWS, validation.Required)),
		validation.Field(&s.EndpointURL,
			validation.When(s.Type == CredentialTypeAWS, validation.Required,
				validation.By(validateHostname))),
		validation.Field(&s.DevicePolicyARN,
			validation.When(s.Type == CredentialTypeAWS, validation.Required)),
	)
}

type IntegrationFilter struct {
	Skip     int64
	Limit    int64
	Provider Provider
	IDs      []uuid.UUID
}
