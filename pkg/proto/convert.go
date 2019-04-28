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

func ConvertFarmsToGrpcDataList(farms []v1.Farm, clusterObject *v1.Cluster, agentNumber, numOfAgents int) []*FarmSpec {
	dataList := make([]*FarmSpec, 0)

	for _, farm := range farms {
		if clusterObject.Status.AllocatedNamespaces == nil {
			return nil
		}

		routerID := clusterObject.Status.AllocatedNamespaces[farm.Spec.ServiceNamespace].RouterID
		data := ConvertFarmToGrpcData(&farm, routerID)

		data.KeepalivedState = "MASTER"
		if agentNumber != 1 {
			data.KeepalivedState = "BACKUP"
		}

		data.Priority = int32(agentNumber)
		if (int(routerID)%numOfAgents)+1 == agentNumber {
			data.Priority += 100
		}

		dataList = append(dataList, data)
	}

	return dataList
}

func ConvertFarmToGrpcData(farm *v1.Farm, routerID int32) *FarmSpec {
	convertServers := make(map[string]*Server)

	for serverName, server := range farm.Spec.Servers {
		backends := make(map[string]*BackendSpec)

		for backendName, backend := range server.Backends {
			backends[backendName] = &BackendSpec{Port: backend.Port, Host: backend.Host, Priority: int32(backend.Priority), Weight: int32(backend.Weight)}
		}

		healthCheck := &HealthCheck{Kind: server.HealthCheck.Kind,
			Fails:    int32(server.HealthCheck.Fails),
			Interval: server.HealthCheck.Interval, Passes: int32(server.HealthCheck.Passes),
			PingTimeoutDuration: server.HealthCheck.PingTimeoutDuration,
			Timeout:             server.HealthCheck.Timeout}

		udp := &UDP{}
		if server.Protocol == "udp" {
			udp.MaxRequests = server.UDP.MaxRequests
			udp.MaxResponses = server.UDP.MaxResponses

		}

		tcp := &TCP{}
		if server.Protocol == "tcp" {
			tcp.BackendConnectionTimeout = server.TCP.BackendConnectionTimeout
			tcp.BackendIdleTimeout = server.TCP.BackendIdleTimeout
			tcp.ClientIdleTimeout = server.TCP.ClientIdleTimeout
			tcp.MaxConnections = server.TCP.MaxConnections
		}

		convertServer := &Server{
			Port:             server.Port,
			Balance:          server.Balance,
			Bind:             server.Bind,
			Protocol:         server.Protocol,
			Backends:         backends,
			HealthCheck:      healthCheck,
			UDP:              udp,
			TCP:              tcp,
			ClusterNamespace: farm.Namespace,
			ClusterName:      serverName}

		convertServers[serverName] = convertServer
	}

	return &FarmSpec{FarmName: farm.Name, Namespace: farm.Spec.ServiceNamespace, RouterID: routerID, Servers: convertServers}
}

