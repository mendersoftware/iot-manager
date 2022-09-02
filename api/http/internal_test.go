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

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/rest.utils"
	"github.com/mendersoftware/iot-manager/app"
	mapp "github.com/mendersoftware/iot-manager/app/mocks"
	"github.com/mendersoftware/iot-manager/client"
	"github.com/mendersoftware/iot-manager/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func validateTenantIDCtx(tenantID string) interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		if id := identity.FromContext(ctx); id != nil {
			return id.Tenant == tenantID
		}
		return false
	})
}

func TestProvisionDevice(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name string

		TenantID string
		App      func(*testing.T, *testCase) *mapp.App
		Body     interface{}

		StatusCode int
		Error      error
	}
	testCases := []testCase{{
		Name: "ok/deprecated payload",

		TenantID: "123456789012345678901234",
		Body: map[string]string{
			"device_id": "b8ea97f2-1c2b-492c-84ce-7a90170291b9",
		},
		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			device := self.Body.(map[string]string)
			mock.On("ProvisionDevice",
				validateTenantIDCtx(self.TenantID),
				model.DeviceEvent{ID: device["device_id"]}).
				Return(nil)
			return mock
		},

		StatusCode: http.StatusNoContent,
	}, {
		Name: "ok/noop",

		TenantID: "123456789012345678901234",
		Body: model.DeviceEvent{
			ID: "b8ea97f2-1c2b-492c-84ce-7a90170291b9",
		},
		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			device := self.Body.(model.DeviceEvent)
			mock.On("ProvisionDevice",
				validateTenantIDCtx(self.TenantID),
				device).
				Return(app.ErrNoCredentials)
			return mock
		},

		StatusCode: http.StatusNoContent,
	}, {
		Name: "error/malformed body",

		TenantID: "123456789012345678901234",
		Body:     []byte("is this supposed to be JSON?"),
		App: func(t *testing.T, self *testCase) *mapp.App {
			return new(mapp.App)
		},

		StatusCode: http.StatusBadRequest,
		Error:      errors.New("malformed request body"),
	}, {
		Name: "error/invalid schema",

		TenantID: "123456789012345678901234",
		Body:     []byte(`{"id":true}`),
		App: func(t *testing.T, self *testCase) *mapp.App {
			return new(mapp.App)
		},

		StatusCode: http.StatusBadRequest,
		Error:      errors.New("malformed request body"),
	}, {
		Name: "error/missing device id",

		TenantID: "123456789012345678901234",
		Body:     []byte("{}"),
		App: func(t *testing.T, self *testCase) *mapp.App {
			return new(mapp.App)
		},

		StatusCode: http.StatusBadRequest,
		Error:      errors.New("missing device ID"),
	}, {
		Name: "error/internal failure",

		TenantID: "123456789012345678901234",
		Body: model.DeviceEvent{
			ID: "b8ea97f2-1c2b-492c-84ce-7a90170291b9",
		},
		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			device := self.Body.(model.DeviceEvent)
			mock.On("ProvisionDevice",
				validateTenantIDCtx(self.TenantID),
				device).
				Return(app.ErrDeviceAlreadyExists)
			return mock
		},

		StatusCode: http.StatusConflict,
		Error:      app.ErrDeviceAlreadyExists,
	}, {
		Name: "error/internal failure",

		TenantID: "123456789012345678901234",
		Body: model.DeviceEvent{
			ID: "b8ea97f2-1c2b-492c-84ce-7a90170291b9",
		},
		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			device := self.Body.(model.DeviceEvent)
			mock.On("ProvisionDevice",
				validateTenantIDCtx(self.TenantID),
				device).
				Return(errors.New("internal error"))
			return mock
		},

		StatusCode: http.StatusInternalServerError,
		Error:      errors.New("internal error"),
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			app := tc.App(t, &tc)
			w := httptest.NewRecorder()
			handler := NewRouter(app)

			var body []byte
			switch t := tc.Body.(type) {
			case []byte:
				body = t
			default:
				body, _ = json.Marshal(tc.Body)
			}

			req, _ := http.NewRequest(http.MethodPost,
				"http://localhost"+
					APIURLInternal+
					strings.ReplaceAll(APIURLTenantDevices, ":tenant_id", tc.TenantID),
				bytes.NewReader(body),
			)

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.StatusCode, w.Code)

			if tc.Error != nil {
				var err rest.Error
				json.Unmarshal(w.Body.Bytes(), &err)
				assert.Regexp(t, tc.Error.Error(), err.Error())
			}
		})
	}
}

