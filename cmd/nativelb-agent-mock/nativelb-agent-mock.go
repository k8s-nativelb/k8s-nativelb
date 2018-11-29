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

package main

import (
	"context"
	"fmt"
	"github.com/k8s-nativelb/pkg/log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/golang/protobuf/ptypes/duration"
	. "github.com/k8s-nativelb/pkg/proto"

	"github.com/vishvananda/netlink"
	"google.golang.org/grpc"
)

func main() {
	clusterName, isExist := os.LookupEnv("CLUSTER_NAME")
	if !isExist {
		panic(fmt.Errorf("CLUSTER_NAME environment variable doesn't exist"))
	}

	controlIP, isExist := os.LookupEnv("CONTROL_IP")
	if !isExist {
		panic(fmt.Errorf("CONTROL_INTERFACE environment variable doesn't exist"))
	}

	controlPort, isExist := os.LookupEnv("CONTROL_PORT")
	if !isExist {
		panic(fmt.Errorf("CONTROL_PORT environment variable doesn't exist"))
	}

	dataInterface, isExist := os.LookupEnv("DATA_INTERFACE")
	if !isExist || dataInterface == "" {
		dataInterface = ""
	}

	syncInterface, isExist := os.LookupEnv("SYNC_INTERFACE")
	if !isExist {
		syncInterface = ""
	}

	agent, err := NewNativeAgent(clusterName, controlIP, controlPort, dataInterface, syncInterface)
	if err != nil {
		panic(err)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	go agent.StopAgent(stopChan)

	err = agent.StartAgent()
	if err != nil {
		panic(err)
	}
}

type NativelbAgent struct {
	agent       *Agent
	agentStatus *AgentStatus
	grpcServer  *grpc.Server
}

type RecvChannelStruct struct {
	command *Command
	err     error
}

func NewNativeAgent(clusterName, controlIp, controlPortStr, dataInterface, syncInterface string) (*NativelbAgent, error) {
	agentData, err := createAgentData(clusterName, controlIp)
	if err != nil {
		return nil, err
	}

	if dataInterface == "" {
		agentData.DataInterface = agentData.ControlInterface
	} else {
		agentData.DataInterface = dataInterface
	}

	if syncInterface == "" {
		agentData.SyncInterface = agentData.ControlInterface
	} else {
		agentData.SyncInterface = syncInterface
	}

	controlPort, err := strconv.Atoi(controlPortStr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert port %s to integer error %v", controlPortStr, err)
	}

	agentData.Port = int32(controlPort)

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	return &NativelbAgent{agent: agentData, grpcServer: grpcServer, agentStatus: &AgentStatus{Status: AgentNewStatus}}, nil
}

func createAgentData(clusterName, controlIp string) (*Agent, error) {
	interfacesList, err := netlink.LinkList()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to list interfaces")
		return nil, err
	}

	for _, iface := range interfacesList {
		ifaceName := iface.Attrs().Name

		controlInterfaceLink, err := netlink.LinkByName(ifaceName)
		if err != nil {
			return nil, err
		}

		ipAddrs, err := netlink.AddrList(controlInterfaceLink, netlink.FAMILY_V4)
		if err != nil {
			return nil, fmt.Errorf("failed to get ip addresses on interface %s error %v", iface, err)
		}

		for _, ipaddr := range ipAddrs {
			if ipaddr.IP.String() == controlIp {

				hostName, err := os.Hostname()
				if err != nil {
					return nil, fmt.Errorf("failed to get hostname error %v", err)
				}

				return &Agent{Cluster: clusterName, HostName: hostName, IPAddress: controlIp, ControlInterface: ifaceName}, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to find interface with %s ip address", controlIp)
}

func (n *NativelbAgent) StartAgent() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", n.agent.IPAddress, n.agent.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on %s error: %v", fmt.Sprintf("%s:%d", n.agent.IPAddress, n.agent.Port), err)
	}

	RegisterNativeLoadBalancerAgentServer(n.grpcServer, n)
	log.Log.Infof("Starting agent on %s address", fmt.Sprintf("%s:%d", n.agent.IPAddress, n.agent.Port))
	return n.grpcServer.Serve(lis)
}

func (n *NativelbAgent) StopAgent(stopChan chan os.Signal) {
	<-stopChan
	log.Log.V(2).Infof("Receive stop signal stop keepalived, loadbalancer process and grpc server")
	n.grpcServer.Stop()
	//TODO: Stop the processes and cleanup the ip configuration
}

func (n *NativelbAgent) CreateServers(ctx context.Context, data *Data) (*Result, error) {
	log.Log.Infof("CreateServers grpc call with data: %v", *data)
	if n.agentStatus.Status != AgentSyncedStatus {
		return nil, fmt.Errorf("failed to create servers for farm %s because the agent is not in synced status", data.FarmName)
	}
	return &Result{}, nil
}
func (n *NativelbAgent) UpdateServers(ctx context.Context, data *Data) (*Result, error) {
	log.Log.Infof("UpdateServers grpc call with data: %v", *data)
	if n.agentStatus.Status != AgentSyncedStatus {
		return nil, fmt.Errorf("failed to update servers for farm %s because the agent is not in synced status", data.FarmName)
	}
	return &Result{}, nil
}
func (n *NativelbAgent) DeleteServers(ctx context.Context, data *Data) (*Result, error) {
	log.Log.Infof("DeleteServers grpc call with data: %v", *data)
	if n.agentStatus.Status != AgentSyncedStatus {
		return nil, fmt.Errorf("failed to delete servers for farm %s because the agent is not in synced status", data.FarmName)
	}
	return &Result{}, nil
}
func (n *NativelbAgent) GetAgentStatus(ctx context.Context, cmd *Command) (*AgentStatus, error) {
	log.Log.Infof("GetAgentStatus grpc call with command: %v", cmd)
	return n.agentStatus, nil
}
func (n *NativelbAgent) GetServerStats(ctx context.Context, cmd *Command) (*ServerStats, error) {
	log.Log.Infof("GetServerStats grpc call with command: %v", cmd)
	if n.agentStatus.Status != AgentSyncedStatus {
		return nil, fmt.Errorf("failed to get server stats because the agent is not in synced status")
	}
	return &ServerStats{}, nil
}

func (n *NativelbAgent) InitAgent(ctx context.Context, data *InitAgentData) (*InitAgentResult, error) {
	log.Log.Infof("InitAgent grpc call with initData: %v", *data)
	// TODO: load all the farms and start the keepalived and gobetween processes

	n.agentStatus.SyncVersion = data.SyncVersion
	n.agentStatus.Status = AgentSyncedStatus

	//TODO: remove this after the agent start a real nginx process
	n.agentStatus.KeepAlivedPid = 1
	n.agentStatus.LBPid = 1
	n.agentStatus.Uptime = &duration.Duration{Seconds: 1}

	return &InitAgentResult{Agent: n.agent, AgentStatus: n.agentStatus}, nil
}

func (n *NativelbAgent) UpdateAgentSyncVersion(ctx context.Context, data *InitAgentData) (*Result, error) {
	log.Log.Infof("get UpdateAgentSyncVersion status grpc call with initData: %v", *data)
	n.agentStatus.SyncVersion = data.SyncVersion
	return &Result{}, nil
}
