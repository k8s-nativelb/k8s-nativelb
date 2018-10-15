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
	"github.com/k8s-nativelb/pkg/kubecli"
	"net"
	"time"

	"github.com/k8s-nativelb/pkg/log"
	"google.golang.org/grpc"

	"context"
	pb "github.com/k8s-nativelb/pkg/proto"
)

type RunTimeAgent struct {
	Data       *pb.Agent
	connection pb.NativeLoadBalancerAgent_ConnectServer
}

type NativeLBGrpcServer struct {
	nativelbClient kubecli.NativelbClient
	GrpcServer     *grpc.Server
	Cluster        map[string][]*RunTimeAgent

	Connection         chan pb.Agent
	AgentStatusChannel chan pb.AgentStatus
	ServerStats        chan pb.ServerStats
	NewAgentChannel    chan pb.Agent
	stopChan           <-chan struct{}
}

func NewNativeLBGrpcServer(nativelbClient kubecli.NativelbClient, stopChan <-chan struct{}) *NativeLBGrpcServer {
	return &NativeLBGrpcServer{nativelbClient: nativelbClient, GrpcServer: grpc.NewServer(), Cluster: make(map[string][]*RunTimeAgent),
		AgentStatusChannel: make(chan pb.AgentStatus, 10),
		ServerStats:        make(chan pb.ServerStats, 10),
		NewAgentChannel:    make(chan pb.Agent, 10), stopChan: stopChan}
}

func (n *NativeLBGrpcServer) StartServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 8080))
	if err != nil {
		log.Log.Errorf("failed to listen: %v", err)
		panic(err)
	}

	pb.RegisterNativeLoadBalancerAgentServer(n.GrpcServer, n)

	log.Log.Infof("GRPC server start lisening on 0.0.0.0:%d", 8080)
	go n.StopServer()
	n.GrpcServer.Serve(lis)
}

func (n *NativeLBGrpcServer) StopServer() {
	<-n.stopChan
	// TODO: Close all open clients
	n.GrpcServer.Stop()
}

func (n *NativeLBGrpcServer) Connect(agent *pb.Agent, con pb.NativeLoadBalancerAgent_ConnectServer) error {
	runTimeAgent := &RunTimeAgent{Data: agent, connection: con}
	_, err := n.nativelbClient.Cluster().Get(agent.Cluster)
	if err != nil {
		log.Log.V(2).Errorf("Receive a connection message from %+v but fail to find the cluster with error: %v", agent, err)
		return fmt.Errorf("Fail to find the cluster name %s", agent.Cluster)
	}

	_, isExist := n.Cluster[agent.Cluster]
	if !isExist {
		n.Cluster[agent.Cluster] = make([]*RunTimeAgent, 0)
	}

	n.Cluster[agent.Cluster] = append(n.Cluster[agent.Cluster], runTimeAgent)

	command := &pb.Command{Command: "keepalive"}

	for {
		err := con.Send(command)
		if err != nil {
			// TODO: Disconnect the client
			return nil
		}

		time.Sleep(30 * time.Second)
	}
}

func (n *NativeLBGrpcServer) UpdateAgentStatus(context context.Context, agentStatus *pb.AgentStatus) (*pb.Result, error) {
	log.Log.Infof("Get agent update status message from %s", agentStatus.Name)
	return &pb.Result{}, nil
}

func (n *NativeLBGrpcServer) UpdateServerStats(context context.Context, serverStats *pb.ServerStats) (*pb.Result, error) {
	log.Log.Infof("Get server update status message from %s", serverStats.Name)
	return &pb.Result{}, nil
}
