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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/requestid"
	"github.com/mendersoftware/go-lib-micro/rest.utils"

	"github.com/mendersoftware/iot-manager/app"
	mapp "github.com/mendersoftware/iot-manager/app/mocks"
	"github.com/mendersoftware/iot-manager/model"
)

var (
	contextMatcher  = mock.MatchedBy(func(_ context.Context) bool { return true })
	validConnString = &model.ConnectionString{
		HostName: "localhost:8080",
		Key:      []byte("not-so-secret-key"),
		Name:     "foobar",
	}
)

func compareParameterValues(t *testing.T, expected interface{}) interface{} {
	return mock.MatchedBy(func(actual interface{}) bool {
		return assert.EqualValues(t, expected, actual)
	})
}

func GenerateJWT(id identity.Identity) string {
	JWT := base64.RawURLEncoding.EncodeToString(
		[]byte(`{"alg":"HS256","typ":"JWT"}`),
	)
	b, _ := json.Marshal(id)
	JWT = JWT + "." + base64.RawURLEncoding.EncodeToString(b)
	hash := hmac.New(sha256.New, []byte("hmac-sha256-secret"))
	JWT = JWT + "." + base64.RawURLEncoding.EncodeToString(
		hash.Sum([]byte(JWT)),
	)
	return JWT
}

func TestGetIntegrations(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		Headers http.Header

		App func(t *testing.T) *mapp.App

		StatusCode int
		Response   interface{}
	}{
		{
			Name: "ok",

			Headers: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},

			App: func(t *testing.T) *mapp.App {
				app := new(mapp.App)
				app.On("GetIntegrations", contextMatcher).
					Return([]model.Integration{
						{
							ID:       uuid.Nil,
							Provider: model.ProviderIoTHub,
							Credentials: model.Credentials{
								Type:             model.CredentialTypeSAS,
								ConnectionString: validConnString,
							},
						},
					}, nil)
				return app
			},

			StatusCode: http.StatusOK,
			Response: []map[string]interface{}{{
				"id":       uuid.Nil,
				"provider": model.ProviderIoTHub,
				"credentials": map[string]interface{}{
					"type":              model.CredentialTypeSAS,
					"connection_string": validConnString.String(),
				},
			}},
		},
		{
			Name: "ok empty settings",

			Headers: http.Header{
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsUser:  true,
					Subject: "829cbefb-70e7-438f-9ac5-35fd131c2111",
					Tenant:  "123456789012345678901234",
				})},
			},

			App: func(t *testing.T) *mapp.App {
				app := new(mapp.App)
				app.On("GetIntegrations", contextMatcher).Return([]model.Integration{}, nil)
				return app
			},

			StatusCode: http.StatusOK,
			Response:   []model.Integration{},
		},
		{
			Name: "error, invalid authorization header",

			Headers: http.Header{
				textproto.CanonicalMIMEHeaderKey(requestid.RequestIdHeader): []string{
					"829cbefb-70e7-438f-9ac5-35fd131c2111",
				},
				"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
					IsDevice: true,
					Subject:  "829cbefb-70e7-438f-9ac5-35fd131c2f76",
					Tenant:   "123456789012345678901234",
				})},
			},
			StatusCode: http.StatusForbidden,
			Response: rest.Error{
				Err:       ErrMissingUserAuthentication.Error(),
				RequestID: "829cbefb-70e7-438f-9ac5-35fd131c2111",
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			var testApp *mapp.App
			if tc.App == nil {
				testApp = new(mapp.App)
			} else {
				testApp = tc.App(t)
			}
			defer testApp.AssertExpectations(t)
			handler := NewRouter(testApp)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET",
				"http://localhost"+
					APIURLManagement+
					APIURLIntegrations,
				nil,
			)
			for key := range tc.Headers {
				req.Header.Set(key, tc.Headers.Get(key))
			}

			handler.ServeHTTP(w, req)
			assert.Equal(t, tc.StatusCode, w.Code, "invalid HTTP status code")
			b, _ := json.Marshal(tc.Response)
			assert.JSONEq(t, string(b), w.Body.String())
		})
	}
}