func TestDecommissionDevice(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name string

		TenantID string
		DeviceID string
		App      func(*testing.T, *testCase) *mapp.App

		StatusCode int
		Error      error
	}
	testCases := []testCase{{
		Name: "ok",

		TenantID: "123456789012345678901234",
		DeviceID: "a8d77d55-ebaa-4ace-b9d4-a2bb581d87f8",

		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			mock.On("DecommissionDevice",
				validateTenantIDCtx(self.TenantID),
				self.DeviceID).
				Return(nil)
			return mock
		},

		StatusCode: http.StatusNoContent,
	}, {
		Name: "ok/noop",

		TenantID: "123456789012345678901234",
		DeviceID: "a8d77d55-ebaa-4ace-b9d4-a2bb581d87f8",

		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			mock.On("DecommissionDevice",
				validateTenantIDCtx(self.TenantID),
				self.DeviceID).
				Return(app.ErrNoCredentials)
			return mock
		},

		StatusCode: http.StatusNoContent,
	}, {
		Name: "error/not found",

		TenantID: "123456789012345678901234",
		DeviceID: "a8d77d55-ebaa-4ace-b9d4-a2bb581d87f8",

		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			mock.On("DecommissionDevice",
				validateTenantIDCtx(self.TenantID),
				self.DeviceID).
				Return(app.ErrDeviceNotFound)
			return mock
		},

		StatusCode: http.StatusNotFound,
		Error:      app.ErrDeviceNotFound,
	}, {
		Name: "error/internal failure",

		TenantID: "123456789012345678901234",
		DeviceID: "a8d77d55-ebaa-4ace-b9d4-a2bb581d87f8",

		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			mock.On("DecommissionDevice",
				validateTenantIDCtx(self.TenantID),
				self.DeviceID).
				Return(errors.New("internal error"))
			return mock
		},

		StatusCode: http.StatusInternalServerError,
		Error:      errors.New("internal error"),
	}}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			app := tc.App(t, &tc)
			w := httptest.NewRecorder()
			handler := NewRouter(app)

			repl := strings.NewReplacer(
				":tenant_id", tc.TenantID,
				":device_id", tc.DeviceID,
			)

			req, _ := http.NewRequest(http.MethodDelete,
				"http://localhost"+
					APIURLInternal+
					repl.Replace(APIURLTenantDevice),
				nil,
			)

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.StatusCode, w.Code)

			if tc.Error != nil {
				var err rest.Error
				json.Unmarshal(w.Body.Bytes(), &err)
				assert.Regexp(t, tc.Error.Error(), err.Error())
			}
		})
	}
}

