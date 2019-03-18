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
	"github.com/k8s-nativelb/pkg/nativelb-agent/udp-loadbalancer"
	. "github.com/k8s-nativelb/pkg/proto"

	"github.com/vishvananda/netlink"
	"google.golang.org/grpc"

	"fmt"
	"net"
	"os"
	"strconv"
)

type NativelbAgent struct {
	agent                     *Agent
	agentStatus               *AgentStatus
	loadBalancerController    loadbalancer.LoadBalancerInterface
	udpLoadBalancerController udp_loadbalancer.UdpLoadBalancerInterface
	keepalivedController      keepalived.KeepalivedInterface
	grpcServer                *grpc.Server
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

	keepalivedController, err := keepalived.NewKeepalived(agentData.DataInterface)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to create a keepalived controller error")
		return nil, err
	}

	loadbalancerController, err := loadbalancer.NewLoadBalancer()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to create a loadbalancer controller")
		return nil, err
	}

	udpLoadBalancer, err := udp_loadbalancer.NewUdpLoadBalancer()
	if err != nil {
		log.Log.Reason(err).Errorf("failed to create a udp loadbalancer(nginx) controller")
		return nil, err
	}

	return &NativelbAgent{agent: agentData, grpcServer: grpcServer, agentStatus: &AgentStatus{Status: AgentNewStatus},
		keepalivedController:      keepalivedController,
		loadBalancerController:    loadbalancerController,
		udpLoadBalancerController: udpLoadBalancer}, nil
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
	log.Log.Infof("Starting listen for controller communication on interface %s agent ip address %s", n.agent.ControlInterface, fmt.Sprintf("%s:%d", n.agent.IPAddress, n.agent.Port))
	return n.grpcServer.Serve(lis)
}

func (n *NativelbAgent) WaitForStopAgent(stopChan chan os.Signal) {
	<-stopChan
	log.Log.Infof("Receive stop signal stop keepalived, loadbalancer process and grpc server")
	n.grpcServer.Stop()
	n.StopEngines()
}
