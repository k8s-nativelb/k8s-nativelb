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
	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/nativelb-agent/keepalived"
	"github.com/k8s-nativelb/pkg/nativelb-agent/loadbalancer"
	. "github.com/k8s-nativelb/pkg/proto"

	"github.com/vishvananda/netlink"
	"google.golang.org/grpc"

	"fmt"
	"net"
	"os"
	"strconv"
)

type NativelbAgent struct {
	controlIPAddress       string
	controlPort            int
	clusterName            string
	controlInterface       string
	dataInterface          string
	syncInterface          string
	loadBalancerController *loadbalancer.LoadBalancer
	keepalivedController   *keepalived.Keepalived
	agentData              *Agent
	grpcServer             *grpc.Server
}

type RecvChannelStruct struct {
	command *Command
	err     error
}

func NewNativeAgent(clusterName, controlInterface, controlPortStr, dataInterface, syncInterface string) (*NativelbAgent, error) {
	agentData, err := createAgentData(clusterName, controlInterface)
	if err != nil {
		return nil, err
	}

	controlPort, err := strconv.Atoi(controlPortStr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert port %s to int error: %v", controlPortStr, err)
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	return &NativelbAgent{controlPort: controlPort, controlIPAddress: agentData.IPAddress, grpcServer: grpcServer,
		keepalivedController:   keepalived.NewKeepalived(),
		loadBalancerController: loadbalancer.NewLoadBalancer(), agentData: agentData, dataInterface: dataInterface, syncInterface: syncInterface}, nil
}

func createAgentData(clusterName, controlInterface string) (*Agent, error) {
	controlInterfaceLink, err := netlink.LinkByName(controlInterface)
	if err != nil {
		return nil, err
	}

	ipAddr, err := netlink.AddrList(controlInterfaceLink, netlink.FAMILY_V4)
	if err != nil {
		return nil, fmt.Errorf("Fail to get ip addresses on interface %s error: %v", controlInterface, err)
	}

	if len(ipAddr) == 0 {
		return nil, fmt.Errorf("None ip addresses on interface %s", controlInterface)
	}

	if len(ipAddr) != 1 {
		return nil, fmt.Errorf("Multiple ip addresses on interface %s", controlInterface)
	}

	hostName, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("Fail to get hostname error: %v", err)
	}

	return &Agent{Cluster: clusterName, HostName: hostName, IPAddress: ipAddr[0].IP.String()}, nil
}

func (n *NativelbAgent) StartAgent() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", n.agentData.IPAddress, n.controlPort))
	if err != nil {
		return fmt.Errorf("failed to listen on %s error: %v", fmt.Sprintf("%s:%d", n.agentData.IPAddress, n.controlPort), err)
	}

	RegisterNativeLoadBalancerAgentServer(n.grpcServer, n)
	return n.grpcServer.Serve(lis)
}

func (n *NativelbAgent) StopAgent(stopChan chan os.Signal) {
	<-stopChan
	log.Log.V(2).Infof("Receive stop signal stop keepalived, loadbalancer process and grpc server")
	n.grpcServer.Stop()
	//TODO: Stop the processes and cleanup the ip configuration
}
