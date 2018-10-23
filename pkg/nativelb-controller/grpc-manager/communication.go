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
package grpc_manager

import (
	"fmt"
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func (n *NativeLBGrpcManager) CreateFarmOnCluster(farm *v1.Farm, cluster *v1.Cluster) error {
	return n.sendDataToAgent("create", farm, cluster)
}

func (n *NativeLBGrpcManager) UpdateFarmOnCluster(farm *v1.Farm, cluster *v1.Cluster) error {
	return n.sendDataToAgent("update", farm, cluster)
}

func (n *NativeLBGrpcManager) DeleteFarmOnCluster(farm *v1.Farm, cluster *v1.Cluster) error {
	return n.sendDataToAgent("delete", farm, cluster)
}

func (n *NativeLBGrpcManager) keepalive() {
	agents, err := n.nativelbClient.Agent().List(&client.ListOptions{})
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get agent list for keepalive check")
	}
	for _, agent := range agents.Items {
		n.getAgentStatus(agent)

	}
}

func (n *NativeLBGrpcManager) getAgentStatus(agent v1.Agent) {
	conn, err := n.connect(agent.GetUrl())
	if err != nil {
		log.Log.Reason(err).Errorf("failed to connect to agent %s url %s error: %v",agent.Name,agent.GetUrl(),err)
		agent.Status.ConnectionStatus = v1.AgentDownStatus
		err = n.updateAgentStatus(agent)
		if err != nil {
			log.Log.Reason(err).Errorf("failed to update agent %s to down status error: %v", agent.Name, err)
		}
		return
	}
	defer conn.Close()

	grpcClient := proto.NewNativeLoadBalancerAgentClient(conn)
	agentStatus, err := grpcClient.GetAgentStatus(context.TODO(), &proto.Command{})
	if err != nil {
		log.Log.Reason(err).Errorf("failed to create grpc client to agent %s url %s error: %v",agent.Name,agent.GetUrl(),err)
		agent.Status.ConnectionStatus = v1.AgentDownStatus
		err = n.updateAgentStatus(agent)
		if err != nil {
			log.Log.Reason(err).Errorf("failed to update agent %s to down status error: %v", agent.Name, err)
		}
		return
	}

	//TODO: parse agent status to k8s object status
	fmt.Println(*agentStatus)
	agent.Status.ConnectionStatus = v1.AgentAliveStatus
	agent.Status.Pid = int(agentStatus.Pid)
	agent.Status.StartTime = agentStatus.StartTime
	agent.Status.Time = agentStatus.Time
	agent.Status.Uptime = time.Duration(agentStatus.Uptime.Seconds)
	agent.Status.Version = agentStatus.Version
	err = n.updateAgentStatus(agent)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to update agent %s status error: %v", agent.Name, err)
	}
}

func (n *NativeLBGrpcManager) connect(serverAddr string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(serverAddr, []grpc.DialOption{grpc.WithInsecure()}...)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to dial to %s error: %v", serverAddr, err)
		return nil, err
	}
	return conn, nil
}

func (n *NativeLBGrpcManager) sendDataToAgent(command string, farm *v1.Farm, cluster *v1.Cluster) error {
	data := proto.ConvertFarmToGrpcData(farm)
	labelSelector := labels.Set{}
	labelSelector[v1.ClusterLabel] = cluster.Name
	agents, err := n.nativelbClient.Agent().List(&client.ListOptions{LabelSelector: labelSelector.AsSelector()})
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get agents list for cluster %s error: %v", cluster.Name, err)
		return err
	}

	isAnyAgentAlive := false
	for _, agentInstance := range agents.Items {
		conn, err := n.connect(agentInstance.GetUrl())
		if err != nil {
			agentInstance.Status.ConnectionStatus = v1.AgentDownStatus
			err = n.updateAgentStatus(agentInstance)
			if err != nil {
				log.Log.Reason(err).Errorf("failed to update agent %s to down status error: %v", agentInstance.Name, err)
			}
			continue
		}
		defer conn.Close()

		grpcClient := proto.NewNativeLoadBalancerAgentClient(conn)
		switch command {
		case "create":
			_, err = grpcClient.CreateServers(context.TODO(), data)
		case "update":
			_, err = grpcClient.UpdateServers(context.TODO(), data)
		case "delete":
			_, err = grpcClient.DeleteServers(context.TODO(), data)
		}

		if err != nil {
			agentInstance.Status.ConnectionStatus = v1.AgentDownStatus
			continue
		}
		isAnyAgentAlive = true
	}

	if !isAnyAgentAlive {
		return fmt.Errorf("failed to find any agent alive for cluster %s", cluster.Name)
	}

	return nil
}

func (n *NativeLBGrpcManager) updateAgentStatus(agent v1.Agent) error {
	n.updateAgentStatusMutex.Lock()
	_, err := n.nativelbClient.Agent().Update(&agent)
	n.updateAgentStatusMutex.Unlock()
	return err
}