func TestBulkSetDeviceStatus(t *testing.T) {
	t.Parallel()
	type testCase struct {
		Name string

		TenantID string
		ReqBody  interface{}
		App      func(t *testing.T, self *testCase) *mapp.App
		Status   model.Status

		StatusCode int
		Response   interface{}
	}
	testCases := []testCase{{
		Name: "ok",

		TenantID: "123456789012345678901234",
		ReqBody: []map[string]interface{}{{
			"id": "960700f7-d563-4a31-94e6-a075fe6566bc",
		}, {
			"id": "3fd916c1-6a5a-423c-b7da-739bf21c7779",
		}, {
			"id": "1cb050b9-c20c-4807-bdbd-bc5650617198",
		}},
		Status: model.StatusAccepted,

		App: func(t *testing.T, self *testCase) *mapp.App {
			var result BulkResult
			mockApp := new(mapp.App)
			req := self.ReqBody.([]map[string]interface{})
			for _, id := range req {
				mockApp.On("SetDeviceStatus",
					contextMatcher,
					id["id"],
					model.StatusAccepted,
				).Return(nil)
				result.Items = append(result.Items, BulkItem{
					Status: http.StatusOK,
					Parameters: map[string]interface{}{
						"id": id["id"],
					},
				})
			}
			self.Response = result
			return mockApp
		},
		StatusCode: http.StatusOK,
	}, {
		Name: "ok, no result",

		TenantID: "123456789012345678901234",
		ReqBody:  []struct{}{},
		Status:   model.StatusPending,

		App: func(t *testing.T, self *testCase) *mapp.App {
			mockApp := new(mapp.App)
			return mockApp
		},
		Response:   BulkResult{Items: []BulkItem{}},
		StatusCode: http.StatusOK,
	}, {
		Name: "error, partial result",

		TenantID: "123456789012345678901234",
		ReqBody: []map[string]interface{}{{
			"id": "960700f7-d563-4a31-94e6-a075fe6566bc",
		}, {
			"id": "3fd916c1-6a5a-423c-b7da-739bf21c7779",
		}, {
			"id": "1cb050b9-c20c-4807-bdbd-bc5650617198",
		}},
		Status: model.StatusPreauthorized,

		App: func(t *testing.T, self *testCase) *mapp.App {
			var result BulkResult
			mockApp := new(mapp.App)
			req := self.ReqBody.([]map[string]interface{})
			mockApp.On("SetDeviceStatus",
				contextMatcher,
				req[0]["id"],
				self.Status,
			).Return(nil).Once()
			result.Items = append(result.Items, BulkItem{
				Status: http.StatusOK,
				Parameters: map[string]interface{}{
					"id": req[0]["id"],
				},
			})
			mockApp.On("SetDeviceStatus",
				contextMatcher,
				req[1]["id"],
				self.Status,
			).Return(errors.New("internal error")).Once()
			result.Items = append(result.Items, BulkItem{
				Status:      http.StatusInternalServerError,
				Description: "internal error",
				Parameters: map[string]interface{}{
					"id": req[1]["id"],
				},
			})
			mockApp.On("SetDeviceStatus",
				contextMatcher,
				req[2]["id"],
				self.Status,
			).Return(client.NewHTTPError(http.StatusConflict)).Once()
			result.Items = append(result.Items, BulkItem{
				Status:      http.StatusConflict,
				Description: client.NewHTTPError(http.StatusConflict).Error(),
				Parameters: map[string]interface{}{
					"id": req[2]["id"],
				},
			})
			result.Error = true
			self.Response = result
			return mockApp
		},
		StatusCode: http.StatusOK,
	}, {
		Name: "error: invalid request body",

		TenantID: "123456789012345678901234",
		ReqBody:  []byte("rawr"),
		Status:   model.StatusRejected,
		App: func(t *testing.T, self *testCase) *mapp.App {
			mockApp := new(mapp.App)
			return mockApp
		},
		StatusCode: http.StatusBadRequest,
		Response:   regexp.MustCompile(`{"error":\s?"invalid request body.*",\s?"request_id":\s?"test"}`),
	}, {
		Name: "error: invalid status parameter",

		TenantID: "123456789012345678901234",
		ReqBody:  []struct{}{},
		Status:   model.Status("foobar"),
		App: func(t *testing.T, self *testCase) *mapp.App {
			mockApp := new(mapp.App)
			return mockApp
		},
		StatusCode: http.StatusBadRequest,
		Response:   regexp.MustCompile(`{"error":\s?"invalid status 'foobar'",\s?"request_id":\s?"test"}`),
	}, {
		Name: "error: too many items",

		TenantID: "123456789012345678901234",
		ReqBody:  make([]struct{}, maxBulkItems+1),
		Status:   model.StatusAccepted,
		App: func(t *testing.T, self *testCase) *mapp.App {
			mockApp := new(mapp.App)
			return mockApp
		},
		StatusCode: http.StatusBadRequest,
		Response: regexp.MustCompile(fmt.Sprintf(
			`{"error":\s?"too many bulk items: max %d items per request",`+
				`\s?"request_id":\s?"test"}`,
			maxBulkItems,
		)),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			app := tc.App(t, &tc)
			defer app.AssertExpectations(t)
			w := httptest.NewRecorder()
			handler := NewRouter(app)
			repl := strings.NewReplacer(
				":tenant_id", tc.TenantID,
				":status", string(tc.Status),
			)
			var b []byte
			switch t := tc.ReqBody.(type) {
			case []byte:
				b = t
			default:
				b, _ = json.Marshal(tc.ReqBody)
			}
			req, _ := http.NewRequest(
				http.MethodPut,
				"http://localhost"+
					APIURLInternal+
					repl.Replace(APIURLTenantBulkStatus),
				bytes.NewReader(b),
			)
			req.Header.Set("X-Men-Requestid", "test")

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.StatusCode, w.Code)
			switch res := tc.Response.(type) {
			case []byte:
				assert.Contains(t, w.Body.Bytes(), res)
			case nil:
				assert.Empty(t, w.Body.Bytes())
			case *regexp.Regexp:
				assert.Regexp(t, res, w.Body.String())
			default:
				b, _ := json.Marshal(res)
				assert.JSONEq(t, string(b), w.Body.String())
			}
		})
	}
}
