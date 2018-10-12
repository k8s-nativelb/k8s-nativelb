/*
Copyright 2018 Sebastian Sch.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package proto

import (
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
)

func ConvertFarmToCommand(farm *v1.Farm) *Command {
	convertServers := make([]*Server, len(farm.Spec.Servers))

	idx := 0
	for _, server := range farm.Spec.Servers {
		backends := make([]*BackendSpec, len(server.Discovery.Backends))

		for idx, backend := range server.Discovery.Backends {
			backends[idx] = &BackendSpec{Port: backend.Port, Host: backend.Host, Priority: int32(backend.Priority), Weight: int32(backend.Weight)}
		}

		discovery := &Discovery{Kind: server.Discovery.Kind, BackendSpec: backends}

		healthCheck := &HealthCheck{Kind: server.HealthCheck.Kind,
			Fails:    int32(server.HealthCheck.Fails),
			Interval: server.HealthCheck.Interval, Passes: int32(server.HealthCheck.Passes),
			PingTimeoutDuration: server.HealthCheck.PingTimeoutDuration,
			Timeout:             server.HealthCheck.Timeout}

		udp := &UDP{MaxRequests: int32(server.UDP.MaxRequests), MaxResponses: int32(server.UDP.MaxResponses)}

		converServer := &Server{BackendConnectionTimeout: server.BackendConnectionTimeout,
			BackendIdleTimeout: server.BackendIdleTimeout,
			Balance:            server.Balance,
			Bind:               server.Bind,
			ClientIdleTimeout:  server.ClientIdleTimeout,
			MaxConnections:     int32(server.MaxConnections),
			Protocol:           server.Protocol,
			Discovery:          discovery,
			HealthCheck:        healthCheck,
			UDP:                udp}

		convertServers[idx] = converServer
		idx++
	}

	return &Command{Command: "", Servers: convertServers}
}
