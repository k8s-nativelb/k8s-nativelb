package server

import (
	"fmt"
	"net"

	"github.com/k8s-nativelb/pkg/log"
	"google.golang.org/grpc"

	pb "github.com/k8s-nativelb/pkg/proto"
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RunTimeAgent struct {
	Data       *pb.Agent
	connection pb.NativeLoadBalancerAgent_ConnectServer
}

type NativeLBGrpcServer struct {
	client client.Client
	GrpcServer *grpc.Server
	Cluster map[string][]*RunTimeAgent

	Connection chan pb.Agent
	AgentStatusChannel chan pb.AgentStatus
	ServerStats chan pb.ServerStats
	NewAgentChannel chan pb.Agent

}

func NewNativeLBGrpcServer(client client.Client) (*NativeLBGrpcServer) {
	return &NativeLBGrpcServer{client:client,GrpcServer:grpc.NewServer(),Cluster:make(map[string][]*RunTimeAgent),
								AgentStatusChannel:make(chan pb.AgentStatus,10),
								ServerStats:make(chan pb.ServerStats, 10),
								NewAgentChannel:make(chan pb.Agent,10)}
}

func(n *NativeLBGrpcServer) StartServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 8080))
	if err != nil {
		log.Log.Errorf("failed to listen: %v", err)
		panic(err)
	}

	pb.RegisterNativeLoadBalancerAgentServer(n.GrpcServer,n)

	log.Log.Infof("GRPC server start lisening on 0.0.0.0:%d", 8080)
	n.GrpcServer.Serve(lis)
}

func(n *NativeLBGrpcServer) StopServer() {
	n.GrpcServer.Stop()
}


func (n *NativeLBGrpcServer) Connect(agent *pb.Agent, con pb.NativeLoadBalancerAgent_ConnectServer) error {

	return nil
}

func (n *NativeLBGrpcServer) UpdateAgentStatus(context context.Context, agentStatus *pb.AgentStatus) (*pb.Result, error) {
	return &pb.Result{}, nil
}

func (n *NativeLBGrpcServer) UpdateServerStats(context context.Context, serverStats *pb.ServerStats) (*pb.Result, error) {
	return &pb.Result{}, nil
}