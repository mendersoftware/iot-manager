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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	common "github.com/mendersoftware/azure-iot-manager/client"
	"github.com/mendersoftware/azure-iot-manager/model"

	"github.com/pkg/errors"
)

const (
	uriTwin      = "/twins"
	uriDevices   = "/devices"
	uriQueryTwin = uriDevices + "/query"

	hdrKeyContentType = "Content-Type"
	hdrKeyContToken   = "X-Ms-Continuation"
	hdrKeyCount       = "X-Ms-Max-Item-Count"

	// https://docs.microsoft.com/en-us/rest/api/iothub/service/devices
	APIVersion = "2021-04-12"
)

func uriDevice(id string) string {
	return uriDevices + "/" + url.QueryEscape(id)
}

const (
	defaultTTL = time.Minute
)

const (
	hdrKeyAuthorization = "Authorization"
)

//nolint:lll
//go:generate ../../utils/mockgen.sh
type Client interface {
	GetDeviceTwins(ctx context.Context, cs *model.ConnectionString) (Cursor, error)
	GetDeviceTwin(ctx context.Context, cs *model.ConnectionString, id string) (*DeviceTwin, error)
	UpdateDeviceTwin(ctx context.Context, cs *model.ConnectionString, id string, r *DeviceTwinUpdate) error

	// UpsertDevice create or update a device with the given ID. If a device
	// is created, the IoT Hub will generate a new 256-bit primary and
	// secondary key used to construct the device connection string:
	// primaryCS := &model.ConnectionString{
	// 	HostName: cs.HostName,
	// 	DeviceID: Device.DeviceID,
	// 	Key:      Device.Auth.SymmetricKey.Primary,
	// }.String()
	// secondary := &model.ConnectionString{
	// 	HostName: cs.HostName,
	// 	DeviceID: Device.DeviceID,
	// 	Key:      Device.Auth.SymmetricKey.Secondary,
	// }.String()
	UpsertDevice(ctx context.Context, cs *model.ConnectionString, id string, deviceUpdate ...*Device) (*Device, error)
	DeleteDevice(ctx context.Context, cs *model.ConnectionString, id string) error
}

type client struct {
	*http.Client
}

type Options struct {
	Client *http.Client
}

func NewOptions(opts ...*Options) *Options {
	opt := new(Options)
	for _, o := range opts {
		if o == nil {
			continue
		}
		if o.Client != nil {
			opt.Client = o.Client
		}
	}
	return opt
}

func (opt *Options) SetClient(client *http.Client) *Options {
	opt.Client = client
	return opt
}

func NewClient(options ...*Options) Client {
	opts := NewOptions(options...)
	if opts.Client == nil {
		opts.Client = new(http.Client)
	}
	return &client{
		Client: opts.Client,
	}
}

func (c *client) NewRequestWithContext(
	ctx context.Context,
	cs *model.ConnectionString,
	method, urlPath string,
	body io.Reader,
) (*http.Request, error) {
	if err := cs.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid connection string")
	}
	hostname := cs.HostName
	if cs.GatewayHostName != "" {
		hostname = cs.GatewayHostName
	}
	uri := "https://" + hostname + "/" +
		strings.TrimPrefix(urlPath, "/")
	if idx := strings.IndexRune(uri, '?'); idx < 0 {
		uri += "?"
	}
	uri += "api-version=" + APIVersion
	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return req, err
	}
	if body != nil {
		req.Header.Set(hdrKeyContentType, "application/json")
	}
	// Ensure that we set the correct Host header (in case GatewayHostName is set)
	req.Host = cs.HostName

	var expireAt time.Time
	if dl, ok := ctx.Deadline(); ok {
		expireAt = dl
	} else {
		expireAt = time.Now().Add(defaultTTL)
	}
	req.Header.Set(hdrKeyAuthorization, cs.Authorization(expireAt))

	return req, err
}

