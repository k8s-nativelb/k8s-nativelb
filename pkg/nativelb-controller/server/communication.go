package server

import (
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
)

func (n *NativeLBGrpcServer) CreateFarmOnCluster(farm *v1.Farm, cluster *v1.Cluster) error {
	return nil
}

func (n *NativeLBGrpcServer) UpdateFarmOnCluster(farm *v1.Farm, cluster *v1.Cluster) error {
	return nil
}

func (n *NativeLBGrpcServer) DeleteFarmOnCluster(farm *v1.Farm, cluster *v1.Cluster) error {
	return nil
}
