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

	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/vishvananda/netlink"
	"google.golang.org/grpc"

	"context"
	"fmt"
	"os"
	"time"
)

type NativelbAgent struct {
	controllerUrl          string
	clusterName            string
	controlInterface       string
	dataInterface          string
	syncInterface          string
	loadBalancerController *loadbalancer.LoadBalancer
	keepalivedController   *keepalived.Keepalived
	clientStream           NativeLoadBalancerAgent_ConnectClient
	agentData              *Agent
	stopChan               <-chan struct{}
}

type RecvChannelStruct struct {
	command *Command
	err     error
}

func NewNativeAgent(controllerUrl, clusterName, controlInterface, dataInterface, syncInterface string) (*NativelbAgent, error) {
	agentData, err := createAgentData(clusterName, controlInterface)
	if err != nil {
		return nil, err
	}

	return &NativelbAgent{controllerUrl: controllerUrl, stopChan: signals.SetupSignalHandler(),
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

	if len(ipAddr) != 1 {
		return nil, fmt.Errorf("Multiple ip addresses on interface %s", controlInterface)
	}

	hostName, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("Fail to get hostname error: %v", err)
	}

	return &Agent{Cluster: clusterName, HostName: hostName, IPAddress: ipAddr[0].IP.String()}, nil
}

func (n *NativelbAgent) connectToController() {
	opts := []grpc.DialOption{grpc.WithInsecure()}

	for true {
		log.Log.V(2).Infof("Trying to connect to %s", n.controllerUrl)
		conn, err := grpc.Dial(n.controllerUrl, opts...)
		if err == nil {
			grpcClient := NewNativeLoadBalancerAgentClient(conn)
			connectClient, err := grpcClient.Connect(context.Background(), n.agentData)
			if err == nil {
				n.clientStream = connectClient
				return
			}
			log.Log.V(2).Errorf("Fail to connect to controller %s error: %v", n.controllerUrl, err)
		} else {
			log.Log.V(2).Errorf("Fail to dial to grpc server %s error: %v", n.controllerUrl, err)
		}

		time.Sleep(5 * time.Second)
	}
}

func receiveDataFromController(recvStream NativeLoadBalancerAgent_ConnectClient, recvChannel chan RecvChannelStruct) {
	for {
		command, err := recvStream.Recv()
		recvChannel <- RecvChannelStruct{command: command, err: err}
		if err != nil {
			return
		}
	}
}

func (n *NativelbAgent) connectionLoop() error {
	recvChannel := make(chan RecvChannelStruct)
	go receiveDataFromController(n.clientStream, recvChannel)

	for {
		select {
		case <-n.stopChan:
			n.StopAgent()
			return nil
		case recvStruct := <-recvChannel:
			if recvStruct.err != nil {
				log.Log.V(2).Errorf("Fail to receive message from controller error: %v", recvStruct.err)
				return recvStruct.err
			}
			var err error
			switch recvStruct.command.Command {
			case AgentKeepAlive:
				log.Log.V(2).Info("Get keepalive from server")
			case AgentCreateCommand:
				err = n.CreateServers(recvStruct.command.Servers)
			case AgentUpdateCommand:
				err = n.UpdateServers(recvStruct.command.Servers)
			case AgentDeleteCommand:
				err = n.DeleteServers(recvStruct.command.Servers)
			default:
				err = fmt.Errorf("Command not found command: %s", recvStruct.command.Command)
			}

			if err != nil {
				return err
			}
		}
	}
}

func (n *NativelbAgent) StartAgent() {
	for {
		n.connectToController()

		err := n.connectionLoop()
		if err == nil {
			n.clientStream.CloseSend()
		}
	}
}

func (n *NativelbAgent) StopAgent() {
	log.Log.V(2).Infof("Receive stop signal stop keepalived and loadbalancer process")
	//TODO: Stop the processes and cleanup the ip configuration
}
