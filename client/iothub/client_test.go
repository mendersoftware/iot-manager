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

package iothub

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mendersoftware/azure-iot-manager/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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
	const testKey = "_TESTING"
	client := NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	mod, err := client.GetDeviceTwin(ctx, externalCS, externalDeviceID)
	if assert.NoError(t, err) {
		assert.Equal(t, externalDeviceID, mod.DeviceID)
	}
	if t.Failed() {
		t.FailNow()
	}

	var nextValue uint32
	if cur, ok := mod.Properties.Desired[testKey].(float64); ok {
		nextValue = uint32(cur) + 1
	}

	err = client.UpdateDeviceTwin(ctx, externalCS, externalDeviceID, &DeviceTwinUpdate{
		Properties: UpdateProperties{
			Desired: map[string]interface{}{
				testKey: nextValue,
			},
		},
	})

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	modUpdated, err := client.GetDeviceTwin(ctx, externalCS, externalDeviceID)
	if assert.NoError(t, err) {
		value, ok := modUpdated.Properties.Desired[testKey].(float64)
		if assert.True(t, ok, "Updated twin does not contain update value") {
			assert.Equal(t, nextValue, uint32(value), "property does not match update")
		}
	}

	cur, err := client.GetDeviceTwins(ctx, externalCS)
	assert.NoError(t, err)
	var v DeviceTwin
	for cur.Next(ctx) {
		err := cur.Decode(&v)
		require.NoError(t, err)
		b, _ := json.Marshal(v)
		t.Log(string(b))
	}
	err = cur.Decode(v)
	assert.EqualError(t, err, io.EOF.Error())
}

type deviceProducer struct {
	deviceNum  int32
	maxDevices int32
	t          *testing.T
}

func maybe() bool {
	var b [1]byte
	_, _ = rand.Read(b[:])
	return b[0]&0x01 > 0
}

func maybeConnected() string {
	if maybe() {
		return "Connected"
	} else {
		return "Disconnected"
	}
}

func genEtag() string {
	var binTag [4]byte
	_, _ = rand.Read(binTag[:])
	return base64.StdEncoding.EncodeToString(binTag[:])
}

func (h *deviceProducer) produceDevice() *DeviceTwin {
	deviceNum := atomic.AddInt32(&h.deviceNum, 1)
	return &DeviceTwin{
		AuthenticationType: "sas",
		Capabilities: DeviceCapabilities{
			IOTEdge: maybe(),
		},
		CloudToDeviceMessageCount: int64(rand.Int31()),
		ConnectionState:           maybeConnected(),
		DeviceEtag:                genEtag(),
		DeviceID:                  fmt.Sprintf("test-device-%03x", deviceNum),
		ETag:                      genEtag(),
		LastActivityTime:          time.Now().Format(time.RFC3339Nano),
		Properties: TwinProperties{
			Desired: map[string]interface{}{
				"device_num":  deviceNum,
				"good_device": true,
			},
			Reported: map[string]interface{}{
				"device_num":  deviceNum,
				"good_device": maybe(),
			},
		},
		Status:  "enabled",
		Version: rand.Int31(),
	}
}

