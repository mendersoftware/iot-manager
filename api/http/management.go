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
	"github.com/pkg/errors"

	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/rest.utils"

	"github.com/mendersoftware/azure-iot-manager/app"
	"github.com/mendersoftware/azure-iot-manager/model"
)

var (
	ErrMissingUserAuthentication = errors.New(
		"user identity missing from authorization token",
	)
)

// ManagementController container for end-points
type ManagementController struct {
	app app.App
}

// NewManagementController returns a new ManagementController
func NewManagementController(app app.App) *ManagementController {
	return &ManagementController{
		app: app,
	}
}

// GET /settings
func (h *ManagementController) GetSettings(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		id  = identity.FromContext(ctx)
	)

	if id == nil || !id.IsUser {
		rest.RenderError(c, http.StatusForbidden, ErrMissingUserAuthentication)
		return
	}
	settings, err := h.app.GetSettings(ctx)
	if err != nil {
		rest.RenderError(c,
			http.StatusInternalServerError,
			errors.New(http.StatusText(http.StatusInternalServerError)),
		)
		return
	}

	c.JSON(http.StatusOK, settings)
}

// PUT /settings
func (h *ManagementController) SetSettings(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		id  = identity.FromContext(ctx)
	)

	if id == nil || !id.IsUser {
		rest.RenderError(c, http.StatusForbidden, ErrMissingUserAuthentication)
		return
	}

	settings := model.Settings{}
	if err := c.ShouldBindJSON(&settings); err != nil {
		rest.RenderError(c,
			http.StatusBadRequest,
			errors.New("malformed request body"),
		)
		return
	}

	err := h.app.SetSettings(ctx, settings)
	if err != nil {
		_ = c.Error(err)
		rest.RenderError(c,
			http.StatusInternalServerError,
			errors.New(http.StatusText(http.StatusInternalServerError)),
		)
		return
	}
	c.Status(http.StatusNoContent)
}
