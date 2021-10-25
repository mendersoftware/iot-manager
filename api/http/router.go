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
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"

	"github.com/mendersoftware/go-lib-micro/accesslog"
	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/log"
	"github.com/mendersoftware/go-lib-micro/requestid"
	"github.com/mendersoftware/go-lib-micro/rest.utils"

	"github.com/mendersoftware/azure-iot-manager/app"
)

// API URL used by the HTTP router
const (
	APIURLInternal = "/api/internal/v1/azure-iot-manager"

	APIURLAlive  = "/alive"
	APIURLHealth = "/health"

	APIURLManagement = "/api/management/v1/azure-iot-manager"

	APIURLSettings = "/settings"
)

const (
	defaultTimeout = time.Second * 10
)

// NewRouter returns the gin router
func NewRouter(app app.App) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	handler := NewAPIHandler(app)

	router := gin.New()
	router.Use(accesslog.Middleware())
	router.Use(requestid.Middleware())

	router.NoMethod(handler.NoMethod)
	router.NoRoute(handler.NoRoute)

	internalAPI := router.Group(APIURLInternal)
	internalAPI.GET(APIURLAlive, handler.Alive)
	internalAPI.GET(APIURLHealth, handler.Health)

	management := NewManagementHandler(handler)
	managementAPI := router.Group(APIURLManagement, identity.Middleware())
	managementAPI.GET(APIURLSettings, management.GetSettings)
	managementAPI.PUT(APIURLSettings, management.SetSettings)

	return router, nil
}

type APIHandler struct {
	app app.App
}

func NewAPIHandler(app app.App) *APIHandler {
	return &APIHandler{
		app: app,
	}
}

// Alive responds to GET /alive
func (h *APIHandler) Alive(c *gin.Context) {
	c.Writer.WriteHeader(http.StatusNoContent)
}

// Health responds to GET /health
func (h *APIHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()
	l := log.FromContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	err := h.app.HealthCheck(ctx)
	if err != nil {
		l.Error(errors.Wrap(err, "health check failed"))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) NoMethod(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, rest.Error{
		Err:       "method not allowed",
		RequestID: requestid.FromContext(c.Request.Context()),
	})
}

func (h *APIHandler) NoRoute(c *gin.Context) {
	c.JSON(http.StatusNotFound, rest.Error{
		Err:       "not found",
		RequestID: requestid.FromContext(c.Request.Context()),
	})
}

// Make gin-gonic use validatable structs instead of relying on go-playground
// validator interface.
type validateValidatableValidator struct{}

func (validateValidatableValidator) ValidateStruct(obj interface{}) error {
	if v, ok := obj.(interface{ Validate() error }); ok {
		return v.Validate()
	}
	return nil
}

func (validateValidatableValidator) Engine() interface{} {
	return nil
}

func init() {
	binding.Validator = validateValidatableValidator{}
}
