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

package workflows

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/requestid"
	"github.com/mendersoftware/go-lib-micro/rest.utils"
	"github.com/pkg/errors"

	common "github.com/mendersoftware/azure-iot-manager/client"
)

const (
	URICheckHealth  = "/api/v1/health"
	URISubmitConfig = "/api/v1/workflow/submit_configuration"
)

const (
	defaultTimeout = time.Duration(10) * time.Second
)

// Client is the workflows client
//go:generate ../../utils/mockgen.sh
type Client interface {
	CheckHealth(ctx context.Context) error
	SubmitDeviceConfiguration(ctx context.Context, devID string, config map[string]string) error
}

type Options struct {
	Client *http.Client
}

func NewOptions(opts ...*Options) *Options {
	ret := new(Options)
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if opt.Client != nil {
			ret.Client = opt.Client
		}
	}
	return ret
}

func (opts *Options) SetClient(client *http.Client) *Options {
	opts.Client = client
	return opts
}

// NewClient returns a new workflows client
func NewClient(url string, opts ...*Options) Client {
	opt := NewOptions(opts...)
	if opt.Client == nil {
		opt.Client = new(http.Client)
	}

	return &client{
		url:    strings.TrimRight(url, "/"),
		Client: opt.Client,
	}
}

type client struct {
	url string
	*http.Client
}

func (c *client) CheckHealth(ctx context.Context) error {
	var (
		apiErr rest.Error
	)

	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	req, _ := http.NewRequestWithContext(
		ctx, "GET", c.url+URICheckHealth, nil,
	)

	rsp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode >= http.StatusOK && rsp.StatusCode < 300 {
		return nil
	}
	decoder := json.NewDecoder(rsp.Body)
	err = decoder.Decode(&apiErr)
	if err != nil {
		return errors.Errorf("health check HTTP error: %s", rsp.Status)
	}
	return &apiErr
}

func (c *client) SubmitDeviceConfiguration(
	ctx context.Context,
	devID string,
	config map[string]string,
) error {
	var workflow struct {
		TenantID  string            `json:"tenant_id"`
		DeviceID  string            `json:"device_id"`
		RequestID string            `json:"request_id"`
		Config    map[string]string `json:"configuration"`
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	workflow.DeviceID = devID
	workflow.RequestID = requestid.FromContext(ctx)
	workflow.Config = config
	if id := identity.FromContext(ctx); id != nil {
		workflow.TenantID = id.Tenant
	}

	b, _ := json.Marshal(workflow)
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		c.url+URISubmitConfig,
		bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "workflows: failed to prepare request")
	}
	rsp, err := c.Do(req)
	if err != nil {
		return errors.Wrap(err, "workflows: failed to execute request")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode >= 400 {
		return common.HTTPError{Code: rsp.StatusCode}
	}
	return nil
}
