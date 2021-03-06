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

package master

import (
	"github.com/control-center/serviced/domain/service"
	"github.com/control-center/serviced/domain/servicedefinition"
)

// Adds a port public endpoint to a service.
func (c *Client) AddPublicEndpointPort(serviceid, endpointName, portAddr string, usetls bool,
	protocol string, isEnabled bool, restart bool) (*servicedefinition.Port, error) {
	request := &PublicEndpointRequest{
		Serviceid:    serviceid,
		EndpointName: endpointName,
		Name:         portAddr,
		UseTLS:       usetls,
		Protocol:     protocol,
		IsEnabled:    isEnabled,
		Restart:      restart,
	}
	var result servicedefinition.Port
	err := c.call("AddPublicEndpointPort", request, &result)
	return &result, err
}

// Remove a port public endpoint from a service.
func (c *Client) RemovePublicEndpointPort(serviceid, endpointName, portAddr string) error {
	request := &PublicEndpointRequest{
		Serviceid:    serviceid,
		EndpointName: endpointName,
		Name:         portAddr,
	}
	return c.call("RemovePublicEndpointPort", request, nil)
}

// Enable/disable a port public endpoint for a service.
func (c *Client) EnablePublicEndpointPort(serviceid, endpointName, portAddr string, isEnabled bool) error {
	request := &PublicEndpointRequest{
		Serviceid:    serviceid,
		EndpointName: endpointName,
		Name:         portAddr,
		IsEnabled:    isEnabled,
	}
	return c.call("EnablePublicEndpointPort", request, nil)
}

// Adds a vhost public endpoint to a service.
func (c *Client) AddPublicEndpointVHost(serviceid, endpointName, vhost string, isEnabled,
	restart bool) (*servicedefinition.VHost, error) {
	request := &PublicEndpointRequest{
		Serviceid:    serviceid,
		EndpointName: endpointName,
		Name:         vhost,
		IsEnabled:    isEnabled,
		Restart:      restart,
	}
	var result servicedefinition.VHost
	err := c.call("AddPublicEndpointVHost", request, &result)
	return &result, err
}

// Remove a vhost public endpoint from a service.
func (c *Client) RemovePublicEndpointVHost(serviceid, endpointName, vhost string) error {
	request := &PublicEndpointRequest{
		Serviceid:    serviceid,
		EndpointName: endpointName,
		Name:         vhost,
	}
	return c.call("RemovePublicEndpointVHost", request, nil)
}

// Enable/disable a vhost public endpoint for a service.
func (c *Client) EnablePublicEndpointVHost(serviceid, endpointName, vhost string, isEnabled bool) error {
	request := &PublicEndpointRequest{
		Serviceid:    serviceid,
		EndpointName: endpointName,
		Name:         vhost,
		IsEnabled:    isEnabled,
	}
	return c.call("EnablePublicEndpointVHost", request, nil)
}

// GetAllPublicEndpoints
func (c *Client) GetAllPublicEndpoints() ([]service.PublicEndpoint, error) {
	var response []service.PublicEndpoint
	if err := c.call("GetAllPublicEndpoints", empty, &response); err != nil {
		return response, err
	}
	return response, nil
}
