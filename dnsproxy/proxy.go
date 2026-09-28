/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

// Package dnsproxy provides a small loopback-only DNS-over-HTTPS forwarder.
package dnsproxy

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"golang.zx2c4.com/wireguard/windows/doh"
)

const (
	DefaultAddress         = "127.0.0.1:53"
	DefaultQueryTimeout    = 10 * time.Second
	DefaultMaxQuerySize    = 65535
	DefaultMaxResponseSize = 65535
	DefaultMaxConcurrent   = 128
	DefaultTCPReadTimeout  = 10 * time.Second
	DefaultTCPWriteTimeout = 10 * time.Second
)

var (
	ErrNotRunning       = errors.New("DNS proxy is not running")
	ErrInvalidQuery     = errors.New("invalid DNS query")
	ErrQueryTooLarge    = errors.New("DNS query is too large")
	ErrResponseTooLarge = errors.New("DNS response is too large")
	ErrConcurrencyLimit = errors.New("DNS proxy concurrency limit reached")
)

type Options struct {
	Address         string
	Client          *doh.Client
	Endpoint        string
	QueryTimeout    time.Duration
	MaxQuerySize    int
	MaxResponseSize int
	MaxConcurrent   int
}

type Status struct {
	Running       bool
	UDPAddress    string
	TCPAddress    string
	ActiveQueries int
	LastError     error
}

type Proxy struct {
	client          *doh.Client
	address         string
	queryTimeout    time.Duration
	maxQuerySize    int
	maxResponseSize int
	maxConcurrent   int

	udp *net.UDPConn
	tcp net.Listener

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	sem    chan struct{}
	mu     sync.Mutex
	status Status
	closed atomic.Bool
}

func New(options Options) (*Proxy, error) {
	address := options.Address
	if address == "" {
		address = DefaultAddress
	}
	if err := validateLoopbackAddress(address); err != nil {
		return nil, err
	}
	client := options.Client
	if client == nil {
		if options.Endpoint == "" {
			return nil, errors.New("DoH endpoint is required")
		}
		var err error
		client, err = doh.NewClient(options.Endpoint, doh.Options{Timeout: options.QueryTimeout})
		if err != nil {
			return nil, err
		}
	}
	queryTimeout := options.QueryTimeout
	if queryTimeout <= 0 {
		queryTimeout = DefaultQueryTimeout
	}
	maxQuerySize := options.MaxQuerySize
	if maxQuerySize <= 0 {
		maxQuerySize = DefaultMaxQuerySize
	}
	maxResponseSize := options.MaxResponseSize
	if maxResponseSize <= 0 {
		maxResponseSize = DefaultMaxResponseSize
	}
	maxConcurrent := options.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = DefaultMaxConcurrent
	}
	if maxQuerySize > DefaultMaxQuerySize || maxQuerySize < 12 {
		return nil, fmt.Errorf("max query size must be between 12 and %d", DefaultMaxQuerySize)
	}
	if maxResponseSize > DefaultMaxResponseSize || maxResponseSize < 12 {
		return nil, fmt.Errorf("max response size must be between 12 and %d", DefaultMaxResponseSize)
	}
	return &Proxy{
		client:          client,
		address:         address,
		queryTimeout:    queryTimeout,
		maxQuerySize:    maxQuerySize,
		maxResponseSize: maxResponseSize,
		maxConcurrent:   maxConcurrent,
		status:          Status{},
	}, nil
}

func (p *Proxy) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.udp != nil || p.tcp != nil {
		return errors.New("DNS proxy is already running")
	}
	udp, err := net.ListenUDP("udp", mustUDPAddr(p.address))
	if err != nil {
		p.status.LastError = err
		return fmt.Errorf("listen for DNS over UDP on %s: %w", p.address, err)
	}
	tcp, err := net.Listen("tcp", p.address)
	if err != nil {
		udp.Close()
		p.status.LastError = err
		return fmt.Errorf("listen for DNS over TCP on %s: %w", p.address, err)
	}
	p.udp, p.tcp = udp, tcp
	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.sem = make(chan struct{}, p.maxConcurrent)
	p.closed.Store(false)
	p.status = Status{Running: true, UDPAddress: udp.LocalAddr().String(), TCPAddress: tcp.Addr().String()}
	p.wg.Add(2)
	go p.serveUDP(udp)
	go p.serveTCP(tcp)
	return nil
}

func (p *Proxy) Stop() error {
	p.mu.Lock()
	if p.udp == nil && p.tcp == nil {
		p.mu.Unlock()
		return ErrNotRunning
	}
	udp, tcp, cancel := p.udp, p.tcp, p.cancel
	p.udp, p.tcp, p.cancel = nil, nil, nil
	p.status.Running = false
	p.mu.Unlock()
	p.closed.Store(true)
	if cancel != nil {
		cancel()
	}
	udpErr := udp.Close()
	tcpErr := tcp.Close()
	p.wg.Wait()
	if udpErr != nil && !errors.Is(udpErr, net.ErrClosed) {
		return udpErr
	}
	if tcpErr != nil && !errors.Is(tcpErr, net.ErrClosed) {
		return tcpErr
	}
	return nil
}

