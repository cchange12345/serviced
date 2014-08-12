// Copyright 2014 The Serviced Authors.
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
package scheduler

import (
	"path"

	"github.com/control-center/serviced/coordinator/client"
	"github.com/control-center/serviced/domain/pool"
	"github.com/control-center/serviced/zzk"
)

const (
	zkPool = "/pools"
)

func poolpath(nodes ...string) string {
	p := append([]string{zkPool}, nodes...)
	return path.Join(p...)
}

type PoolNode struct {
	*pool.ResourcePool
	version interface{}
}

// ID implements zzk.Node
func (node *PoolNode) GetID() string {
	return node.ID
}

// Create implements zzk.Node
func (node *PoolNode) Create(conn client.Connection) error {
	return AddResourcePool(conn, node.ID)
}

// Update implements zzk.Node
func (node *PoolNode) Update(conn client.Connection) error {
	return nil
}

// Delete implements zzk.Node
func (node *PoolNode) Delete(conn client.Connection) error {
	return RemoveResourcePool(conn, node.ID)
}

func (node *PoolNode) Version() interface{}           { return node.version }
func (node *PoolNode) SetVersion(version interface{}) { node.version = version }

func SyncResourcePools(conn client.Connection, pools []*pool.ResourcePool) error {
	nodes := make([]zzk.Node, len(pools))
	for i := range pools {
		nodes[i] = &PoolNode{ResourcePool: pools[i]}
	}
	return zzk.Sync(conn, nodes, poolpath())
}

func AddResourcePool(conn client.Connection, poolID string) error {
	return conn.CreateDir(poolpath(poolID))
}

func RemoveResourcePool(conn client.Connection, poolID string) error {
	return conn.Delete(poolpath(poolID))
}