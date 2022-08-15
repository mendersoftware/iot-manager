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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/iot-manager/model"
)

func TestStatus(t *testing.T) {
	status := NewStatusFromMenderStatus(model.StatusAccepted)
	assert.Equal(t, StatusEnabled, status)

	status = NewStatusFromMenderStatus(model.StatusRejected)
	assert.Equal(t, StatusDisabled, status)

	statusText := string(StatusDisabled)
	assert.Equal(t, "disabled", statusText)

	err := status.Validate()
	assert.NoError(t, err)

	err = Status("dummy").Validate()
	assert.Error(t, err)
	assert.EqualError(t, err, "must be a valid value")
}
