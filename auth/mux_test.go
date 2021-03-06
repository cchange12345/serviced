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

// +build unit

package auth_test

import (
	"bytes"
	"time"

	"github.com/control-center/serviced/auth"
	. "gopkg.in/check.v1"
)

func (s *TestAuthSuite) TestBuildHeaderBadAddr(c *C) {
	token, _, _ := auth.CreateJWTIdentity(s.hostId, s.poolId, s.admin, s.dfs, s.delegatePubPEM, time.Hour)
	addr := "this is more than 6 bytes"
	var b bytes.Buffer
	err := auth.AddSignedMuxHeader(&b, []byte(addr), token)
	c.Assert(err, Equals, auth.ErrBadMuxAddress)
}

func (s *TestAuthSuite) TestExtractBadHeader(c *C) {
	mockHeader := []byte{0, 0, 0, 19, 109, 121, 32, 115, 117, 112, 101, 114, 32, 102}
	b := bytes.NewBuffer(mockHeader)
	_, _, err := auth.ReadMuxHeader(b)
	c.Assert(err, Not(IsNil))
}

func (s *TestAuthSuite) TestBuildAndExtractHeader(c *C) {
	token, _, _ := auth.CreateJWTIdentity(s.hostId, s.poolId, s.admin, s.dfs, s.delegatePubPEM, time.Hour)
	addr := "zenoss" // Not valid, but it is 6 bytes!
	var b bytes.Buffer

	// build header
	err := auth.AddSignedMuxHeader(&b, []byte(addr), token)
	c.Assert(err, Equals, nil)

	// extract header
	extractedAddr, ident, err := auth.ReadMuxHeader(&b)
	// check the address is correctly decoded
	c.Assert(err, IsNil)
	c.Assert(string(extractedAddr), DeepEquals, addr)
	// check the identity has been correctly extracted
	c.Assert(s.hostId, DeepEquals, ident.HostID())
	c.Assert(s.poolId, DeepEquals, ident.PoolID())
	c.Assert(s.admin, Equals, ident.HasAdminAccess())
	c.Assert(s.dfs, Equals, ident.HasDFSAccess())
}
