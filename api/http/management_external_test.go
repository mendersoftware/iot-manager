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
	"flag"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	mapp "github.com/mendersoftware/azure-iot-manager/app/mocks"
	"github.com/mendersoftware/azure-iot-manager/model"
	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	externalCS       *model.ConnectionString
	externalDeviceID string
)

func parseConnString(connection string) error {
	var err error
	externalCS, err = model.ParseConnectionString(connection)
	return err
}

func init() {
	flag.Func("test.connection-string",
		"Connection string for external iothub "+
			"(overwrite with env var TEST_CONNECTION_STRING).",
		parseConnString)
	flag.StringVar(&externalDeviceID,
		"test.device-id",
		"",
		"The id of a device on the iothub pointed to by connection-string"+
			" (overwrite with env TEST_DEVICE_ID).")
	cStr, ok := os.LookupEnv("TEST_CONNECTION_STRING")
	if ok {
		externalCS, _ = model.ParseConnectionString(cStr)
	}
	idStr, ok := os.LookupEnv("TEST_DEVICE_ID")
	if ok {
		externalDeviceID = idStr
	}

	testing.Init()
}

// TestIOTHubExternal runs against a real IoT Hub using the provided command line
// arguments / environment variables. The test updates fields in the device's
// desired state, so it's important that the hub-device is not used by a real
// device.
func TestIOTHubExternal(t *testing.T) {
	if externalCS == nil {
		t.Skip("test.connection-string is not provided or valid")
		return
	} else if externalDeviceID == "" {
		t.Skip("test.device-id is not provided nor valid")
		return
	}
	// The following gets the device and updates (increments)
	// the "desired" property "_TESTING" and checks the expected
	// value.
	authz := "Bearer " + GenerateJWT(identity.Identity{
		Subject: "7e57dc61-cd13-4d8a-beee-3cfa885c9cae",
		IsUser:  true,
	})
	w := httptest.NewRecorder()
	const testKey = "_TESTING"
	mockApp := new(mapp.App)
	defer mockApp.AssertExpectations(t)
	mockApp.On("GetSettings", mock.Anything).
		Return(model.Settings{
			ConnectionString: externalCS,
		}, nil)
	handler := NewRouter(mockApp)
	srv := httptest.NewServer(handler)
	client := srv.Client()
	client.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial(network, srv.Listener.Addr().String())
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	uri := "http://localhost" +
		APIURLManagement +
		strings.ReplaceAll(APIURLDeviceTwin, ":id", externalDeviceID)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	req.Header.Set(HdrKeyAuthz, authz)
	rsp, err := client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	defer rsp.Body.Close()
	type bodyOfInterest struct {
		Properties struct {
			Desired map[string]interface{} `json:"desired"`
		} `json:"properties"`
	}
	var boi bodyOfInterest
	dec := json.NewDecoder(rsp.Body)
	err = dec.Decode(&boi)
	require.NoError(t, err)
	var nextValue uint32
	if cur, ok := boi.Properties.Desired[testKey].(float64); ok {
		nextValue = uint32(cur) + 1
	}
	boi.Properties.Desired[testKey] = nextValue
	b, _ := json.Marshal(map[string]interface{}{"properties": boi.Properties.Desired})
	req, _ = http.NewRequestWithContext(ctx, http.MethodPatch, uri, bytes.NewReader(b))
	req.Header.Set(HdrKeyAuthz, authz)
	require.NoError(t, err)
	rspPatch, err := client.Do(req)
	require.NoError(t, err)
	defer rspPatch.Body.Close()
	assert.Equal(t, http.StatusOK, rspPatch.StatusCode)

	var boiUpdated bodyOfInterest
	req, _ = http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	req.Header.Set(HdrKeyAuthz, authz)
	rspGet, err := client.Do(req)
	require.NoError(t, err)
	defer rspGet.Body.Close()
	assert.Equal(t, http.StatusOK, rspGet.StatusCode)
	dec = json.NewDecoder(rspGet.Body)
	_ = dec.Decode(&boiUpdated)
	if updatedFloat, ok := boiUpdated.Properties.
		Desired[testKey].(float64); assert.True(t, ok) {
		assert.Equal(t, nextValue, uint32(updatedFloat))
	}
}
