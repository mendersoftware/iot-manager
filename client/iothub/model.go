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

package iothub

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/mendersoftware/azure-iot-manager/model"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

type Key []byte

func (k Key) MarshalText() ([]byte, error) {
	n := base64.StdEncoding.EncodedLen(len(k))
	ret := make([]byte, n)
	base64.StdEncoding.Encode(ret, k)
	return ret, nil
}

type SymmetricKey struct {
	Primary   Key `json:"primaryKey"`
	Secondary Key `json:"secondaryKey"`
}

type AuthType string

const (
	AuthTypeSymmetric   AuthType = "sas"
	AuthTypeCertificate AuthType = "certificate"
	AuthTypeNone        AuthType = "none"
	AuthTypeAuthority   AuthType = "Authority"
	AuthTypeSelfSigned  AuthType = "selfSigned"
)

type Auth struct {
	Type            AuthType
	*SymmetricKey   `json:"symmetricKey,omitempty"`
	*X509ThumbPrint `json:"x509Thumbprint,omitempty"`
}

func NewSymmetricAuth() (*Auth, error) {
	var primKey, secKey [48]byte
	_, err := io.ReadFull(rand.Reader, primKey[:])
	if err != nil {
		return nil, err
	}
	_, err = io.ReadFull(rand.Reader, secKey[:])
	if err != nil {
		return nil, err
	}
	return &Auth{
		Type: AuthTypeSymmetric,
		SymmetricKey: &SymmetricKey{
			Primary:   Key(primKey[:]),
			Secondary: Key(secKey[:]),
		},
	}, nil
}

type Status string

const (
	StatusEnabled  Status = "enabled"
	StatusDisabled Status = "disabled"
)

func (s *Status) UnmarshalText(b []byte) error {
	*s = Status(bytes.ToLower(b))
	return s.Validate()
}

var validateStatus = validation.In(
	StatusEnabled,
	StatusDisabled,
)

func (s Status) Validate() error {
	return validateStatus.Validate(s)
}

type DeviceCapabilities struct {
	IOTEdge bool `json:"iotEdge"`
}

type TwinProperties struct {
	Desired  map[string]interface{} `json:"desired"`
	Reported map[string]interface{} `json:"reported"`
}

type X509ThumbPrint struct {
	Primary   string `json:"primaryThumbprint"`
	Secondary string `json:"secondaryThumbprint"`
}

type Device struct {
	*Auth                  `json:"authentication,omitempty"`
	*DeviceCapabilities    `json:"capabilities,omitempty"`
	C2DMessageCount        int    `json:"cloudToDeviceMessageCount,omitempty"`
	ConnectionState        string `json:"connectionState,omitempty"`
	ConnectionStateUpdated string `json:"connectionStateUpdatedTime,omitempty"`

	DeviceID         string `json:"deviceId"`
	DeviceScope      string `json:"deviceScope,omitempty"`
	ETag             string `json:"etag,omitempty"`
	GenerationID     string `json:"generationId,omitempty"`
	LastActivityTime string `json:"lastActivityTime,omitempty"`
	Status           Status `json:"status,omitempty"`
	StatusReason     string `json:"statusReason,omitempty"`
	StatusUpdateTime string `json:"statusUpdateTime,omitempty"`
}

func mergeDevices(devices ...*Device) *Device {
	var device *Device
	for _, device = range devices {
		if device != nil {
			break
		}
	}
	if device == nil {
		return new(Device)
	}
	rDevice := reflect.ValueOf(device).Elem()
	for _, dev := range devices {
		if dev == nil {
			continue
		}
		rDev := reflect.ValueOf(*dev)
		for i := 0; i < rDev.NumField(); i++ {
			fDev := rDev.Field(i)
			if fDev.IsZero() {
				continue
			}
			fDevice := rDevice.Field(i)
			fDevice.Set(fDev)
		}
	}
	return device
}

