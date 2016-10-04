// Copyright 2016 The Serviced Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mux

import (
	"bytes"
	"io"

	"github.com/control-center/serviced/utils"
)

// AddrLen is the size of the address in the mux header
const AddrLen = 6

// Addr is the local address that the mux dials
type Addr [AddrLen]byte

// Resolve resolves the ip:port and returns the address
func Resolve(addr string) (*Addr, error) {
	packedAddr, err := utils.PackTCPAddressString(addr)
	if err != nil {
		return nil, err
	}
	return &Addr(packedAddr[:AddrLen]), nil
}

// Network returns the network
func (a Addr) Network() string {
	return "ip"
}

func (a Addr) String() string {
	return utils.UnpackTCPAddressToString(a[:])
}

// Header contains information pertaining to the header of the mux
type Header struct {
	addr *Addr
}

// NewHeader initializes the raw header packet
func NewHeader(addr string) (*Header, error) {
	addr, err := Resolve(addr)
	if err != nil {
		return nil, err
	}

	return &Header{addr: addr}
}

// ReadFrom loads the header information from the reader
func (h *Header) ReadFrom(r io.Reader) (n int64, err error) {
	var addr Addr
	n, err = io.ReadFull(r, addr[:])
	if err != nil {
		return
	}
	h.addr = &addr
	return
}

// WriteTo writes the header information to the writer
func (h Header) WriteTo(w io.Writer) (n int64, err error) {
	return w.Write(h.addr[:])
}

// Address returns the address to connect
func (h Header) Address() string {
	return h.addr.String()
}

// Verifier verifies the validity of a signature after loading it from the
// reader.
type Verifier interface {
	Verify(r io.Reader, m []byte) (n int64, err error)
	ReadToken(r io.Reader) (n int64, err error)
}

// ReadHeader extracts the mux header from the stream and authenticates.
func ReadHeader(r io.Reader, v Verifier) (h *Header, n int64, err error) {

	// if verifier is nil, just return the header
	if v == nil {
		h = &Header{}
		n, err = io.Copy(h, r)
		return
	}

	// set up a buffer to tee the signed message
	msg := &bytes.Buffer{}
	tee := io.TeeReader(r, msg)

	// read the token
	nsize, err := v.ReadToken(tee)
	n += nsize
	if err != nil {
		return nil, n, err
	}

	// read the header
	h = &Header{}
	nsize, err = io.Copy(h, tee)
	n += nsize
	if err != nil {
		return nil, n, err
	}

	// verify the signature (use the raw stream; not the tee)
	nsize, err = v.Verify(r, msg.Bytes())
	n += nsize
	if err != nil {
		return nil, n, err
	}

	return
}

// Signer generates the signature and can write the signature to a writer
type Signer interface {
	Sign(w io.Writer, m []byte) (n int64, err error)
	WriteToken(w io.Writer) (n int64, err error)
}

// WriteHeader dumps the mux header to the stream and signs.
func WriteHeader(w io.Writer, h *Header, s Signer) (n int64, err error) {

	// if signer is nil, just write the header
	if s == nil {
		return io.Copy(w, h)
	}

	// set up a buffer to multi
	buffer := &bytes.Buffer{}
	multi := io.MultiWriter(w, buffer)

	// write the token
	nsize, err := s.WriteToken(multi)
	n += nsize
	if err != nil {
		return
	}

	// write the header
	nsize, err = io.Copy(multi, h)
	n += nsize
	if err != nil {
		return
	}

	// sign the header (use the raw stream; not the multi)
	nsize, err = s.Sign(w, buffer.Bytes())
	n += nsize
	return
}