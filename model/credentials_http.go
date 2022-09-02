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
	"encoding/hex"
	"encoding/json"
	"net"
	"net/url"

	"github.com/mendersoftware/iot-manager/crypto"
	inet "github.com/mendersoftware/iot-manager/internal/net"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type HexSecret crypto.String

func (sec *HexSecret) UnmarshalText(b []byte) error {
	dst := make([]byte, hex.DecodedLen(len(b)))
	n, err := hex.Decode(dst, b)
	if err != nil {
		return err
	}
	*sec = HexSecret(dst[:n])
	return nil
}

func (sec HexSecret) MarshalText() ([]byte, error) {
	return []byte("<omitted>"), nil
}

func (sec HexSecret) MarshalBSON() ([]byte, error) {
	cStr := crypto.String(sec)
	return (&cStr).MarshalBSON()
}

func (sec *HexSecret) UnmarshalBSON(b []byte) error {
	cStr := (*crypto.String)(sec)
	return cStr.UnmarshalBSON(b)
}

type HTTPCredentials struct {
	URL    string     `json:"url,omitempty" bson:"url,omitempty"`
	Secret *HexSecret `json:"secret,omitempty" bson:"secret,omitempty"`

	// private field toggling validation verbosity
	// - only set if unmarshaled from JSON
	validateAddr bool
}

func (cred *HTTPCredentials) UnmarshalJSON(b []byte) error {
	type creds HTTPCredentials
	if err := json.Unmarshal(b, (*creds)(cred)); err != nil {
		return err
	}
	cred.validateAddr = true
	return nil
}

func (cred HTTPCredentials) validateURL(interface{}) error {
	uu, err := url.Parse(cred.URL)
	if err != nil {
		return err
	}
	if !cred.validateAddr {
		return nil
	}
	ips, err := net.LookupIP(uu.Hostname())
	if err != nil {
		return err
	}
	for _, ip := range ips {
		if !inet.IsGlobalUnicast(ip) {
			return net.InvalidAddrError(
				"hostname resolves to reserved address",
			)
		}
	}
	return nil
}

func (cred HTTPCredentials) Validate() error {
	return validation.ValidateStruct(&cred,
		validation.Field(&cred.URL,
			validation.Required,
			validation.By(cred.validateURL),
		),
	)
}
