package nativelb_agent

import (
	"github.com/k8s-nativelb/pkg/nativelb-agent/loadbalancer"
	"github.com/k8s-nativelb/pkg/nativelb-agent/keepalived"
	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/proto"

	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/vishvananda/netlink"
	"google.golang.org/grpc"

	"time"
	"context"
	"fmt"
	"os"
)

type NativelbAgent struct {
	controllerUrl string
	clusterName string
	controlInterface string
	dataInterface string
	syncInterface string
	loadBalancerController *loadbalancer.LoadBalancer
	keepalivedController *keepalived.Keepalived
	clientStream proto.NativeLoadBalancerAgent_ConnectClient
	agentData *proto.Agent
	stopChan <- chan struct{}
}

type RecvChannelStruct struct {
	command *proto.Command
	err error
}

func NewNativeAgent(controllerUrl,clusterName,controlInterface,dataInterface,syncInterface string) (*NativelbAgent, error) {
	agentData, err := createAgentData(clusterName,controlInterface)
	if err != nil {
		return nil, err
	}

	return &NativelbAgent{controllerUrl:controllerUrl,stopChan:signals.SetupSignalHandler(),
						  keepalivedController:keepalived.NewKeepalived(),
						  loadBalancerController:loadbalancer.NewLoadBalancer(),agentData:agentData,dataInterface:dataInterface,syncInterface:syncInterface}, nil
}

func createAgentData(clusterName,controlInterface string) (*proto.Agent, error) {
	controlInterfaceLink, err := netlink.LinkByName(controlInterface)
	if err != nil {
		return nil, err
	}

	ipAddr, err := netlink.AddrList(controlInterfaceLink,netlink.FAMILY_V4)
	if err != nil {
		return nil, fmt.Errorf("Fail to get ip addresses on interface %s error: %v",controlInterface,err)
	}

	if len(ipAddr) != 1 {
		return nil, fmt.Errorf("Multiple ip addresses on interface %s",controlInterface)
	}

	hostName, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("Fail to get hostname error: %v",err)
	}

	return &proto.Agent{Cluster:clusterName,HostName:hostName,IPAddress:ipAddr[0].IP.String()}, nil
}

func (n *NativelbAgent)connectToController() {
	opts := []grpc.DialOption{grpc.WithInsecure()}

	for true {
		log.Log.V(2).Infof("Trying to connect to %s",n.controllerUrl)
		conn, err := grpc.Dial(n.controllerUrl,opts...)
		if err == nil {
			grpcClient := proto.NewNativeLoadBalancerAgentClient(conn)
			connectClient, err := grpcClient.Connect(context.Background(),n.agentData)
			if err == nil {
				n.clientStream = connectClient
				return
			}
			log.Log.V(2).Errorf("Fail to connect to controller %s error: %v",n.controllerUrl,err)
		} else {
			log.Log.V(2).Errorf("Fail to dial to grpc server %s error: %v",n.controllerUrl,err)
		}

		time.Sleep(5 * time.Second)
	}
}

func receiveDataFromController(recvStream proto.NativeLoadBalancerAgent_ConnectClient ,recvChannel chan RecvChannelStruct) {
	for {
		command, err := recvStream.Recv()
		recvChannel <- RecvChannelStruct{command:command,err:err}
		if err != nil {
			return
		}
	}
}

func (n *NativelbAgent)connectionLoop() (error){
	recvChannel := make(chan RecvChannelStruct)
	go receiveDataFromController(n.clientStream,recvChannel)

	for {
		select {
		case <-n.stopChan:
			n.StopAgent()
			return nil
		case recvStruct := <-recvChannel:
			if recvStruct.err != nil {
				log.Log.V(2).Errorf("Fail to receive message from controller error: %v",recvStruct.err)
				return recvStruct.err
			}
			// TODO Work with the command receive
			log.Log.V(2).Infof("Get command from server %+v",recvStruct.command)
		}
	}
}

func (n *NativelbAgent)StartAgent() {
	for {
		n.connectToController()

		err := n.connectionLoop()
		if err == nil {
			return
		}
	}
}

func (n *NativelbAgent)StopAgent() {
	log.Log.V(2).Infof("Receive stop signal stop keepalived and loadbalancer process")
	//TODO: Stop the processes and cleanup the ip configuration
}