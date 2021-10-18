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
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/mendersoftware/go-lib-micro/accesslog"
	"github.com/mendersoftware/go-lib-micro/requestid"

	"github.com/mendersoftware/azure-iot-manager/app"
)

// API URL used by the HTTP router
const (
	APIURLInternal = "/api/internal/v1/azure-iot-manager"

	APIURLAlive  = "/alive"
	APIURLHealth = "/health"
)

// NewRouter returns the gin router
func NewRouter(app app.App) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	router := gin.New()
	router.Use(accesslog.Middleware())
	router.Use(requestid.Middleware())

	status := NewStatusController(app)
	internalAPI := router.Group(APIURLInternal)
	internalAPI.GET(APIURLAlive, status.Alive)
	internalAPI.GET(APIURLHealth, status.Health)

	return router, nil
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
