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

package cluster_controller

import (
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/kubecli"
	"github.com/k8s-nativelb/pkg/nativelb-controller/grpc-manager"

	//pb "github.com/k8s-nativelb/pkg/proto"
	"github.com/k8s-nativelb/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/agent"
)

type ClusterController struct {
	kubecli.NativelbClient
	Controller        controller.Controller
	Reconcile         *Reconcile
	agentController   *agent_controller.AgentController
	clusterConnection *grpc_manager.NativeLBGrpcManager
	allocator         map[string]*Allocator
}

func NewClusterController(nativelbClient kubecli.NativelbClient, agentController *agent_controller.AgentController, clusterConnection *grpc_manager.NativeLBGrpcManager) (*ClusterController, error) {
	reconcileInstance := newReconciler(nativelbClient)
	controllerInstance, err := newController(nativelbClient, reconcileInstance)
	if err != nil {
		return nil, err
	}

	clusterController := &ClusterController{NativelbClient: nativelbClient, Controller: controllerInstance,
		Reconcile: reconcileInstance, agentController: agentController, allocator: make(map[string]*Allocator), clusterConnection: clusterConnection}

	return clusterController, nil
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(nativelbClient kubecli.NativelbClient) *Reconcile {
	return &Reconcile{NativelbClient: nativelbClient,
		scheme: nativelbClient.GetScheme(),
		Event:  nativelbClient.GetRecorder(v1.EventRecorderName)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func newController(nativelbClient kubecli.NativelbClient, r reconcile.Reconciler) (controller.Controller, error) {
	// Create a new controller
	c, err := controller.New("cluster-controller", nativelbClient.GetManager(), controller.Options{Reconciler: r})
	if err != nil {
		return nil, err
	}

	// Watch for changes to Provider
	err = c.Watch(&source.Kind{Type: &v1.Cluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return nil, err
	}

	return c, nil
}

var _ reconcile.Reconciler = &Reconcile{}

// Reconcile reconcile object
type Reconcile struct {
	kubecli.NativelbClient
	Event  record.EventRecorder
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Agent object and makes changes based on the state read
// and what is in the Agent.Spec
// +kubebuilder:rbac:groups=k8s.native-lb,resources=cluster,verbs=get;list;watch;create;update;patch;delete
func (r *Reconcile) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance, err := r.Cluster().Get(request.NamespacedName.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	log.Log.V(4).Infof("%+v\n", instance)
	//TODO: Create and remove allocators for every new cluster created, updated or deleted
	return reconcile.Result{}, nil
}