func (c *client) UpsertDevice(ctx context.Context,
	cs *model.ConnectionString,
	deviceID string,
	deviceUpdate ...*Device,
) (*Device, error) {
	dev := mergeDevices(deviceUpdate...)
	dev.DeviceID = deviceID
	b, _ := json.Marshal(dev)
	req, err := c.NewRequestWithContext(
		ctx,
		cs,
		http.MethodPut,
		uriDevice(deviceID),
		bytes.NewReader(b),
	)
	if err != nil {
		return nil, errors.Wrap(err, "iothub: failed to prepare request")
	}
	rsp, err := c.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "iothub: failed to execute request")
	}
	defer rsp.Body.Close()
	if rsp.StatusCode >= 400 {
		return nil, common.HTTPError{Code: rsp.StatusCode}
	}
	dec := json.NewDecoder(rsp.Body)
	if err = dec.Decode(dev); err != nil {
		return nil, errors.Wrap(err, "iothub: failed to decode updated device")
	}
	return dev, nil
}

func (c *client) DeleteDevice(ctx context.Context, cs *model.ConnectionString, id string) error {
	req, err := c.NewRequestWithContext(ctx,
		cs,
		http.MethodDelete,
		uriDevice(id),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "iothub: failed to prepare request")
	}
	req.Header.Set("If-Match", "*")
	rsp, err := c.Do(req)
	if err != nil {
		return errors.Wrap(err, "iothub: failed to execute request")
	}
	defer rsp.Body.Close()
	if rsp.StatusCode >= 400 {
		return common.HTTPError{Code: rsp.StatusCode}
	}
	return nil
}

func (c *client) GetDeviceTwins(
	ctx context.Context, cs *model.ConnectionString,
) (Cursor, error) {
	const (
		SQLQuery              = `{"query":"SELECT * FROM devices"}`
		pageSize, pageSizeStr = 100, "100"
	)
	q := bytes.NewReader([]byte(SQLQuery))
	req, err := c.NewRequestWithContext(ctx, cs, http.MethodPost, uriQueryTwin, q)
	if err != nil {
		return nil, errors.Wrap(err, "iothub: failed to prepare request")
	}
	req.Header.Set(hdrKeyCount, pageSizeStr)

	cur := &cursor{
		mut:    new(sync.Mutex),
		client: c.Client,
		req:    req,
		cs:     cs,
	}
	err = cur.fetchPage(ctx)
	if err != nil {
		return nil, err
	}
	return cur, nil
}

func (c *client) GetDeviceTwin(
	ctx context.Context,
	cs *model.ConnectionString,
	id string,
) (*DeviceTwin, error) {
	uri := uriTwin + "/" + id
	req, err := c.NewRequestWithContext(ctx, cs, http.MethodGet, uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, "iothub: failed to prepare request")
	}

	rsp, err := c.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "iothub: failed to fetch device twin")
	}
	defer rsp.Body.Close()
	if rsp.StatusCode >= 400 {
		return nil, common.HTTPError{
			Code: rsp.StatusCode,
		}
	}
	twin := new(DeviceTwin)
	dec := json.NewDecoder(rsp.Body)
	if err = dec.Decode(twin); err != nil {
		return nil, errors.Wrap(err, "iothub: failed to decode API response")
	}
	rsp.Header.Get("x-ms-continuation")
	return twin, nil
}

func (c *client) UpdateDeviceTwin(
	ctx context.Context,
	cs *model.ConnectionString,
	id string,
	r *DeviceTwinUpdate,
) error {
	method := http.MethodPatch
	if r.Replace {
		method = http.MethodPut
	}

	b, _ := json.Marshal(r)

	req, err := c.NewRequestWithContext(ctx, cs, method, uriTwin+"/"+id, bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "iothub: failed to prepare request")
	}
	rsp, err := c.Do(req)
	if err != nil {
		return errors.Wrap(err, "iothub: failed to submit device twin update")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode >= 400 {
		return common.HTTPError{
			Code: rsp.StatusCode,
		}
	}
	return nil
}
