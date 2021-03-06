// Copyright 2015 The Serviced Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/control-center/serviced/cli/api"
	"github.com/control-center/serviced/config"
	"github.com/control-center/serviced/rpc/rpcutils"
)

// Initializer for serviced server
func (c *ServicedCli) initServer() {
	c.app.Commands = append(c.app.Commands, cli.Command{
		Name:        "server",
		Usage:       "Starts serviced",
		Description: "serviced server",
		Action:      c.cmdServer,
	})
}

// serviced server
func (c *ServicedCli) cmdServer(ctx *cli.Context) {
	opts := config.GetOptions()
	if err := api.ValidateServerOptions(&opts); err != nil {
		fmt.Printf("Unable to validate server options: %s", err)
		os.Exit(1)
	}

	config.LoadOptions(opts)

	// Start server mode
	rpcutils.RPC_CLIENT_SIZE = api.GetOptionsMaxRPCClients()
	if err := c.driver.StartServer(); err != nil {
		log.WithError(err).Fatal("Unable to start server")
	}
}
