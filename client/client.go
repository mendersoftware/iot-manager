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

package client

import (
	"crypto/tls"
	"net"
	"net/http"
	"syscall"
	"time"

	inet "github.com/mendersoftware/iot-manager/internal/net"
)

const (
	ParamAlgorithmType = "X-Men-Algorithm"
	//	ParamExpire        = "X-Men-Expire"
	ParamSignedHeaders = "X-Men-Signedheaders"
	ParamSignature     = "X-Men-Signature"

	AlgorithmTypeHMAC256 = "MEN-HMAC-SHA256"
)

func New() *http.Client {
	return &http.Client{
		Transport: NewTransport(),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func addrIsGlobalUnicast(network, address string, _ syscall.RawConn) error {
	ip := net.ParseIP(address)
	if ip == nil {
		return &net.ParseError{
			Type: "IP address",
			Text: address,
		}
	} else if !inet.IsGlobalUnicast(ip) {
		return net.InvalidAddrError("destination address is in reserved address range")
	}
	return nil
}

func NewTransport() http.RoundTripper {
	dialer := &net.Dialer{
		Control: addrIsGlobalUnicast,
	}
	tlsDialer := &tls.Dialer{
		NetDialer: dialer,
	}
	return &http.Transport{
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		DialTLSContext:        tlsDialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
