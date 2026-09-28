/* SPDX-License-Identifier: MIT */

package doh

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

type tlsEOFConn struct{}

func (tlsEOFConn) Read([]byte) (int, error)         { return 0, io.EOF }
func (tlsEOFConn) Write(p []byte) (int, error)      { return len(p), nil }
func (tlsEOFConn) Close() error                     { return nil }
func (tlsEOFConn) LocalAddr() net.Addr              { return &net.TCPAddr{} }
func (tlsEOFConn) RemoteAddr() net.Addr             { return &net.TCPAddr{} }
func (tlsEOFConn) SetDeadline(time.Time) error      { return nil }
func (tlsEOFConn) SetReadDeadline(time.Time) error  { return nil }
func (tlsEOFConn) SetWriteDeadline(time.Time) error { return nil }

// TestTLSClientHelloDoesNotPanic exercises crypto/tls far enough to construct
// a ClientHello. The injected EOF is expected after the ClientHello write.
// CI runs this test with the production Go overlay enabled.
func TestTLSClientHelloDoesNotPanic(t *testing.T) {
	err := tls.Client(tlsEOFConn{}, &tls.Config{ServerName: "example.com"}).Handshake()
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF after ClientHello, got %v", err)
	}
}
