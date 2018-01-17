// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Inspiration and code have been taken from
// https://golang.org/src/encoding/json/stream.go

package bencode

import (
	"io"
)

// A Decoder reads and decodes bencode values from an input stream.
type Decoder struct {
	r       io.Reader
	buf     []byte
	d       decodeState
	err     error
	minRead int
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may
// read data from r beyond the JSON values requested.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r, minRead: 512}
}

// Decode -
func (dec *Decoder) Decode(v interface{}) (err error) {
	if dec.err != nil {
		return dec.err
	}
	if dec.d.lastReadValueOff > 0 { // Clean buffer from what already have been parsed
		n := copy(dec.buf, dec.buf[dec.d.lastReadValueOff:])
		dec.buf = dec.buf[:n]
	}
	for {
		if cap(dec.buf)-len(dec.buf) < dec.minRead {
			dec.extendBuffer()
		}
		n, err := dec.r.Read(dec.buf[len(dec.buf):cap(dec.buf)])
		dec.buf = dec.buf[0 : len(dec.buf)+n]
		if err != nil {
			break
		}
	}
	dec.d.init(dec.buf)
	return dec.d.unmarshal(v)
}

func (dec *Decoder) extendBuffer() {
	newBuf := make([]byte, len(dec.buf), 2*cap(dec.buf)+dec.minRead)
	copy(newBuf, dec.buf)
	dec.buf = newBuf
}
