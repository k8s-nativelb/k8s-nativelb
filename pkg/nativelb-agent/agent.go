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
	agent                  *Agent
	loadBalancerController *loadbalancer.LoadBalancer
	keepalivedController   *keepalived.Keepalived
	grpcServer             *grpc.Server
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
		return nil, fmt.Errorf("failed to convert port %s to int error: %v", controlPortStr, err)
	}

	agentData.Port = int32(controlPort)

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	return &NativelbAgent{agent: agentData, grpcServer: grpcServer,
		keepalivedController:   keepalived.NewKeepalived(),
		loadBalancerController: loadbalancer.NewLoadBalancer()}, nil
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
			return nil, fmt.Errorf("Fail to get ip addresses on interface %s error: %v", iface, err)
		}

		for _, ipaddr := range ipAddrs {
			if ipaddr.IP.String() == controlIp {

				hostName, err := os.Hostname()
				if err != nil {
					return nil, fmt.Errorf("Fail to get hostname error: %v", err)
				}

				return &Agent{Cluster: clusterName, HostName: hostName, IPAddress: controlIp, ControlInterface: ifaceName}, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to find interface related to ip address %s", controlIp)
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
