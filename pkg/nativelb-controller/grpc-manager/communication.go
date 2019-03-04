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
	"strconv"
	"time"

	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (n *NativeLBGrpcManager) GetAgentStatus(agent *v1.Agent, agentNumber, numOfAgents int) {
	conn, err := n.connect(agent.GetUrl())
	if err != nil {
		log.Log.Reason(err).Errorf("failed to connect to agent %s url %s error: %v", agent.Name, agent.GetUrl(), err)
		agent.Status.ConnectionStatus = v1.AgentDownStatus
		_, err = n.updateAgentStatus(agent)
		if err != nil {
			log.Log.Reason(err).Errorf("failed to update agent %s to down status error: %v", agent.Name, err)
		}
		return
	}
	defer conn.Close()

	grpcClient := proto.NewNativeLoadBalancerAgentClient(conn)
	agentStatus, err := grpcClient.GetAgentStatus(context.TODO(), &proto.Command{})
	if err != nil {
		log.Log.Reason(err).Errorf("failed to create grpc client to agent %s url %s error: %v", agent.Name, agent.GetUrl(), err)
		agent.Status.ConnectionStatus = v1.AgentDownStatus
		_, err = n.updateAgentStatus(agent)
		if err != nil {
			log.Log.Reason(err).Errorf("failed to update agent %s to down status error: %v", agent.Name, err)
		}
		return
	}

	agentData := &proto.Agent{}
	if agentStatus.Status == proto.AgentNewStatus {
		log.Log.Infof("new agent %s status received sending initAgent", agent.Name)
		agentData, agentStatus, err = n.InitAgent(&grpcClient, agent, agentNumber, numOfAgents)
		if err != nil {
			log.Log.Reason(err).Errorf("failed to initAgent for agent %s with error %v", agent.Name, err)
			return
		}
	}

	// Resync new agent
	intResourceVersion, err := strconv.Atoi(agent.ResourceVersion)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to convert resourceVersion from agent %s value %s error %v", agent.Name, agent.ResourceVersion, err)
		agent.Status.ConnectionStatus = v1.AgentDownStatus
		_, err = n.updateAgentStatus(agent)
		if err != nil {
			log.Log.Reason(err).Errorf("failed to update agent %s to down status error: %v", agent.Name, err)
		}
		return
	}

	// Resync agent out of sync
	if agentStatus.SyncVersion != int32(intResourceVersion) {
		log.Log.Infof("agent %s is out of sync sending initAgent", agent.Name)
		agentData, agentStatus, err = n.InitAgent(&grpcClient, agent, agentNumber, numOfAgents)
		if err != nil {
			log.Log.Reason(err).Errorf("failed to initAgent for agent %s with error %v", agent.Name, err)
			return
		}
	}

	if int(agentStatus.KeepAlivedPid) == 0 || int(agentStatus.LBPid) == 0 {
		log.Log.Errorf("get bad response from agent %s, pid can't be 0", agent.Name)
		agent.Status.ConnectionStatus = v1.AgentDownStatus
		_, err = n.updateAgentStatus(agent)
		if err != nil {
			log.Log.Reason(err).Errorf("failed to update agent %s to down status error: %v", agent.Name, err)
		}
		return
	}

	// TODO Add more specs here
	agent.Spec.HostName = agentData.HostName

	agent.Status.ConnectionStatus = v1.AgentAliveStatus
	agent.Status.LBPid = int(agentStatus.LBPid)
	agent.Status.KeepAlivedPid = int(agentStatus.KeepAlivedPid)
	agent.Status.StartTime = agentStatus.StartTime
	agent.Status.Time = agentStatus.Time
	agent.Status.Uptime = time.Duration(agentStatus.Uptime.Seconds)
	agent.Status.Version = agentStatus.Version
	updatedAgent, err := n.updateAgentStatus(agent)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to update agent %s status error: %v", agent.Name, err)
		return
	}

	agent = updatedAgent
	intResourceVersion, err = strconv.Atoi(agent.ResourceVersion)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to convert agent resource version %s from string to int", agent.ResourceVersion)
		return
	}

	_, err = grpcClient.UpdateAgentSyncVersion(context.TODO(), &proto.InitAgentData{SyncVersion: int32(intResourceVersion)})
	if err != nil {
		log.Log.Reason(err).Errorf("failed to update agent sync version")
		agent.Status.ConnectionStatus = v1.AgentDownStatus
		_, err = n.updateAgentStatus(agent)
		if err != nil {
			log.Log.Reason(err).Errorf("failed to update agent %s to down status error: %v", agent.Name, err)
		}
	}
}