type DeviceTwin struct {
	AuthenticationType string              `json:"authenticationType,omitempty"`
	Capabilities       *DeviceCapabilities `json:"capabilities,omitempty"`

	CloudToDeviceMessageCount int64 `json:"cloudToDeviceMessageCount,omitempty"`

	ConnectionState  string                 `json:"connectionState,omitempty"`
	DeviceEtag       string                 `json:"deviceEtag,omitempty"`
	DeviceID         string                 `json:"deviceId,omitempty"`
	DeviceScope      string                 `json:"deviceScope,omitempty"`
	ETag             string                 `json:"etag,omitempty"`
	LastActivityTime string                 `json:"lastActivityTime,omitempty"`
	ModuleID         string                 `json:"moduleId,omitempty"`
	Properties       TwinProperties         `json:"properties,omitempty"`
	Status           Status                 `json:"status,omitempty"`
	StatusReason     string                 `json:"statusReason,omitempty"`
	StatusUpdateTime string                 `json:"statusUpdateTime,omitempty"`
	Tags             map[string]interface{} `json:"tags,omitempty"`
	Version          int32                  `json:"version,omitempty"`
	X509ThumbPrint   X509ThumbPrint         `json:"x509Thumbprint,omitempty"`
}

type UpdateProperties struct {
	Desired map[string]interface{} `json:"desired"`
}

type DeviceTwinUpdate struct {
	Properties UpdateProperties       `json:"properties,omitempty"`
	Tags       map[string]interface{} `json:"tags,omitempty"`
	Replace    bool                   `json:"-"`
}

type Cursor interface {
	Next(ctx context.Context) bool
	Decode(v interface{}) error
}

type cursor struct {
	err     error
	buf     bytes.Buffer
	mut     *sync.Mutex
	cs      *model.ConnectionString
	client  *http.Client
	req     *http.Request
	dec     *json.Decoder
	current json.RawMessage
}

func (cur *cursor) fetchPage(ctx context.Context) error {
	cur.buf.Reset()
	req := cur.req.WithContext(ctx)
	req.Body, _ = cur.req.GetBody()
	var expireAt time.Time
	if dl, ok := ctx.Deadline(); ok {
		expireAt = dl
	} else {
		expireAt = time.Now().Add(defaultTTL)
	}
	req.Header.Set(hdrKeyAuthorization, cur.cs.Authorization(expireAt))
	rsp, err := cur.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "iothub: failed to execute request")
	}
	defer rsp.Body.Close()
	if rsp.StatusCode >= 400 {
		return errors.Wrapf(err,
			"iothub: received unexpected status code from server: %s",
			rsp.Status,
		)
	}
	_, err = io.Copy(&cur.buf, rsp.Body)
	if err != nil {
		return errors.Wrap(err, "iothub: failed to buffer HTTP response")
	}
	cur.req.Header.Set(hdrKeyContToken, rsp.Header.Get(hdrKeyContToken))
	cur.dec = json.NewDecoder(&cur.buf)
	tkn, err := cur.dec.Token()
	if err != nil {
		return errors.Wrap(err, "iothub: failed to decode response from hub")
	} else if tkn != json.Delim('[') {
		return errors.Wrap(err, "iothub: unexpected json response from hub")
	}
	return nil
}

func (cur *cursor) Next(ctx context.Context) bool {
	if cur.err != nil {
		return false
	}
	cur.mut.Lock()
	defer cur.mut.Unlock()
	if cur.dec == nil || !cur.dec.More() {
		if cur.req.Header.Get(hdrKeyContToken) == "" {
			cur.err = io.EOF
			return false
		}
		err := cur.fetchPage(ctx)
		if err != nil {
			cur.err = err
			return false
		}
	}
	err := cur.dec.Decode(&cur.current)
	if err != nil {
		cur.err = errors.Wrap(err, "iothub: failed to retrieve next element")
		return false
	}
	return true
}

func (cur *cursor) Decode(v interface{}) error {
	if cur.err != nil {
		return cur.err
	}
	return json.Unmarshal(cur.current, v)
}
