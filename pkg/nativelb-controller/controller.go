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
	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/apis"
	"github.com/k8s-nativelb/pkg/nativelb-controller/server"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/agent"

	"k8s.io/client-go/kubernetes"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/backend"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/server"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/farm"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/cluster"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/service"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/node"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/endpoint"
)

type NativeLBManager struct {
	manager.Manager
	kubeClient *kubernetes.Clientset
	nativeLBGrpcServer *server.NativeLBGrpcServer

	agentController *agent_controller.AgentController
	backendController *backend_controller.BackendController
	serverController *server_controller.ServerController
	farmController *farm_controller.FarmController
	clusterController *cluster_controller.ClusterController
}

func NewNativeLBManager() (*NativeLBManager) {
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		panic(err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	log.Log.Infof("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		panic(err)
	}


	nativeLBGrpcServer := server.NewNativeLBGrpcServer(mgr.GetClient())
	nativeLBManager := &NativeLBManager{mgr,
	kubeClient,
	nativeLBGrpcServer,
	nil,
	nil,
	nil,
	nil,
	nil}

	err = nativeLBManager.addToManager()
	if err != nil {
		panic(err)
	}

	return nativeLBManager
}

func (n *NativeLBManager)StartManager() {
	go n.nativeLBGrpcServer.StartServer()

	log.Log.Infof("Starting Native LB Manager")
	n.GetCache().WaitForCacheSync(signals.SetupSignalHandler())

	//Start channel listener on controllers
	go n.agentController.WaitForStatusUpdate()
	go n.serverController.WaitForStatusUpdate()

	n.Start(signals.SetupSignalHandler())
}


// AddToManager adds all Controllers to the Manager
func(n *NativeLBManager) addToManager() error {

	log.Log.V(2).Infof("Creating Agent controller")
	agentController, err := agent_controller.NewAgentController(n.Manager,n.nativeLBGrpcServer.AgentStatusChannel)
	if err != nil {
		return err
	}
	n.agentController = agentController

	log.Log.V(2).Infof("Creating Cluster controller")
	clusterController, err := cluster_controller.NewClusterController(n.Manager,agentController,n.nativeLBGrpcServer)
	if err != nil {
		return err
	}
	n.clusterController = clusterController


	log.Log.V(2).Infof("Creating Backend controller")
	backendController, err := backend_controller.NewBackendController(n.Manager)
	if err != nil {
		return err
	}
	n.backendController = backendController

	log.Log.V(2).Infof("Creating Server controller")
	serverController, err := server_controller.NewServerController(n.Manager,backendController,n.nativeLBGrpcServer.ServerStats)
	if err != nil {
		return err
	}
	n.serverController = serverController

	log.Log.V(2).Infof("Creating Farm controller")
	farmController, err := farm_controller.NewFarmController(n.Manager,serverController,clusterController)
	if err != nil {
		return err
	}
	n.farmController = farmController

	log.Log.V(2).Infof("Creating Service controller")
	serviceController, err := service_controller.NewServiceController(n.Manager, n.kubeClient, farmController)
	if err != nil {
		return err
	}

	log.Log.V(2).Infof("Creating Node controller")
	_, err = node.NewNodeController(n.Manager, n.kubeClient, serviceController)
	if err != nil {
		return err
	}

	log.Log.V(2).Infof("Creating Endpoint controller")
	_, err = endpoint.NewEndPointController(n.Manager, n.kubeClient, serviceController)
	if err != nil {
		return err
	}
	return nil
}
