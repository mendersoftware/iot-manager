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

package model

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
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
	CredentialTypeSAS CredentialType = "sas"
)

var credentialTypeRule = validation.In(
	CredentialTypeSAS,
)

func (typ CredentialType) Validate() error {
	return credentialTypeRule.Validate(typ)
}

type Credentials struct {
	Type CredentialType `json:"type" bson:"type"`
	//nolint:lll
	ConnectionString *ConnectionString `json:"connection_string,omitempty" bson:"connection_string,omitempty"`
}

func (s Credentials) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Type, validation.Required),
		validation.Field(&s.ConnectionString,
			validation.When(s.Type == CredentialTypeSAS, validation.Required)),
	)
}