func ConvertAgentStatusProtoToK8sAgent(agent *AgentStatus) v1.AgentStatus {
	a := v1.AgentStatus{}
	a.ConnectionStatus = v1.AgentAliveStatus
	a.KeepAlivedPid = agent.KeepAlivedPid
	a.HaproxyPid = agent.HaproxyPid
	a.NginxPid = agent.NginxPid
	a.OperationStatus = agent.OperationStatus

	haproxy := &v1.Haproxy{}
	if agent.HaproxyStatus != nil {
		haproxy = &v1.Haproxy{
			CompressBpsIn:              agent.HaproxyStatus.CompressBpsIn,
			CompressBpsOut:             agent.HaproxyStatus.CompressBpsOut,
			CompressBpsRateLim:         agent.HaproxyStatus.CompressBpsRateLim,
			ConnRate:                   agent.HaproxyStatus.ConnRate,
			ConnRateLimit:              agent.HaproxyStatus.ConnRateLimit,
			CumConns:                   agent.HaproxyStatus.CumConns,
			CumReq:                     agent.HaproxyStatus.CumReq,
			CumSslConns:                agent.HaproxyStatus.CumSslConns,
			CurrConns:                  agent.HaproxyStatus.CurrConns,
			CurrSslConns:               agent.HaproxyStatus.CurrSslConns,
			HardMaxconn:                agent.HaproxyStatus.HardMaxconn,
			IdlePct:                    agent.HaproxyStatus.IdlePct,
			Maxconn:                    agent.HaproxyStatus.Maxconn,
			MaxConnRate:                agent.HaproxyStatus.MaxConnRate,
			Maxpipes:                   agent.HaproxyStatus.Maxpipes,
			MaxSessRate:                agent.HaproxyStatus.MaxSessRate,
			Maxsock:                    agent.HaproxyStatus.Maxsock,
			MaxSslConns:                agent.HaproxyStatus.MaxSslConns,
			MaxSslRate:                 agent.HaproxyStatus.MaxSslRate,
			MemMaxMB:                   agent.HaproxyStatus.MemMaxMB,
			Nbproc:                     agent.HaproxyStatus.Nbproc,
			Pid:                        agent.HaproxyStatus.Pid,
			PipesFree:                  agent.HaproxyStatus.PipesFree,
			PipesUsed:                  agent.HaproxyStatus.PipesUsed,
			ProcessNum:                 agent.HaproxyStatus.ProcessNum,
			ReleaseDate:                agent.HaproxyStatus.ReleaseDate,
			RunQueue:                   agent.HaproxyStatus.RunQueue,
			SessRate:                   agent.HaproxyStatus.SessRate,
			SessRateLimit:              agent.HaproxyStatus.SessRateLimit,
			SslBackendKeyRate:          agent.HaproxyStatus.SslBackendKeyRate,
			SslBackendMaxKeyRate:       agent.HaproxyStatus.SslBackendMaxKeyRate,
			SslCacheLookups:            agent.HaproxyStatus.SslCacheLookups,
			SslCacheMisses:             agent.HaproxyStatus.SslCacheMisses,
			SslFrontendKeyRate:         agent.HaproxyStatus.SslFrontendKeyRate,
			SslFrontendMaxKeyRate:      agent.HaproxyStatus.SslFrontendMaxKeyRate,
			SslFrontendSessionReusePct: agent.HaproxyStatus.SslFrontendSessionReusePct,
			SslRate:                    agent.HaproxyStatus.SslRate,
			SslRateLimit:               agent.HaproxyStatus.SslRateLimit,
			Tasks:                      agent.HaproxyStatus.Tasks,
			UlimitN:                    agent.HaproxyStatus.UlimitN,
			Uptime:                     agent.HaproxyStatus.Uptime,
			UptimeSec:                  agent.HaproxyStatus.UptimeSec,
			Version:                    agent.HaproxyStatus.Version,
		}
	}

	nginx := &v1.Nginx{}
	if agent.NginxStatus != nil {
		nginx = &v1.Nginx{
			Pid:               agent.NginxStatus.Pid,
			Version:           agent.NginxStatus.Version,
			ActiveConnections: agent.NginxStatus.ActiveConnections,
			Reading:           agent.NginxStatus.Reading,
			Waiting:           agent.NginxStatus.Waiting,
			Writing:           agent.NginxStatus.Writing,
		}
	}

	if agent.KeepalivedState == nil {
		agent.KeepalivedState = make(map[string]string)
	}

	a.LoadBalancer = &v1.LoadBalancer{Haproxy: haproxy, Nginx: nginx, Keepalived: &v1.Keepalived{Pid: uint64(agent.KeepAlivedPid), InstancesStatus: agent.KeepalivedState}}
	return a
}

func ConvertStatusProtoToK8sHaproxyStatus(stat *Status) *v1.HaproxyStatus {
	return &v1.HaproxyStatus{
		Status:   stat.Status,
		PxName:   stat.PxName,
		Pid:      stat.Pid,
		Type:     stat.Type,
		Act:      stat.Act,
		Bck:      stat.Bck,
		Bin:      stat.Bin,
		Bout:     stat.Bout,
		ChkDown:  stat.ChkDown,
		ChkFail:  stat.ChkFail,
		Downtime: stat.Downtime,
		Dreq:     stat.Dreq,
		Dresp:    stat.Dresp,
		Econ:     stat.Econ,
		Ereq:     stat.Ereq,
		Eresp:    stat.Eresp,
		Iid:      stat.Iid,
		Lastchg:  stat.Lastchg,
		Lbtot:    stat.Lbtot,
		Qcur:     stat.Qcur,
		Qlimit:   stat.Qlimit,
		Qmax:     stat.Qmax,
		Rate:     stat.Rate,
		RateLim:  stat.RateLim,
		RateMax:  stat.RateMax,
		Scur:     stat.Scur,
		Sid:      stat.Sid,
		Slim:     stat.Slim,
		Smax:     stat.Smax,
		Stot:     stat.Stot,
		SvName:   stat.SvName,
		Throttle: stat.Throttle,
		Tracked:  stat.Tracked,
		Weight:   stat.Weight,
		Wredis:   stat.Wredis,
		Wretr:    stat.Wretr,
	}
}
