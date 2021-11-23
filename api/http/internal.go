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

package http

import (
	"net/http"

	"github.com/mendersoftware/azure-iot-manager/app"

	"github.com/gin-gonic/gin"
	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/rest.utils"
	"github.com/pkg/errors"
)

const (
	ParamTenantID = "tenant_id"
	ParamDeviceID = "device_id"
)

type InternalHandler APIHandler

// POST /tenants/:tenant_id/devices
// code: 202 - device provisioned to iothub
//       204 - nothing happened
//       500 - internal server error
func (h *InternalHandler) ProvisionDevice(c *gin.Context) {
	var device struct {
		ID string `json:"device_id"`
	}
	tenantID := c.Param(ParamTenantID)
	if err := c.ShouldBindJSON(&device); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"))
		return
	} else if device.ID == "" {
		rest.RenderError(c, http.StatusBadRequest, errors.New("missing device ID"))
		return
	}

	ctx := identity.WithContext(c.Request.Context(), &identity.Identity{
		Subject: device.ID,
		Tenant:  tenantID,
	})
	err := h.app.ProvisionDevice(ctx, device.ID)
	switch errors.Cause(err) {
	case nil:
		c.Status(http.StatusAccepted)
	case app.ErrNoConnectionString:
		c.Status(http.StatusNoContent)
	default:
		rest.RenderError(c, http.StatusInternalServerError, err)
	}
}

func (h *InternalHandler) DecomissionDevice(c *gin.Context) {
	deviceID := c.Param(ParamDeviceID)
	tenantID := c.Param(ParamTenantID)

	ctx := identity.WithContext(c.Request.Context(), &identity.Identity{
		Subject: deviceID,
		Tenant:  tenantID,
	})
	err := h.app.DeleteIOTHubDevice(ctx, deviceID)
	switch errors.Cause(err) {
	case nil:
		c.Status(http.StatusAccepted)
	case app.ErrNoConnectionString:
		c.Status(http.StatusNoContent)
	default:
		rest.RenderError(c, http.StatusInternalServerError, err)
	}
}
