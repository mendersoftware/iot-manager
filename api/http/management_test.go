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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/log"
	"github.com/mendersoftware/go-lib-micro/requestid"
	"github.com/mendersoftware/go-lib-micro/rest.utils"

	mapp "github.com/mendersoftware/azure-iot-manager/app/mocks"
	"github.com/mendersoftware/azure-iot-manager/model"
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

func TestGetSettings(t *testing.T) {
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
				app.On("GetSettings",
					contextMatcher).
					Return(model.Settings{
						ConnectionString: validConnString,
					}, nil)
				return app
			},

			StatusCode: http.StatusOK,
			Response: map[string]interface{}{
				"connection_string": validConnString.String(),
			},
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
				app.On("GetSettings",
					contextMatcher).
					Return(model.Settings{}, nil)
				return app
			},

			StatusCode: http.StatusOK,
			Response:   model.Settings{},
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
					APIURLSettings,
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

func TestSetSettings(t *testing.T) {
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

		RequestBody: map[string]string{
			"connection_string": validConnString.String(),
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
			a.On("SetSettings", contextMatcher, mock.AnythingOfType("model.Settings")).
				Return(nil)
			return a
		},

		RspCode: http.StatusNoContent,
	}, {
		Name: "internal error",

		RequestBody: map[string]string{
			"connection_string": validConnString.String(),
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
			a.On("SetSettings", contextMatcher, mock.AnythingOfType("model.Settings")).
				Return(errors.New("internal error"))
			return a
		},

		RspCode: http.StatusInternalServerError,
		Error:   errors.New(http.StatusText(http.StatusInternalServerError)),
	}, {
		Name: "settings string too long",

		RequestBody: map[string]string{
			"connection_string": validConnString.String() + ";DeviceId=" + jitter,
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
		Error: errors.Wrap(model.ErrConnectionStringTooLong,
			"malformed request body: connection string invalid",
		),
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
			req, _ := http.NewRequest("PUT",
				"http://localhost"+APIURLManagement+APIURLSettings,
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

func TestProxyAzureRequest(t *testing.T) {
	t.Parallel()
	logger := log.NewEmpty()
	logger.Logger.Out = io.Discard
	ctxWithoutLog := log.WithContext(context.Background(), logger)
	type testCase struct {
		Name string

		ConnString *model.ConnectionString
		App        func(t *testing.T, self *testCase) *mapp.App

		ClientError error

		Req *http.Request

		// Response
		Code   int
		Header http.Header
		Body   interface{}
	}
	testCases := []testCase{{
		Name: "ok, GET twin",

		App: func(t *testing.T, self *testCase) *mapp.App {
			app := new(mapp.App)
			app.On("GetSettings", contextMatcher).
				Return(model.Settings{
					ConnectionString: self.ConnString,
				}, nil)
			return app
		},
		ConnString: &model.ConnectionString{
			HostName:        "acme.iot.hub",
			Key:             []byte("not-so-secret-key"),
			Name:            "foobar",
			GatewayHostName: "localhost:8080",
		},
		Req: func() *http.Request {
			r, _ := http.NewRequestWithContext(
				neverExpireContext{Context: ctxWithoutLog},
				http.MethodGet,
				"http://localhost"+APIURLManagement+strings.Replace(
					APIURLDeviceTwin,
					":id",
					uuid.New().String(),
					1),
				nil,
			)
			r.Header.Set("Authorization", "Bearer "+GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.Nil, []byte("Hans")).String(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			}))
			r.Header.Set("X-Test", "test")
			return r
		}(),
	}, {
		Name: "ok, PUT twin",

		App: func(t *testing.T, self *testCase) *mapp.App {
			app := new(mapp.App)
			app.On("GetSettings", contextMatcher).
				Return(model.Settings{
					ConnectionString: self.ConnString,
				}, nil)
			return app
		},
		ConnString: validConnString,
		Req: func() *http.Request {
			b, _ := json.Marshal(map[string]interface{}{
				"properties": map[string]string{
					"testing": "test",
				},
			})
			r, _ := http.NewRequestWithContext(
				ctxWithoutLog,
				http.MethodPut,
				"http://localhost"+APIURLManagement+strings.Replace(
					APIURLDeviceTwin,
					":id",
					uuid.New().String(),
					1),
				bytes.NewReader(b),
			)
			r.Header.Set("Authorization", "Bearer "+GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.Nil, []byte("Hans")).String(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			}))
			r.RemoteAddr = "test.subject.io:8080"
			return r
		}(),
	}, {
		Name: "ok, PATCH twin",

		App: func(t *testing.T, self *testCase) *mapp.App {
			app := new(mapp.App)
			app.On("GetSettings", contextMatcher).
				Return(model.Settings{
					ConnectionString: self.ConnString,
				}, nil)
			return app
		},
		ConnString: validConnString,
		Req: func() *http.Request {
			b, _ := json.Marshal(map[string]interface{}{
				"properties": map[string]string{
					"testing": "test",
				},
			})
			r, _ := http.NewRequestWithContext(
				ctxWithoutLog,
				http.MethodPatch,
				"http://localhost"+APIURLManagement+strings.Replace(
					APIURLDeviceTwin,
					":id",
					uuid.New().String(),
					1),
				bytes.NewReader(b),
			)
			r.Header.Set("Authorization", "Bearer "+GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.Nil, []byte("Hans")).String(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			}))
			return r
		}(),
	}, {
		Name: "ok, GET device",

		App: func(t *testing.T, self *testCase) *mapp.App {
			app := new(mapp.App)
			app.On("GetSettings", contextMatcher).
				Return(model.Settings{
					ConnectionString: self.ConnString,
				}, nil)
			return app
		},
		ConnString: &model.ConnectionString{
			HostName:        "acme.iot.hub",
			Key:             []byte("not-so-secret-key"),
			Name:            "foobar",
			GatewayHostName: "localhost:8080",
		},
		Req: func() *http.Request {
			r, _ := http.NewRequestWithContext(
				neverExpireContext{Context: ctxWithoutLog},
				http.MethodGet,
				"http://localhost"+APIURLManagement+strings.Replace(
					APIURLDevice,
					":id",
					uuid.New().String(),
					1),
				nil,
			)
			r.Header.Set("Authorization", "Bearer "+GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.Nil, []byte("Hans")).String(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			}))
			r.Header.Set("X-Test", "test")
			return r
		}(),
	}, {
		Name: "ok, GET device modules",

		App: func(t *testing.T, self *testCase) *mapp.App {
			app := new(mapp.App)
			app.On("GetSettings", contextMatcher).
				Return(model.Settings{
					ConnectionString: self.ConnString,
				}, nil)
			return app
		},
		ConnString: &model.ConnectionString{
			HostName:        "acme.iot.hub",
			Key:             []byte("not-so-secret-key"),
			Name:            "foobar",
			GatewayHostName: "localhost:8080",
		},
		Req: func() *http.Request {
			r, _ := http.NewRequestWithContext(
				neverExpireContext{Context: ctxWithoutLog},
				http.MethodGet,
				"http://localhost"+APIURLManagement+strings.Replace(
					APIURLDeviceModules,
					":id",
					uuid.New().String(),
					1),
				nil,
			)
			r.Header.Set("Authorization", "Bearer "+GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.Nil, []byte("Hans")).String(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			}))
			r.Header.Set("X-Test", "test")
			return r
		}(),
	}, {
		Name: "internal client error",

		App: func(t *testing.T, self *testCase) *mapp.App {
			app := new(mapp.App)
			app.On("GetSettings", contextMatcher).
				Return(model.Settings{
					ConnectionString: self.ConnString,
				}, nil)
			return app
		},
		ConnString: validConnString,
		Req: func() *http.Request {
			r, _ := http.NewRequestWithContext(
				ctxWithoutLog,
				http.MethodGet,
				"http://localhost"+APIURLManagement+strings.Replace(
					APIURLDeviceTwin,
					":id",
					uuid.New().String(),
					1),
				nil,
			)
			r.Header.Set("Authorization", "Bearer "+GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.Nil, []byte("Hans")).String(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			}))
			return r
		}(),
		ClientError: errors.New("internal error"),
		Code:        http.StatusBadGateway,
		Body:        "failed to proxy request to IoT Hub",
	}, {
		Name: "error/no connection string",

		App: func(t *testing.T, self *testCase) *mapp.App {
			app := new(mapp.App)
			app.On("GetSettings", contextMatcher).
				Return(model.Settings{}, nil)
			return app
		},
		ConnString: validConnString,
		Req: func() *http.Request {
			r, _ := http.NewRequestWithContext(
				ctxWithoutLog,
				http.MethodGet,
				"http://localhost"+APIURLManagement+strings.Replace(
					APIURLDeviceTwin,
					":id",
					uuid.New().String(),
					1),
				nil,
			)
			r.Header.Set("Authorization", "Bearer "+GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.Nil, []byte("Hans")).String(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			}))
			return r
		}(),
		Code: http.StatusConflict,
		Body: ErrMissingConnectionString.Error(),
	}, {
		Name: "error/fail to get settings",

		App: func(t *testing.T, self *testCase) *mapp.App {
			app := new(mapp.App)
			app.On("GetSettings", contextMatcher).
				Return(model.Settings{}, errors.New("internal error"))
			return app
		},
		ConnString: validConnString,
		Req: func() *http.Request {
			r, _ := http.NewRequestWithContext(
				ctxWithoutLog,
				http.MethodGet,
				"http://localhost"+APIURLManagement+strings.Replace(
					APIURLDeviceTwin,
					":id",
					uuid.New().String(),
					1),
				nil,
			)
			r.Header.Set("Authorization", "Bearer "+GenerateJWT(identity.Identity{
				Subject: uuid.NewSHA1(uuid.Nil, []byte("Hans")).String(),
				Tenant:  "123456789012345678901234",
				IsUser:  true,
			}))
			return r
		}(),
		Code: http.StatusInternalServerError,
		Body: http.StatusText(http.StatusInternalServerError),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			app := tc.App(t, &tc)
			defer app.AssertExpectations(t)
			var bodyCopy []byte
			if tc.Req.Body != nil {
				var err error
				bodyCopy, err = io.ReadAll(tc.Req.Body)
				require.NoError(t, err, "failed to setup test case")
				tc.Req.Body = io.NopCloser(bytes.NewReader(bodyCopy))
			}

			// Create test server for assessing that the request
			// is proxied correctly
			assertionClient := &http.Client{
				Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
					if tc.ClientError != nil {
						return nil, tc.ClientError
					}
					authz := r.Header.Get(HdrKeyAuthz)
					validateAuthz(t, tc.ConnString.Key, authz)
					delete(tc.Req.Header, HdrKeyAuthz)
					// Check that headers equal
					for k, v := range tc.Req.Header {
						rv, ok := r.Header[k]
						if assert.Truef(t, ok,
							"Header '%s' does not exist in proxy request.",
							k,
						) {
							assert.Equalf(t, v, rv, "Header '%s' does not "+
								"match proxy request", k)
						}
					}

					// Check that body matches
					if r.Body != nil {
						// Transform body to azure schema
						var m map[string]interface{}
						expectedBody := bodyCopy
						err := json.Unmarshal(expectedBody, &m)
						require.NoError(t, err)
						if p, ok := m["properties"]; ok {
							m["properties"] = map[string]interface{}{
								"desired": p,
							}
							expectedBody, _ = json.Marshal(m)
						}

						rb, _ := io.ReadAll(r.Body)
						assert.Contains(t,
							string(rb),
							string(expectedBody),
							"Proxy request body does not match",
						)
					}
					if t.Failed() {
						t.FailNow()
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte("PASS"))),
						Header:     r.Header,
					}, nil
				}),
			}
			router := NewRouter(app, NewConfig().
				SetClient(assertionClient),
			)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tc.Req)
			if tc.Code > 0 {
				assert.Equal(t, tc.Code, w.Code)
				hdr := w.Header()
				for k, v := range tc.Header {
					assert.Equalf(t, v, hdr[k],
						"header '%s' does not match expected value", k,
					)
				}
				switch typ := tc.Body.(type) {
				case *regexp.Regexp:
					assert.True(t, typ.Match(w.Body.Bytes()),
						"body does not match expected pattern",
					)
				case string:
					assert.Contains(t, w.Body.String(), typ)
				case []byte:
					assert.Equal(t, typ, w.Body.Bytes())

				case nil:

				default:
					assert.Fail(t, "I don't know how to assert response body")
				}
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "PASS", w.Body.String())
				hdr := w.Header()
				for k, v := range tc.Req.Header {
					if assert.Contains(t, hdr, k) {
						assert.Equal(t, v, hdr[k])
					}
				}
			}
		})
	}
}
