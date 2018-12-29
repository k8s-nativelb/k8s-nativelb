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

func ConvertFarmsToGrpcDataList(farms []v1.Farm, clusterObject *v1.Cluster, keepaliveState string, priority int32) []*Data {
	dataList := make([]*Data, 0)

	for _, farm := range farms {
		data := ConvertFarmToGrpcData(&farm, clusterObject.Status.AllocatedNamespaces[farm.Spec.ServiceNamespace].RouterID)
		data.KeepalivedState = keepaliveState
		data.Priority = priority
		dataList = append(dataList, data)
	}

	return dataList
}

func ConvertFarmToGrpcData(farm *v1.Farm, routerID int32) *Data {
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
			Port:               server.Port,
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

	return &Data{FarmName: farm.Name, Namespace: farm.Spec.ServiceNamespace, RouterID: routerID, Servers: convertServers}
}

func ConvertAgentProtoToK8sAgent(agent *Agent) *v1.Agent {
	a := &v1.Agent{}
	a.Namespace = v1.ControllerNamespace
	a.Name = agent.HostName
	a.Spec = v1.AgentSpec{HostName: agent.HostName, IPAddress: agent.IPAddress, Cluster: agent.Cluster}
	a.Status = v1.AgentStatus{ConnectionStatus: v1.AgentAliveStatus}

	return a
}