func (n *NativeLBGrpcManager) InitAgent(grpcClient *proto.NativeLoadBalancerAgentClient, agent *v1.Agent, agentNumber, numOfAgents int) (*proto.Agent, *proto.AgentStatus, error) {
	clusterObject, err := n.nativelbClient.Cluster().Get(agent.Spec.Cluster)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to find cluster %s for agent %s", agent.Spec.Cluster, agent.Name)
		return nil, nil, err
	}

	labelSelector := labels.Set{}
	labelSelector[v1.ClusterLabel] = clusterObject.Name
	labelSelector[v1.FarmStatusLabel] = v1.FarmStatusLabelSynced
	farms, err := n.nativelbClient.Farm().List(&client.ListOptions{Namespace: v1.ControllerNamespace, LabelSelector: labelSelector.AsSelector()})
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get the list of farms related to %s cluster", clusterObject.Name)
		return nil, nil, err
	}

	dataList := proto.ConvertFarmsToGrpcDataList(farms.Items, clusterObject, agentNumber, numOfAgents)

	syncVersion, err := strconv.Atoi(agent.ResourceVersion)
	if err != nil {
		return nil, nil, err
	}
	initData := &proto.InitAgentData{SyncVersion: int32(syncVersion), Data: dataList}

	result, err := (*grpcClient).InitAgent(context.TODO(), initData)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to send init data to agent %s error %v", agent.Name, err)
		return nil, nil, err
	}

	return result.Agent, result.AgentStatus, nil
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
	labelSelector := labels.Set{}
	labelSelector[v1.ClusterLabel] = cluster.Name
	agents, err := n.nativelbClient.Agent().List(&client.ListOptions{LabelSelector: labelSelector.AsSelector()})
	if err != nil {
		log.Log.Reason(err).Errorf("failed to get agents list for cluster %s error: %v", cluster.Name, err)
		return err
	}

	if len(agents.Items) == 0 {
		return fmt.Errorf("no agent founds for cluster %s", cluster.Name)
	}

	data := proto.ConvertFarmToGrpcData(farm, cluster.Status.AllocatedNamespaces[farm.Spec.ServiceNamespace].RouterID)

	if cluster.Status.Agents == nil {
		cluster.Status.Agents = make(map[string]*v1.Agent)
	}

	isAnyAgentAlive := false
	numOfAgents := len(agents.Items)
	for idx, agentInstance := range agents.Items {
		agentNumber := idx + 1
		cluster.Status.Agents[agentInstance.Name] = &agentInstance

		conn, err := n.connect(agentInstance.GetUrl())
		if err != nil {
			agentInstance.Status.ConnectionStatus = v1.AgentDownStatus
			_, err = n.updateAgentStatus(&agentInstance)
			if err != nil {
				log.Log.Reason(err).Errorf("failed to update agent %s to down status error: %v", agentInstance.Name, err)
			}
			continue
		}
		defer conn.Close()

		data.KeepalivedState = "MASTER"
		if agentNumber != 1 {
			data.KeepalivedState = "BACKUP"
		}

		data.Priority = int32(10 + agentNumber)
		if int(data.RouterID)%numOfAgents == agentNumber {
			data.Priority += 50
		}

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

func (n *NativeLBGrpcManager) updateAgentStatus(agent *v1.Agent) (*v1.Agent, error) {
	n.updateAgentStatusMutex.Lock()
	agent.Status.LastUpdate = metav1.Time{Time: time.Now()}
	updatedAgent, err := n.nativelbClient.Agent().Update(agent)
	n.updateAgentStatusMutex.Unlock()
	if err != nil {
		return agent, err
	}
	return updatedAgent, nil
}
