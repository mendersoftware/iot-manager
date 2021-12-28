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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/rest.utils"

	"github.com/mendersoftware/iot-manager/app"
	"github.com/mendersoftware/iot-manager/model"
)

var (
	ErrMissingUserAuthentication = errors.New(
		"user identity missing from authorization token",
	)
	ErrIntegrationNotFound = errors.New("integration not found")
)

// ManagementHandler is the namespace for management API handlers.
type ManagementHandler APIHandler

// GET /integrations
func (h *ManagementHandler) GetIntegrations(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		id  = identity.FromContext(ctx)
	)

	if id == nil || !id.IsUser {
		rest.RenderError(c, http.StatusForbidden, ErrMissingUserAuthentication)
		return
	}

	integrations, err := h.app.GetIntegrations(ctx)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			errors.New(http.StatusText(http.StatusInternalServerError)),
		)
		return
	}

	c.JSON(http.StatusOK, integrations)
}

// POST /integrations
func (h *ManagementHandler) CreateIntegration(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		id  = identity.FromContext(ctx)
	)

	if id == nil || !id.IsUser {
		rest.RenderError(c, http.StatusForbidden, ErrMissingUserAuthentication)
		return
	}

	integration := model.Integration{}
	if err := c.ShouldBindJSON(&integration); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	// TODO verify that Azure connectionstring / AWS equivalent has correct permissions
	//      - service
	//      - registry read/write

	err := h.app.CreateIntegration(ctx, integration)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case app.ErrIntegrationExists:
			// NOTE: temporary limitation
			rest.RenderError(c, http.StatusConflict, cause)
		default:
			_ = c.Error(err)
			rest.RenderError(c,
				http.StatusInternalServerError,
				errors.New(http.StatusText(http.StatusInternalServerError)),
			)
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// GET /integrations/{id}
func (h *ManagementHandler) GetIntegration(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		id  = identity.FromContext(ctx)
	)

	if id == nil || !id.IsUser {
		rest.RenderError(c, http.StatusForbidden, ErrMissingUserAuthentication)
		return
	}
	integrationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		rest.RenderError(c, http.StatusNotFound, ErrIntegrationNotFound)
		return
	}

	integration, err := h.app.GetIntegrationById(ctx, integrationID)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case app.ErrIntegrationNotFound:
			rest.RenderError(c, http.StatusNotFound, ErrIntegrationNotFound)
		default:
			rest.RenderError(c,
				http.StatusInternalServerError,
				errors.New(http.StatusText(http.StatusInternalServerError)),
			)
		}
		return
	}

	c.JSON(http.StatusOK, integration)
}
