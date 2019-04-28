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

package nativelb_agent

import (
	"context"
	"fmt"

	"github.com/k8s-nativelb/pkg/log"
	. "github.com/k8s-nativelb/pkg/proto"
)

func (n *NativelbAgent) CreateServers(ctx context.Context, data *FarmSpec) (*Result, error) {
	log.Log.Infof("CreateServers grpc call with data: %v", *data)
	if n.agentStatus.Status != AgentSyncedStatus {
		log.Log.Infof("failed to create servers for farm %s because the agent is not in synced status", data.FarmName)
		return nil, fmt.Errorf("failed to create servers for farm %s because the agent is not in synced status", data.FarmName)
	}

	err := n.UpdateAndReload(data)
	if err != nil {
		return &Result{}, err
	}

	return &Result{}, nil
}
func (n *NativelbAgent) UpdateServers(ctx context.Context, data *FarmSpec) (*Result, error) {
	log.Log.Infof("UpdateServers grpc call with data: %v", *data)
	if n.agentStatus.Status != AgentSyncedStatus {
		log.Log.Infof("failed to update servers for farm %s because the agent is not in synced status", data.FarmName)
		return nil, fmt.Errorf("failed to update servers for farm %s because the agent is not in synced status", data.FarmName)
	}

	err := n.UpdateAndReload(data)
	if err != nil {
		return nil, err
	}

	return &Result{}, nil
}
func (n *NativelbAgent) DeleteServers(ctx context.Context, data *FarmSpec) (*Result, error) {
	log.Log.Infof("DeleteServers grpc call with data: %v", *data)
	if n.agentStatus.Status != AgentSyncedStatus {
		log.Log.Infof("failed to delete servers for farm %s because the agent is not in synced status", data.FarmName)
		return nil, fmt.Errorf("failed to delete servers for farm %s because the agent is not in synced status", data.FarmName)
	}

	err := n.DeleteAndReload(data)
	if err != nil {
		return nil, err
	}

	return &Result{}, nil
}
func (n *NativelbAgent) GetAgentStatus(ctx context.Context, cmd *Command) (*AgentStatus, error) {
	log.Log.Infof("GetAgentStatus grpc call with command: %v", cmd)

	if n.agentStatus.Status != AgentNewStatus {
		haproxy, err := n.loadBalancerController.GetStatus()
		if err != nil {
			return nil, err
		}
		n.agentStatus.HaproxyStatus = haproxy
		n.agentStatus.HaproxyPid = int32(haproxy.Pid)

		nginx, err := n.udpLoadBalancerController.GetStatus()
		if err != nil {
			return nil, err
		}
		n.agentStatus.NginxStatus = nginx

		n.agentStatus.KeepAlivedPid = n.keepalivedController.GetPid()
		n.agentStatus.KeepalivedState, err = n.keepalivedController.GetStatus()
	}

	log.Log.Infof("agent status data: %v", *n.agentStatus)
	return n.agentStatus, nil
}

func (n *NativelbAgent) InitAgent(ctx context.Context, data *InitAgentData) (*InitAgentResult, error) {
	log.Log.Infof("InitAgent grpc call with initData: %v", *data)
	if data.Operational {
		err := n.LoadInitToEngines(data)
		if err != nil {
			return nil, err
		}

		haproxy, err := n.loadBalancerController.GetStatus()
		if err != nil {
			return nil, err
		}
		n.agentStatus.HaproxyStatus = haproxy
		n.agentStatus.HaproxyPid = int32(haproxy.Pid)

		nginx, err := n.udpLoadBalancerController.GetStatus()
		if err != nil {
			return nil, err
		}
		n.agentStatus.NginxStatus = nginx
		n.agentStatus.KeepAlivedPid = n.keepalivedController.GetPid()

		n.agentStatus.KeepalivedState, err = n.keepalivedController.GetStatus()
	}

	n.agentStatus.SyncVersion = data.SyncVersion
	n.agentStatus.Status = AgentSyncedStatus

	return &InitAgentResult{Agent: n.agent, AgentStatus: n.agentStatus}, nil
}

func (n *NativelbAgent) GetServersStats(ctx context.Context, cmd *Command) (*ServersStats, error) {
	log.Log.Infof("GetFarmsStats grpc call with command: %v", cmd)
	if n.agentStatus.Status != AgentSyncedStatus {
		return nil, fmt.Errorf("failed to get server stats because the agent is not in synced status")
	}

	serversStatsMap, err := n.loadBalancerController.GetStats()
	if err != nil {
		return nil, err
	}

	return serversStatsMap, nil
}

func (n *NativelbAgent) UpdateAgentSyncVersion(ctx context.Context, data *InitAgentData) (*Result, error) {
	log.Log.Infof("get UpdateAgentSyncVersion status grpc call with initData: %v", *data)
	n.agentStatus.SyncVersion = data.SyncVersion
	return &Result{}, nil
}

func (n *NativelbAgent) StopAgent(ctx context.Context, cmd *Command) (*Result, error) {
	return &Result{}, nil
}
