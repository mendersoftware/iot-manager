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

type Credentials struct {
	Type  string            `json:"type" bson:"type"`
	Creds *ConnectionString `json:"credentials" bson:"credentials"`
}

func (s Credentials) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Creds),
	)
}