func (h *deviceProducer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	count, err := strconv.ParseInt(r.Header.Get(hdrKeyCount), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var nextCount int32
	tok := r.Header.Get(hdrKeyContToken)
	if tok != "" {
		rawCount, err := base64.StdEncoding.DecodeString(tok)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if len(rawCount) < 4 {
			http.Error(w, "bad continuation token", http.StatusBadRequest)
		}
		nextCount = int32(binary.BigEndian.Uint32(rawCount))
	}
	assert.Equal(h.t, h.deviceNum, nextCount)
	resCount := int64(h.maxDevices - nextCount)
	if resCount > count {
		resCount = count
	}
	res := make([]*DeviceTwin, resCount)
	for i := int64(0); i < resCount; i++ {
		res[i] = h.produceDevice()
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(h.deviceNum))
	if h.deviceNum < h.maxDevices {
		w.Header().Set(hdrKeyContToken, base64.StdEncoding.EncodeToString(hdr[:]))
	}
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.Encode(res)
}

func (h *deviceProducer) RoundTrip(r *http.Request) (*http.Response, error) {
	select {
	case <-r.Context().Done():
		return nil, r.Context().Err()
	default:
		// pass
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Result(), nil
}

type ContextExpireInAMinute struct{ context.Context }

func (ctx ContextExpireInAMinute) Deadline() (time.Time, bool) {
	return time.Now().Add(time.Minute), true
}

type ContextExpireOnDone struct {
	context.Context
	Chan   chan struct{}
	DoneIn int32
	err    error
}

func NewContextExpireOnDone(in int32) context.Context {
	return &ContextExpireOnDone{
		Context: context.Background(),
		DoneIn:  in,
	}
}

func (ctx *ContextExpireOnDone) Done() <-chan struct{} {
	if ctx.Chan == nil {
		ctx.Chan = make(chan struct{})
	}
	res := atomic.AddInt32(&ctx.DoneIn, -1)
	if res <= 0 {
		select {
		case <-ctx.Chan:
			// pass (already closed)
		default:
			close(ctx.Chan)
			ctx.err = context.DeadlineExceeded
		}
	}
	return ctx.Chan
}

func (ctx *ContextExpireOnDone) Err() error {
	return ctx.err
}

func TestGetDevices(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string

		CTX context.Context

		ConnStr    *model.ConnectionString
		NumDevices int32
		LastError  error

		Error error
	}{{
		Name: "ok",

		CTX:        context.Background(),
		NumDevices: 101,

		Error: nil,
	}, {
		Name: "ok/with expire",

		CTX:        ContextExpireInAMinute{Context: context.Background()},
		NumDevices: 101,

		Error: nil,
	}, {
		Name: "error/context cancelled",

		CTX: func() context.Context {
			ctx, cancel := context.WithCancel(context.TODO())
			cancel()
			return ctx
		}(),
		NumDevices: 101,

		Error: context.Canceled,
	}, {
		Name: "error/context expires on next",

		CTX:        NewContextExpireOnDone(2),
		NumDevices: 101,
		LastError:  context.DeadlineExceeded,
	}, {
		Name: "error/nil context",

		Error: errors.New("iothub: failed to prepare request"),
	}, {
		Name: "error/invalid connection string",

		ConnStr: &model.ConnectionString{
			HostName: "localhost",
		},
		Error: errors.New("iothub: failed to prepare request: invalid connection string"),
	}}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			httpClient := &http.Client{
				Transport: &deviceProducer{
					t:          t,
					maxDevices: tc.NumDevices,
				},
			}
			client := NewClient(NewOptions().SetClient(httpClient))
			connStr := tc.ConnStr
			if connStr == nil {
				connStr = &model.ConnectionString{
					Key:             []byte("c3VwZXIgc2VjcmV0Cg=="),
					HostName:        "localhost",
					GatewayHostName: "localhost:8080",
					Name:            "admin_sas",
				}
			}

			cur, err := client.GetDeviceTwins(tc.CTX, connStr)
			if tc.Error != nil {
				if assert.Error(t, err) {
					assert.Regexp(t, tc.Error.Error(), err.Error())
				}
			} else if assert.Nil(t, err) && assert.NotNil(t, cur) {
				if tc.LastError == nil {
					tc.LastError = io.EOF
				}
				var twin DeviceTwin
				for cur.Next(tc.CTX) {
					err = cur.Decode(&twin)
					assert.NoError(t, err)
				}
				err := cur.Decode(&twin)
				if assert.Error(t, err) {
					assert.Regexp(t, tc.LastError.Error(), err.Error())
				}
			}
		})
	}
}
