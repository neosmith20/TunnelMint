/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package dnsproxy

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.zx2c4.com/wireguard/windows/doh"
)

func testQuery(id byte) []byte {
	return []byte{id, 2, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0, 0, 1, 0, 1}
}

func testDoHClient(t *testing.T, handler http.Handler) (*doh.Client, func()) {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	client, err := doh.NewClient(server.URL, doh.Options{HTTPClient: server.Client(), Timeout: time.Second})
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	return client, server.Close
}

func echoHandler(t *testing.T, calls *atomic.Int32) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/dns-message" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read failed", http.StatusBadRequest)
			return
		}
		if len(body) < 12 {
			http.Error(w, "short query", http.StatusBadRequest)
			return
		}
		body[2] |= 0x80
		w.Header().Set("Content-Type", "application/dns-message")
		_, _ = w.Write(body)
	})
}

func newTestProxy(t *testing.T, handler http.Handler, options Options) (*Proxy, func()) {
	t.Helper()
	client, closeServer := testDoHClient(t, handler)
	options.Address = "127.0.0.1:0"
	options.Client = client
	p, err := New(options)
	if err != nil {
		closeServer()
		t.Fatal(err)
	}
	if err := p.Start(); err != nil {
		closeServer()
		t.Fatal(err)
	}
	return p, func() {
		_ = p.Stop()
		closeServer()
	}
}

func TestUDPQuerySuccess(t *testing.T) {
	var calls atomic.Int32
	p, cleanup := newTestProxy(t, echoHandler(t, &calls), Options{})
	defer cleanup()

	conn, err := net.Dial("udp", p.Status().UDPAddress)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	query := testQuery(0x42)
	if _, err := conn.Write(query); err != nil {
		t.Fatal(err)
	}
	response := make([]byte, 65536)
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	n, err := conn.Read(response)
	if err != nil {
		t.Fatal(err)
	}
	want := append([]byte(nil), query...)
	want[2] |= 0x80
	if !bytes.Equal(response[:n], want) {
		t.Fatalf("unexpected response wire data: %x", response[:n])
	}
	if calls.Load() != 1 {
		t.Fatalf("DoH call count = %d, want 1", calls.Load())
	}
}

func TestTCPQuerySuccessAndFraming(t *testing.T) {
	var calls atomic.Int32
	p, cleanup := newTestProxy(t, echoHandler(t, &calls), Options{})
	defer cleanup()

	conn, err := net.Dial("tcp", p.Status().TCPAddress)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	query := testQuery(0x24)
	frame := make([]byte, 2+len(query))
	frame[0] = byte(len(query) >> 8)
	frame[1] = byte(len(query))
	copy(frame[2:], query)
	if _, err := conn.Write(frame); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	var length [2]byte
	if _, err := io.ReadFull(conn, length[:]); err != nil {
		t.Fatal(err)
	}
	response := make([]byte, int(length[0])<<8|int(length[1]))
	if _, err := io.ReadFull(conn, response); err != nil {
		t.Fatal(err)
	}
	if response[0] != query[0] || response[1] != query[1] {
		t.Fatalf("DNS ID or flags changed: %x", response[:2])
	}
	if calls.Load() != 1 {
		t.Fatalf("DoH call count = %d, want 1", calls.Load())
	}
}

func TestConcurrentQueries(t *testing.T) {
	var calls atomic.Int32
	p, cleanup := newTestProxy(t, echoHandler(t, &calls), Options{MaxConcurrent: 32})
	defer cleanup()

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(id byte) {
			defer wg.Done()
			conn, err := net.Dial("udp", p.Status().UDPAddress)
			if err != nil {
				t.Error(err)
				return
			}
			defer conn.Close()
			if _, err := conn.Write(testQuery(id)); err != nil {
				t.Error(err)
				return
			}
			buf := make([]byte, 65536)
			_ = conn.SetReadDeadline(time.Now().Add(time.Second))
			if _, err := conn.Read(buf); err != nil {
				t.Error(err)
			}
		}(byte(i))
	}
	wg.Wait()
	if calls.Load() != 16 {
		t.Fatalf("DoH call count = %d, want 16", calls.Load())
	}
}

func TestMalformedAndOversizedInputRejected(t *testing.T) {
	var calls atomic.Int32
	p, cleanup := newTestProxy(t, echoHandler(t, &calls), Options{MaxQuerySize: 512})
	defer cleanup()

	conn, err := net.Dial("udp", p.Status().UDPAddress)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if _, err := conn.Write([]byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Read(make([]byte, 65536)); err == nil {
		t.Fatal("malformed query received a response")
	}
	_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if _, err := conn.Write(bytes.Repeat([]byte{1}, 513)); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Read(make([]byte, 65536)); err == nil {
		t.Fatal("oversized query received a response")
	}
	if calls.Load() != 0 {
		t.Fatalf("DoH call count = %d, want 0", calls.Load())
	}
}

func TestDoHFailureAndTimeoutFailClosed(t *testing.T) {
	failedClient, failedClose := testDoHClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream failure", http.StatusBadGateway)
	}))
	p, err := New(Options{Address: "127.0.0.1:0", Client: failedClient})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	conn, err := net.Dial("udp", p.Status().UDPAddress)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, _ = conn.Write(testQuery(1))
	if _, err := conn.Read(make([]byte, 65536)); err == nil {
		t.Fatal("failed DoH request received plaintext fallback response")
	}
	_ = conn.Close()
	_ = p.Stop()
	failedClose()
	if p.Status().LastError == nil {
		t.Fatal("DoH failure was not exposed in status")
	}

	timedClient, timedClose := testDoHClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(200 * time.Millisecond):
		}
	}))
	timed, err := New(Options{Address: "127.0.0.1:0", Client: timedClient, QueryTimeout: 20 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if err := timed.Start(); err != nil {
		t.Fatal(err)
	}
	conn, err = net.Dial("udp", timed.Status().UDPAddress)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	_, _ = conn.Write(testQuery(2))
	if _, err := conn.Read(make([]byte, 65536)); err == nil {
		t.Fatal("timed out DoH request received a response")
	}
	_ = conn.Close()
	_ = timed.Stop()
	timedClose()
}

func TestLoopbackBindingAndRestart(t *testing.T) {
	client, closeServer := testDoHClient(t, echoHandler(t, new(atomic.Int32)))
	defer closeServer()
	if _, err := New(Options{Address: "0.0.0.0:0", Client: client}); err == nil {
		t.Fatal("wildcard address was accepted")
	}
	if _, err := New(Options{Address: "localhost:0", Client: client}); err == nil {
		t.Fatal("hostname address was accepted")
	}
	p, err := New(Options{Address: "127.0.0.1:0", Client: client})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(p.Status().UDPAddress, "127.0.0.1:") || !strings.HasPrefix(p.Status().TCPAddress, "127.0.0.1:") {
		t.Fatalf("proxy did not bind loopback addresses: %+v", p.Status())
	}
	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := p.Stop(); err != ErrNotRunning {
		t.Fatalf("second stop error = %v, want ErrNotRunning", err)
	}
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateEndpointRequiresHTTPS(t *testing.T) {
	if err := ValidateEndpoint("http://dns.example/"); err == nil {
		t.Fatal("plaintext endpoint was accepted")
	}
	if err := ValidateEndpoint("https://dns.example/dns-query"); err != nil {
		t.Fatal(err)
	}
}
