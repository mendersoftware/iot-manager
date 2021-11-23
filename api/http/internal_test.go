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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mendersoftware/azure-iot-manager/app"
	mapp "github.com/mendersoftware/azure-iot-manager/app/mocks"
	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/rest.utils"
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
	type intrnlDevice struct {
		ID string `json:"device_id"`
	}
	type testCase struct {
		Name string

		TenantID string
		App      func(*testing.T, *testCase) *mapp.App
		Body     interface{}

		StatusCode int
		Error      error
	}
	testCases := []testCase{{
		Name: "ok",

		TenantID: "123456789012345678901234",
		Body: intrnlDevice{
			ID: "b8ea97f2-1c2b-492c-84ce-7a90170291b9",
		},
		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			device := self.Body.(intrnlDevice)
			mock.On("ProvisionDevice",
				validateTenantIDCtx(self.TenantID),
				device.ID).
				Return(nil)
			return mock
		},

		StatusCode: http.StatusAccepted,
	}, {
		Name: "ok/noop",

		TenantID: "123456789012345678901234",
		Body: intrnlDevice{
			ID: "b8ea97f2-1c2b-492c-84ce-7a90170291b9",
		},
		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			device := self.Body.(intrnlDevice)
			mock.On("ProvisionDevice",
				validateTenantIDCtx(self.TenantID),
				device.ID).
				Return(app.ErrNoConnectionString)
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
		Body: intrnlDevice{
			ID: "b8ea97f2-1c2b-492c-84ce-7a90170291b9",
		},
		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			device := self.Body.(intrnlDevice)
			mock.On("ProvisionDevice",
				validateTenantIDCtx(self.TenantID),
				device.ID).
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
			mock.On("DeleteIOTHubDevice",
				validateTenantIDCtx(self.TenantID),
				self.DeviceID).
				Return(nil)
			return mock
		},

		StatusCode: http.StatusAccepted,
	}, {
		Name: "ok/noop",

		TenantID: "123456789012345678901234",
		DeviceID: "a8d77d55-ebaa-4ace-b9d4-a2bb581d87f8",

		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			mock.On("DeleteIOTHubDevice",
				validateTenantIDCtx(self.TenantID),
				self.DeviceID).
				Return(app.ErrNoConnectionString)
			return mock
		},

		StatusCode: http.StatusNoContent,
	}, {
		Name: "error/internal failure",

		TenantID: "123456789012345678901234",
		DeviceID: "a8d77d55-ebaa-4ace-b9d4-a2bb581d87f8",

		App: func(t *testing.T, self *testCase) *mapp.App {
			mock := new(mapp.App)
			mock.On("DeleteIOTHubDevice",
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
