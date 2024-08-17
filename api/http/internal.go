// Copyright 2024 Northern.tech AS
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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mendersoftware/iot-manager/app"
	"github.com/mendersoftware/iot-manager/model"

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

type internalDevice model.DeviceEvent

func (dev *internalDevice) UnmarshalJSON(b []byte) error {
	type deviceAlias struct {
		// device_id kept for backward compatibility
		ID string `json:"device_id"`
		model.DeviceEvent
	}
	var aDev deviceAlias
	err := json.Unmarshal(b, &aDev)
	if err != nil {
		return err
	}
	if aDev.ID != "" {
		aDev.DeviceEvent.ID = aDev.ID
	}
	*dev = internalDevice(aDev.DeviceEvent)
	return nil
}

// DELETE /tenants/:tenant_id
// code: 204 - all tenant data removed
//
//	500 - internal server error on removal
func (h *InternalHandler) DeleteTenant(c *gin.Context) {
	tenantID := c.Param(ParamTenantID)

	ctx := identity.WithContext(
		c.Request.Context(),
		&identity.Identity{
			Tenant: tenantID,
		},
	)
	err := h.app.DeleteTenant(ctx)
	if err != nil {
		rest.RenderError(c, http.StatusInternalServerError, err)
	}
	c.Status(http.StatusNoContent)
}

// POST /tenants/:tenant_id/devices
// code: 204 - device provisioned to iothub
//
//	500 - internal server error
func (h *InternalHandler) ProvisionDevice(c *gin.Context) {
	tenantID := c.Param(ParamTenantID)
	var device internalDevice
	if err := c.ShouldBindJSON(&device); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"))
		return
	}
	if device.ID == "" {
		rest.RenderError(c, http.StatusBadRequest, errors.New("missing device ID"))
		return
	}

	ctx := identity.WithContext(c.Request.Context(), &identity.Identity{
		Subject: device.ID,
		Tenant:  tenantID,
	})
	err := h.app.ProvisionDevice(ctx, model.DeviceEvent(device))
	switch cause := errors.Cause(err); cause {
	case nil, app.ErrNoCredentials:
		c.Status(http.StatusAccepted)
	case app.ErrDeviceAlreadyExists:
		rest.RenderError(c, http.StatusConflict, cause)
	default:
		rest.RenderError(c, http.StatusInternalServerError, err)
	}
}

func (h *InternalHandler) DecommissionDevice(c *gin.Context) {
	deviceID := c.Param(ParamDeviceID)
	tenantID := c.Param(ParamTenantID)

	ctx := identity.WithContext(c.Request.Context(), &identity.Identity{
		Subject: deviceID,
		Tenant:  tenantID,
	})
	err := h.app.DecommissionDevice(ctx, deviceID)
	switch errors.Cause(err) {
	case nil, app.ErrNoCredentials:
		c.Status(http.StatusAccepted)
	case app.ErrDeviceNotFound:
		rest.RenderError(c, http.StatusNotFound, err)
	default:
		rest.RenderError(c, http.StatusInternalServerError, err)
	}
}

const (
	maxBulkItems = 100
)

// PUT /tenants/:tenant_id/devices/status/{status}
func (h *InternalHandler) BulkSetDeviceStatus(c *gin.Context) { // here
	var schema []struct {
		DeviceID string `json:"id"`
	}
	status := model.Status(c.Param("status"))
	if err := status.Validate(); err != nil {
		rest.RenderError(c, http.StatusBadRequest, err)
		return
	}
	if err := c.ShouldBindJSON(&schema); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "invalid request body"),
		)
		return
	} else if len(schema) > maxBulkItems {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("too many bulk items: max 100 items per request"),
		)
		return
	}
	ctx := identity.WithContext(
		c.Request.Context(),
		&identity.Identity{
			Tenant: c.Param("tenant_id"),
		},
	)
	for _, item := range schema {
		_ = h.app.SetDeviceStatus(ctx, item.DeviceID, status)
	}
	c.Status(http.StatusAccepted)
}

// POST /tenants/:tenant_id/auth
func (h *InternalHandler) PreauthorizeHandler(c *gin.Context) {
	tenantID, okTenant := c.Params.Get("tenant_id")
	if !(okTenant) {
		(*APIHandler)(h).NoRoute(c)
		return
	}
	var req model.PreauthRequest
	if err := c.BindJSON(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	sepIdx := strings.Index(req.DeviceID, " ")
	if sepIdx < 0 {
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("invalid parameter `external_id`"))
		return
	}
	// DeviceID is formatted accordingly: {provider:[iot-hub]}
	provider := req.DeviceID[:sepIdx]
	req.DeviceID = req.DeviceID[sepIdx+1:]

	ctx := identity.WithContext(c.Request.Context(), &identity.Identity{
		IsDevice: true,
		Subject:  req.DeviceID,
		Tenant:   tenantID,
	})
	var err error
	switch provider {
	case string(model.ProviderIoTHub):
		err = h.app.VerifyDeviceTwin(ctx, req)
	default:
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("external provider not supported"))
		return
	}
	if err != nil {
		_ = c.Error(err)
		c.Status(http.StatusUnauthorized)
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /tenants/:tenant_id/bulk/devices/inventory
func (h *InternalHandler) InventoryHandler(c *gin.Context) {
	tenantID, okTenant := c.Params.Get(ParamTenantID)
	if !(okTenant) {
		(*APIHandler)(h).NoRoute(c)
		return
	}

	var req []model.InventoryWebHookData
	if err := c.BindJSON(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	ctx := identity.WithContext(c.Request.Context(), &identity.Identity{
		Tenant: tenantID,
	})
	err := h.app.InventoryChanged(ctx, req)

	if err != nil {
		_ = c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetIntegrationsMap is an internal handler to return information whether a tenant
// has integration of a scope enabled. it returns an array grouped by over all the records
// in the integration collection with tenant id and scope. inventory-enterprise calls this
// endpoint to find a priori tenants with inventory scope integrations -- so we do not have
// to call every time inventory attributes are updated. We really should use Entity tags here.
// GET /integrations?scope=inventory scope is optional
func (h *InternalHandler) GetIntegrationsMap(c *gin.Context) {
	ctx := c.Request.Context()

	scopeParam := c.Query("scope")
	etag := c.Request.Header.Get("If-None-Match")
	if !h.app.IntegrationsChanged(ctx, etag) {
		c.JSON(http.StatusNotModified, []model.IntegrationMap{})
		return
	}

	var scope *string
	if len(scopeParam) > 0 &&
		(scopeParam == model.ScopeDeviceAuth ||
			scopeParam == model.ScopeInventory) {
		scope = &scopeParam
	} else {
		if len(scopeParam) > 0 {
			rest.RenderError(c,
				http.StatusBadRequest,
				errors.New(
					"the scope given is not supported; currently supported: "+
						model.ScopeDeviceAuth+
						" and "+
						model.ScopeInventory,
				),
			)
			return
		}
		scope = nil
	}
	integrations, err := h.app.GetIntegrationsMap(ctx, scope)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			err,
		)
		return
	}
	etag = h.app.GetIntegrationsETag(ctx)
	if len(etag) > 0 {
		c.Header("ETag", etag)
	}

	c.JSON(http.StatusOK, integrations)
}