func TestCreateIntegration(t *testing.T) {
	t.Parallel()
	var jitter string
	for i := 0; i < 4096; i++ {
		jitter += "1"
	}
	testCases := []struct {
		Name string

		RequestBody interface{}
		RequestHdrs http.Header

		App func(t *testing.T) *mapp.App

		RspCode int
		Error   error
	}{{
		Name: "ok",

		RequestBody: map[string]interface{}{
			"provider": model.ProviderIoTHub,
			"credentials": map[string]interface{}{
				"type":              model.CredentialTypeSAS,
				"connection_string": validConnString.String(),
			},
		},
		RequestHdrs: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewString(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			})},
		},

		App: func(t *testing.T) *mapp.App {
			a := new(mapp.App)
			a.On("CreateIntegration",
				contextMatcher,
				mock.AnythingOfType("model.Integration")).
				Return(nil)
			return a
		},

		RspCode: http.StatusNoContent,
	}, {
		Name: "duplicate integration",

		RequestBody: map[string]interface{}{
			"provider": model.ProviderIoTHub,
			"credentials": model.Credentials{
				Type:             model.CredentialTypeSAS,
				ConnectionString: validConnString,
			},
		},
		RequestHdrs: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewString(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			})},
		},

		App: func(t *testing.T) *mapp.App {
			a := new(mapp.App)
			a.On("CreateIntegration", contextMatcher, mock.AnythingOfType("model.Integration")).
				Return(app.ErrIntegrationExists)
			return a
		},

		RspCode: http.StatusConflict,
		Error:   app.ErrIntegrationExists,
	}, {
		Name: "internal error",

		RequestBody: map[string]interface{}{
			"provider": model.ProviderIoTHub,
			"credentials": model.Credentials{
				Type:             model.CredentialTypeSAS,
				ConnectionString: validConnString,
			},
		},
		RequestHdrs: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewString(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			})},
		},

		App: func(t *testing.T) *mapp.App {
			a := new(mapp.App)
			a.On("CreateIntegration", contextMatcher, mock.AnythingOfType("model.Integration")).
				Return(errors.New("internal error"))
			return a
		},

		RspCode: http.StatusInternalServerError,
		Error:   errors.New(http.StatusText(http.StatusInternalServerError)),
	}, {
		Name: "malformed request body",

		RequestBody: map[string]interface{}{
			"provider":    model.ProviderIoTHub,
			"credentials": 1234,
		},
		RequestHdrs: http.Header{
			"Authorization": []string{"Bearer " + GenerateJWT(identity.Identity{
				Subject: uuid.NewString(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			})},
		},

		App: func(t *testing.T) *mapp.App { return new(mapp.App) },

		RspCode: http.StatusBadRequest,
		Error:   errors.New("malformed request body: json:"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			app := tc.App(t)
			defer app.AssertExpectations(t)
			var body io.Reader
			if tc.RequestBody != nil {
				b, _ := json.Marshal(tc.RequestBody)
				body = bytes.NewReader(b)
			}
			req, _ := http.NewRequest(http.MethodPost,
				"http://localhost"+APIURLManagement+APIURLIntegrations,
				body,
			)
			for k, v := range tc.RequestHdrs {
				req.Header[k] = v
			}

			router := NewRouter(app)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tc.RspCode, w.Code)
			if tc.Error != nil {
				var erro rest.Error
				if assert.NotNil(t, w.Body) {
					err := json.Unmarshal(w.Body.Bytes(), &erro)
					require.NoError(t, err)
					assert.Regexp(t, tc.Error.Error(), erro.Error())
				}
			} else {
				assert.Empty(t, w.Body.Bytes(), string(w.Body.Bytes()))
			}
		})
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func validateAuthz(t *testing.T, key []byte, hdr string) bool {
	if !assert.True(t, strings.HasPrefix(hdr, "SharedAccessSignature")) {
		return false
	}
	hdr = strings.TrimPrefix(hdr, "SharedAccessSignature")
	hdr = strings.TrimLeft(hdr, " ")
	q, err := url.ParseQuery(hdr)
	if !assert.NoError(t, err) {
		return false
	}
	for _, key := range []string{"sr", "se", "sig"} {
		if !assert.Contains(t, q, key, "missing signature parameters") {
			return false
		}
	}
	msg := fmt.Sprintf("%s\n%s", url.QueryEscape(q.Get("sr")), q.Get("se"))
	digest := hmac.New(sha256.New, key)
	digest.Write([]byte(msg))
	expected := digest.Sum(nil)
	return assert.Equal(t, base64.StdEncoding.EncodeToString(expected), q.Get("sig"))
}

type neverExpireContext struct {
	context.Context
}

func (neverExpireContext) Deadline() (time.Time, bool) {
	return time.Now().Add(time.Hour), true
}
