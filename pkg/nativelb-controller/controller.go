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
package nativelb_controller

import (
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/kubecli"
	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/nativelb-controller/grpc-manager"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/agent"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/backend"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/cluster"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/endpoint"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/farm"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/node"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/server"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/service"
)

type NativeLBManager struct {
	nativelbCli         kubecli.NativelbClient
	nativeLBGrpcManager *grpc_manager.NativeLBGrpcManager

	agentController   *agent_controller.AgentController
	backendController *backend_controller.BackendController
	serverController  *server_controller.ServerController
	farmController    *farm_controller.FarmController
	clusterController *cluster_controller.ClusterController

	stopChan <-chan struct{}
}

func NewNativeLBManager() *NativeLBManager {
	nativelbCli, err := kubecli.GetNativelbClient()
	if err != nil {
		panic(err)
	}

	stopChan := signals.SetupSignalHandler()
	nativeLBGrpcManager := grpc_manager.NewNativeLBGrpcManager(nativelbCli, stopChan)
	nativeLBManager := &NativeLBManager{nativelbCli,
		nativeLBGrpcManager,
		nil,
		nil,
		nil,
		nil,
		nil, stopChan}

	err = nativeLBManager.addToManager()
	if err != nil {
		panic(err)
	}

	return nativeLBManager
}

func (n *NativeLBManager) StartManager() {
	log.Log.Infof("Clear nativeLB Labels from all the services")
	err := n.ClearLabels()
	if err != nil {
		log.Log.Errorf("Fail to clean labels from all the services on the cluster error: %v", err)
		panic(err)
	}

	log.Log.Infof("Starting Native LB Manager")

	go n.nativeLBGrpcManager.StartKeepalive()

	n.nativelbCli.GetManager().Start(n.stopChan)
}

// AddToManager adds all Controllers to the Manager
func (n *NativeLBManager) addToManager() error {

	log.Log.V(2).Infof("Creating Agent controller")
	agentController, err := agent_controller.NewAgentController(n.nativelbCli)
	if err != nil {
		return err
	}
	n.agentController = agentController

	log.Log.V(2).Infof("Creating Cluster controller")
	clusterController, err := cluster_controller.NewClusterController(n.nativelbCli, agentController, n.nativeLBGrpcManager)
	if err != nil {
		return err
	}
	n.clusterController = clusterController

	log.Log.V(2).Infof("Creating Backend controller")
	backendController, err := backend_controller.NewBackendController(n.nativelbCli)
	if err != nil {
		return err
	}
	n.backendController = backendController

	log.Log.V(2).Infof("Creating Server controller")
	serverController, err := server_controller.NewServerController(n.nativelbCli, backendController)
	if err != nil {
		return err
	}
	n.serverController = serverController

	log.Log.V(2).Infof("Creating Farm controller")
	farmController, err := farm_controller.NewFarmController(n.nativelbCli, serverController, clusterController)
	if err != nil {
		return err
	}
	n.farmController = farmController

	log.Log.V(2).Infof("Creating Service controller")
	serviceController, err := service_controller.NewServiceController(n.nativelbCli, farmController)
	if err != nil {
		return err
	}

	log.Log.V(2).Infof("Creating Node controller")
	_, err = node.NewNodeController(n.nativelbCli, serviceController)
	if err != nil {
		return err
	}

	log.Log.V(2).Infof("Creating Endpoint controller")
	_, err = endpoint.NewEndPointController(n.nativelbCli, serviceController)
	if err != nil {
		return err
	}
	return nil
}

func (n *NativeLBManager) ClearLabels() error {
	kubeClient := n.nativelbCli.GetKubeClient()
	serviceList, err := kubeClient.CoreV1().Services("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, serviceObject := range serviceList.Items {
		if serviceObject.Labels != nil {
			if _, ok := serviceObject.Labels[v1.ServiceStatusLabel]; ok {
				delete(serviceObject.Labels, v1.ServiceStatusLabel)
				_, err := kubeClient.CoreV1().Services(serviceObject.Namespace).Update(&serviceObject)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
