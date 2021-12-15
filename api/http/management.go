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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/rest.utils"

	"github.com/mendersoftware/azure-iot-manager/model"
)

const (
	// https://docs.microsoft.com/en-us/rest/api/iothub/service/devices
	AzureAPIVersion = "2021-04-12"

	AzureURIDeviceTwin    AzureDeviceURI = "/twins/:id"
	AzureURIDevice        AzureDeviceURI = "/devices/:id"
	AzureURIDeviceModules AzureDeviceURI = "/devices/:id/modules"

	defaultTTL = time.Minute
)

type AzureDeviceURI string

func (s AzureDeviceURI) URI(deviceID string) string {
	return strings.Replace(string(s), ":id", deviceID, 1)
}

const (
	HdrKeyAuthz       = "Authorization"
	HdrKeyXFF         = "X-Forwarded-For"
	HdrKeyMSRequestID = "X-Ms-Request-Id"
)

// Hop-by-hop headers (RFC2616 section 13.5.1)
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hbhHeaders = [...]string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
	"X-Men-Requestid",
}

func delHbHHeaders(header http.Header) {
	for _, hdr := range hbhHeaders {
		delete(header, hdr)
	}
}

var (
	ErrMissingUserAuthentication = errors.New(
		"user identity missing from authorization token",
	)
	ErrMissingConnectionString = errors.New("connection string is not configured")
)

// ManagementHandler is the namespace for management API handlers.
type ManagementHandler APIHandler

func (h *ManagementHandler) proxyAzureRequest(c *gin.Context, dstPath string) {
	req := c.Request
	ctx := req.Context()
	settings, err := h.app.GetSettings(ctx)
	switch {
	case err != nil:
		_ = c.Error(err)
		rest.RenderError(c,
			http.StatusInternalServerError,
			errors.New(http.StatusText(http.StatusInternalServerError)),
		)
		return
	case settings.ConnectionString == nil:
		fallthrough
	case settings.ConnectionString.Validate() != nil:
		rest.RenderError(c, http.StatusConflict, ErrMissingConnectionString)
		return
	default:
	}
	cs := settings.ConnectionString
	q := c.Request.URL.Query()
	q.Set("api-version", AzureAPIVersion)
	req.URL.Scheme = "https"
	req.URL.RawQuery = q.Encode()
	req.Host = cs.HostName
	req.URL.Path = dstPath

	req.URL.Host = cs.HostName
	if cs.GatewayHostName != "" {
		req.URL.Host = cs.GatewayHostName
	}
	req.RequestURI = ""

	delHbHHeaders(req.Header)

	var expireAt time.Time
	if dl, ok := ctx.Deadline(); ok {
		expireAt = dl
	} else {
		var cancel context.CancelFunc
		expireAt = time.Now().Add(defaultTTL)
		ctx, cancel = context.WithDeadline(ctx, expireAt)
		defer cancel()
		req = req.WithContext(ctx)
	}
	req.Header.Set(HdrKeyAuthz, cs.Authorization(expireAt))
	if req.Header.Get(HdrKeyXFF) == "" {
		if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			req.Header.Set(HdrKeyXFF, host)
		}
	}
	rsp, err := h.Do(req)
	if err != nil {
		_ = c.Error(err)
		rest.RenderError(c,
			http.StatusBadGateway,
			errors.New("failed to proxy request to IoT Hub"),
		)
		return
	}
	defer rsp.Body.Close()
	delHbHHeaders(rsp.Header)
	delete(rsp.Header, HdrKeyMSRequestID)
	rspHdrs := c.Writer.Header()
	for k, v := range rsp.Header {
		rspHdrs[k] = v
	}
	c.Status(rsp.StatusCode)
	_, err = io.Copy(c.Writer, rsp.Body)
	if err != nil {
		_ = c.Error(err)
	}
}

func (h *ManagementHandler) GetDeviceModules(c *gin.Context) {
	h.proxyAzureRequest(c, AzureURIDeviceModules.URI(c.Param("id")))
}

func (h *ManagementHandler) GetDevice(c *gin.Context) {
	h.proxyAzureRequest(c, AzureURIDevice.URI(c.Param("id")))
}

// GET /device/:id/twin
func (h *ManagementHandler) GetDeviceTwin(c *gin.Context) {
	h.proxyAzureRequest(c, AzureURIDeviceTwin.URI(c.Param("id")))
}

// PATCH /device/:id/twin
func (h *ManagementHandler) UpdateDeviceTwin(c *gin.Context) {
	var schema struct {
		Properties map[string]interface{} `json:"properties"`
		Tags       map[string]interface{} `json:"tags,omitempty"`
	}
	var azureSchema struct {
		Properties struct {
			Desired map[string]interface{} `json:"desired"`
		} `json:"properties"`
		Tags map[string]interface{} `json:"tags,omitempty"`
	}
	err := c.ShouldBindJSON(&schema)
	c.Request.Body.Close()
	if err != nil {
		rest.RenderError(c, http.StatusBadRequest, errors.Wrap(err,
			"malformed request body",
		))
		return
	}
	azureSchema.Properties.Desired = schema.Properties
	azureSchema.Tags = schema.Tags
	b, _ := json.Marshal(azureSchema)
	c.Request.Body = io.NopCloser(bytes.NewReader(b))
	c.Request.ContentLength = int64(len(b))
	h.proxyAzureRequest(c, AzureURIDeviceTwin.URI(c.Param("id")))
}

// GET /settings
func (h *ManagementHandler) GetSettings(c *gin.Context) {
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
func (h *ManagementHandler) SetSettings(c *gin.Context) {
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
			errors.Wrap(err, "malformed request body"),
		)
		return
	}

	// TODO verify that connectionstring has correct permissions
	//      - service
	//      - registry read/write

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
