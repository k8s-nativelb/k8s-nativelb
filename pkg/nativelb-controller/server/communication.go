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
package server

import (
	"fmt"
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/proto"
)

func (n *NativeLBGrpcServer) CreateFarmOnCluster(farm *v1.Farm, cluster *v1.Cluster) error {
	command := proto.ConvertFarmToCommand(farm)

	command.Command = proto.AgentCreateCommand
	err := n.sendCommand(cluster.Name, command)

	return err
}

func (n *NativeLBGrpcServer) UpdateFarmOnCluster(farm *v1.Farm, cluster *v1.Cluster) error {
	command := proto.ConvertFarmToCommand(farm)

	command.Command = proto.AgentUpdateCommand
	err := n.sendCommand(cluster.Name, command)

	return err
}

func (n *NativeLBGrpcServer) DeleteFarmOnCluster(farm *v1.Farm, cluster *v1.Cluster) error {
	command := proto.ConvertFarmToCommand(farm)

	command.Command = proto.AgentDeleteCommand
	err := n.sendCommand(cluster.Name, command)

	return err
}

func (n *NativeLBGrpcServer) sendCommand(clusterName string, command *proto.Command) error {
	isAnyAgentAlive := false

	for _, agent := range n.Cluster[clusterName] {
		agentInstance, err := n.nativelbClient.Agent().Get(agent.Data.HostName)
		if err != nil {
			log.Log.Errorf("Fail to get agent %s object error: %v", agent.Data.HostName, err)
			continue
		}

		err = agent.connection.Send(command)
		if err != nil {
			agentInstance.Status.ConnectionStatus = v1.AgentDownStatus
			log.Log.Errorf("Fail to send command to agent %s error: %v", agent.Data.HostName, err)
		} else {
			agentInstance.Status.ConnectionStatus = v1.AgentAliveStatus
			isAnyAgentAlive = true
		}

		_, err = n.nativelbClient.Agent().Update(agentInstance)
		if err != nil {
			log.Log.Errorf("Fail to update agent %s object error: %v", agent.Data.HostName, err)
		}
	}

	if !isAnyAgentAlive {
		return fmt.Errorf("Not agent alive on cluster %s", clusterName)
	}

	return nil
}
