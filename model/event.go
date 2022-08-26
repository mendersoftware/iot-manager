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
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

type Event struct {
	ID             uuid.UUID      `json:"id" bson:"_id"`
	Type           EventType      `json:"type" bson:"type"`
	Data           interface{}    `json:"data" bson:"data"`
	DeliveryStatus DeliveryStatus `json:"delivery_status" bson:"delivery_status"`
	// EventTS is the timestamp when the event has been produced.
	EventTS time.Time `json:"time" bson:"event_ts"`
	// ExpireTS contains the timestamp when this event entry expires from the
	// database.
	ExpireTS time.Time `json:"-" bson:"expire_ts,omitempty"`
}

func (event Event) Validate() error {
	return validation.ValidateStruct(&event,
		validation.Field(&event.ID),
		validation.Field(&event.Type, validation.Required),
	)
}

type DeliveryStatus string

const (
	DeliveryStatusDelivered    DeliveryStatus = "delivered"
	DeliveryStatusNotDelivered DeliveryStatus = "not-delivered"
)

type EventType string

const (
	EventTypeDeviceProvisioned    EventType = "device-provisioned"
	EventTypeDeviceDecommissioned EventType = "device-decommissioned"
	EventTypeDeviceStatusChanged  EventType = "device-status-changed"
)

var eventTypeRule = validation.In(
	EventTypeDeviceProvisioned,
	EventTypeDeviceDecommissioned,
)

func (typ EventType) Validate() error {
	return eventTypeRule.Validate(typ)
}

type EventsFilter struct {
	Skip  int64
	Limit int64
}

// data objects for different event types
type EventDeviceDecommissionedData struct {
	DeviceID string `json:"device_id" bson:"device_id"`
}

type EventDeviceProvisionedData struct {
	DeviceID string `json:"device_id" bson:"device_id"`
}

type EventDeviceStatusChangedData struct {
	DeviceID  string `json:"device_id" bson:"device_id"`
	NewStatus Status `json:"new_status" bson:"new_status"`
}
