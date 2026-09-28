// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is the small Go 1.27 crypto/rand compatibility surface still required
// by crypto/rsa after TunnelMint excludes crypto/rand/util.go to avoid the
// FIPS DRBG dependency tree. It intentionally uses the supplied reader
// directly and does not import crypto/internal/rand or crypto/internal/fips140.

package rand

import (
	"errors"
	"io"
	"math/big"
)

// Prime returns a number of the given bit length that is prime with high probability.
func Prime(r io.Reader, bits int) (*big.Int, error) {
	if bits < 2 {
		return nil, errors.New("crypto/rand: prime size must be at least 2-bit")
	}
	b := uint(bits % 8)
	if b == 0 {
		b = 8
	}
	bytes := make([]byte, (bits+7)/8)
	p := new(big.Int)
	for {
		if _, err := io.ReadFull(r, bytes); err != nil {
			return nil, err
		}
		bytes[0] &= uint8(int(1<<b) - 1)
		if b >= 2 {
			bytes[0] |= 3 << (b - 2)
		} else {
			bytes[0] |= 1
			if len(bytes) > 1 {
				bytes[1] |= 0x80
			}
		}
		bytes[len(bytes)-1] |= 1
		p.SetBytes(bytes)
		if p.ProbablyPrime(20) {
			return p, nil
		}
	}
}

// Int returns a uniform random value in [0, max). It panics if max <= 0.
func Int(r io.Reader, max *big.Int) (n *big.Int, err error) {
	if max.Sign() <= 0 {
		panic("crypto/rand: argument to Int is <= 0")
	}
	n = new(big.Int)
	n.Sub(max, n.SetUint64(1))
	bitLen := n.BitLen()
	if bitLen == 0 {
		return
	}
	k := (bitLen + 7) / 8
	b := uint(bitLen % 8)
	if b == 0 {
		b = 8
	}
	bytes := make([]byte, k)
	for {
		_, err = io.ReadFull(r, bytes)
		if err != nil {
			return nil, err
		}
		bytes[0] &= uint8(int(1<<b) - 1)
		n.SetBytes(bytes)
		if n.Cmp(max) < 0 {
			return
		}
	}
}