func (p *Proxy) Status() Status {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status
}

func (p *Proxy) serveUDP(conn *net.UDPConn) {
	defer p.wg.Done()
	buf := make([]byte, p.maxQuerySize+1)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if p.closed.Load() || errors.Is(err, net.ErrClosed) {
				return
			}
			p.setError(err)
			continue
		}
		query := append([]byte(nil), buf[:n]...)
		if n > p.maxQuerySize {
			continue
		}
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			if !p.acquire() {
				p.setError(ErrConcurrencyLimit)
				return
			}
			defer p.release()
			response, err := p.forward(query)
			if err != nil {
				p.setError(err)
				return
			}
			_, _ = conn.WriteToUDP(response, addr)
		}()
	}
}

func (p *Proxy) serveTCP(listener net.Listener) {
	defer p.wg.Done()
	for {
		conn, err := listener.Accept()
		if err != nil {
			if p.closed.Load() || errors.Is(err, net.ErrClosed) {
				return
			}
			p.setError(err)
			continue
		}
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.serveTCPConn(conn)
		}()
	}
}

func (p *Proxy) serveTCPConn(conn net.Conn) {
	defer conn.Close()
	for {
		if err := conn.SetDeadline(time.Now().Add(DefaultTCPReadTimeout)); err != nil {
			return
		}
		var length [2]byte
		if _, err := io.ReadFull(conn, length[:]); err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
				p.setError(err)
			}
			return
		}
		n := int(binary.BigEndian.Uint16(length[:]))
		if n < 12 || n > p.maxQuerySize {
			p.setError(ErrInvalidQuery)
			return
		}
		query := make([]byte, n)
		if _, err := io.ReadFull(conn, query); err != nil {
			return
		}
		if !p.acquire() {
			p.setError(ErrConcurrencyLimit)
			return
		}
		response, err := p.forward(query)
		p.release()
		if err != nil {
			p.setError(err)
			return
		}
		if len(response) > DefaultMaxResponseSize {
			p.setError(ErrResponseTooLarge)
			return
		}
		if err := conn.SetWriteDeadline(time.Now().Add(DefaultTCPWriteTimeout)); err != nil {
			return
		}
		binary.BigEndian.PutUint16(length[:], uint16(len(response)))
		if _, err := conn.Write(length[:]); err != nil {
			return
		}
		if _, err := conn.Write(response); err != nil {
			return
		}
	}
}

func (p *Proxy) forward(query []byte) ([]byte, error) {
	if err := validateQuery(query, p.maxQuerySize); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(p.ctx, p.queryTimeout)
	defer cancel()
	response, err := p.client.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(response) > p.maxResponseSize {
		return nil, ErrResponseTooLarge
	}
	if len(response) < 12 {
		return nil, ErrInvalidQuery
	}
	return response, nil
}

func (p *Proxy) acquire() bool {
	select {
	case p.sem <- struct{}{}:
		p.mu.Lock()
		p.status.ActiveQueries++
		p.mu.Unlock()
		return true
	default:
		return false
	}
}

func (p *Proxy) release() {
	<-p.sem
	p.mu.Lock()
	p.status.ActiveQueries--
	p.mu.Unlock()
}

func (p *Proxy) setError(err error) {
	if err == nil {
		return
	}
	p.mu.Lock()
	p.status.LastError = err
	p.mu.Unlock()
}

func validateLoopbackAddress(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid DNS proxy address %q: %w", address, err)
	}
	ip, err := netip.ParseAddr(host)
	if err != nil || !ip.IsLoopback() {
		return fmt.Errorf("DNS proxy address must be an IPv4 or IPv6 loopback address: %q", address)
	}
	if _, err := net.ResolveUDPAddr("udp", address); err != nil {
		return fmt.Errorf("invalid DNS proxy address %q: %w", address, err)
	}
	return nil
}

func mustUDPAddr(address string) *net.UDPAddr {
	addr, _ := net.ResolveUDPAddr("udp", address)
	return addr
}

func validateQuery(query []byte, max int) error {
	if len(query) < 12 {
		return ErrInvalidQuery
	}
	if len(query) > max {
		return ErrQueryTooLarge
	}
	return nil
}

// ValidateEndpoint is exposed for callers that want to reject plaintext DNS
// configuration before constructing a proxy.
func ValidateEndpoint(endpoint string) error {
	u, err := url.ParseRequestURI(endpoint)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" {
		return errors.New("DNS proxy requires an HTTPS DoH endpoint")
	}
	return nil
}
