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

import "context"

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

type DeviceTwin struct {
	AuthenticationType string             `json:"authenticationType,omitempty"`
	Capabilities       DeviceCapabilities `json:"capabilities,omitempty"`

	CloudToDeviceMessageCount int64 `json:"cloudToDeviceMessageCount,omitempty"`

	ConnectionState  string                 `json:"connectionState,omitempty"`
	DeviceEtag       string                 `json:"deviceEtag,omitempty"`
	DeviceID         string                 `json:"deviceId,omitempty"`
	DeviceScope      string                 `json:"deviceScope,omitempty"`
	ETag             string                 `json:"etag,omitempty"`
	LastActivityTime string                 `json:"lastActivityTime,omitempty"`
	ModuleID         string                 `json:"moduleId,omitempty"`
	Properties       TwinProperties         `json:"properties,omitempty"`
	Status           string                 `json:"status,omitempty"`
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
