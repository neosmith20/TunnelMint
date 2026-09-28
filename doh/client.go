/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

// Package doh provides the small HTTPS transport primitive used for DNS-over-HTTPS.
package doh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultTimeout bounds one DoH request when the caller does not provide a shorter context.
	DefaultTimeout = 10 * time.Second
	// DefaultMaxResponseSize prevents a resolver from returning an unbounded body.
	DefaultMaxResponseSize int64 = 1 << 20
	// MaxQuerySize is the largest DNS wire-format query accepted by the transport.
	MaxQuerySize = 65535
)

// Options controls the transport limits and allows tests or later integration
// layers to provide an HTTP client with a custom dial path.
type Options struct {
	HTTPClient      *http.Client
	Timeout         time.Duration
	MaxResponseSize int64
}

// Client sends raw DNS wire-format queries to one validated HTTPS endpoint.
type Client struct {
	endpoint        *url.URL
	httpClient      *http.Client
	timeout         time.Duration
	maxResponseSize int64
}

// NewClient validates endpoint and returns a DoH client using normal system
// TLS certificate validation. A custom HTTP client may provide a later
// bootstrap-aware dialer; its TLS configuration is otherwise left intact.
func NewClient(endpoint string, options Options) (*Client, error) {
	parsed, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid DoH endpoint: %w", err)
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return nil, fmt.Errorf("DoH endpoint must use HTTPS")
	}
	if parsed.Hostname() == "" {
		return nil, fmt.Errorf("DoH endpoint must contain a hostname")
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	maxResponseSize := options.MaxResponseSize
	if maxResponseSize <= 0 {
		maxResponseSize = DefaultMaxResponseSize
	}
	return &Client{
		endpoint:        parsed,
		httpClient:      httpClient,
		timeout:         timeout,
		maxResponseSize: maxResponseSize,
	}, nil
}

// Query sends query as a DNS-over-HTTPS POST and returns the raw DNS response.
func (c *Client) Query(ctx context.Context, query []byte) ([]byte, error) {
	if len(query) == 0 {
		return nil, fmt.Errorf("DNS query is empty")
	}
	if len(query) > MaxQuerySize {
		return nil, fmt.Errorf("DNS query exceeds %d bytes", MaxQuerySize)
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestCtx, http.MethodPost, c.endpoint.String(), bytes.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("create DoH request: %w", err)
	}
	request.Header.Set("Accept", "application/dns-message")
	request.Header.Set("Content-Type", "application/dns-message")

	// Reject redirects entirely. This keeps the endpoint identity and TLS
	// policy explicit while still allowing the caller to provide a custom
	// transport for endpoint dialing.
	httpClient := *c.httpClient
	httpClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("DoH request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("DoH endpoint returned HTTP status %s", response.Status)
	}
	contentType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || !strings.EqualFold(contentType, "application/dns-message") {
		return nil, fmt.Errorf("DoH endpoint returned incompatible Content-Type %q", response.Header.Get("Content-Type"))
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, c.maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("read DoH response: %w", err)
	}
	if int64(len(body)) > c.maxResponseSize {
		return nil, fmt.Errorf("DoH response exceeds %d bytes", c.maxResponseSize)
	}
	return body, nil
}
