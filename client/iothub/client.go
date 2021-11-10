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
	"strings"
	"sync"
	"time"

	common "github.com/mendersoftware/azure-iot-manager/client"
	"github.com/mendersoftware/azure-iot-manager/model"

	"github.com/pkg/errors"
)

const (
	uriTwin      = "/twins"
	uriQueryTwin = "/devices/query"

	hdrKeyContentType = "Content-Type"
	hdrKeyContToken   = "X-Ms-Continuation"
	hdrKeyCount       = "X-Ms-Max-Item-Count"

	// https://docs.microsoft.com/en-us/rest/api/iothub/service/devices
	APIVersion = "2021-04-12"
)

const (
	defaultTTL = time.Minute
)

const (
	hdrKeyAuthorization = "Authorization"
)

//nolint:lll
//go:generate ../../utils/mockgen.sh
type Client interface {
	GetDeviceTwins(ctx context.Context, sas *model.ConnectionString) (Cursor, error)
	GetDeviceTwin(ctx context.Context, sas *model.ConnectionString, id string) (*DeviceTwin, error)
	UpdateDeviceTwin(ctx context.Context, sas *model.ConnectionString, id string, r *DeviceTwinUpdate) error
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
		if opt == nil {
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

type cursor struct {
	err     error
	buf     bytes.Buffer
	mut     *sync.Mutex
	cs      *model.ConnectionString
	client  *http.Client
	req     *http.Request
	dec     *json.Decoder
	current json.RawMessage
}

func (cur *cursor) fetchPage(ctx context.Context) error {
	cur.buf.Reset()
	req := cur.req.WithContext(ctx)
	req.Body, _ = cur.req.GetBody()
	var expireAt time.Time
	if dl, ok := ctx.Deadline(); ok {
		expireAt = dl
	} else {
		expireAt = time.Now().Add(defaultTTL)
	}
	req.Header.Set(hdrKeyAuthorization, cur.cs.Authorization(expireAt))
	rsp, err := cur.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "iothub: failed to execute request")
	}
	defer rsp.Body.Close()
	if rsp.StatusCode >= 400 {
		return errors.Wrapf(err,
			"iothub: received unexpected status code from server: %s",
			rsp.Status,
		)
	}
	_, err = io.Copy(&cur.buf, rsp.Body)
	if err != nil {
		return errors.Wrap(err, "iothub: failed to buffer HTTP response")
	}
	cur.req.Header.Set(hdrKeyContToken, rsp.Header.Get(hdrKeyContToken))
	cur.dec = json.NewDecoder(&cur.buf)
	tkn, err := cur.dec.Token()
	if err != nil {
		return errors.Wrap(err, "iothub: failed to decode response from hub")
	} else if tkn != json.Delim('[') {
		return errors.Wrap(err, "iothub: unexpected json response from hub")
	}
	return nil
}

func (cur *cursor) Next(ctx context.Context) bool {
	if cur.err != nil {
		return false
	}
	cur.mut.Lock()
	defer cur.mut.Unlock()
	if cur.dec == nil || !cur.dec.More() {
		if cur.req.Header.Get(hdrKeyContToken) == "" {
			cur.err = io.EOF
			return false
		}
		err := cur.fetchPage(ctx)
		if err != nil {
			cur.err = err
			return false
		}
	}
	err := cur.dec.Decode(&cur.current)
	if err != nil {
		cur.err = errors.Wrap(err, "iothub: failed to retrieve next element")
		return false
	}
	return true
}

func (cur *cursor) Decode(v interface{}) error {
	if cur.err != nil {
		return cur.err
	}
	return json.Unmarshal(cur.current, v)
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
