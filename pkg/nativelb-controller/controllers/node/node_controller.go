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

package node

import (
	"context"
	"github.com/k8s-nativelb/pkg/kubecli"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/service"
)

type NodeController struct {
	Controller    controller.Controller
	ReconcileNode reconcile.Reconciler
}

func NewNodeController(nativelbClient kubecli.NativelbClient, serviceController *service_controller.ServiceController) (*NodeController, error) {
	reconcileNode := newReconciler(nativelbClient, serviceController)

	controllerInstance, err := newNodeControllerController(nativelbClient, reconcileNode)
	if err != nil {
		return nil, err
	}
	nodeController := &NodeController{Controller: controllerInstance,
		ReconcileNode: reconcileNode}

	return nodeController, nil

}

func loadNodes(kubeClient *kubernetes.Clientset) map[string]string {
	nodes, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	nodeMap := make(map[string]string)
	for _, nodeInstance := range nodes.Items {
		for _, IpAddr := range nodeInstance.Status.Addresses {
			if IpAddr.Type == "InternalIP" {
				if value, ok := nodeMap[nodeInstance.Name]; !ok || value != IpAddr.Address {
					nodeMap[nodeInstance.Name] = IpAddr.Address
				}
			}
		}
	}

	return nodeMap
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(nativelbClient kubecli.NativelbClient, serviceController *service_controller.ServiceController) *ReconcileNode {

	return &ReconcileNode{NativelbClient: nativelbClient,
		serviceController: serviceController,
		scheme:            nativelbClient.GetScheme(),
		Event:             nativelbClient.GetRecorder(v1.EventRecorderName),
		NodeMap:           loadNodes(nativelbClient.GetKubeClient())}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func newNodeControllerController(nativelbClient kubecli.NativelbClient, r reconcile.Reconciler) (controller.Controller, error) {
	// Create a new controller
	c, err := controller.New("node-controller", nativelbClient.GetManager(), controller.Options{Reconciler: r})
	if err != nil {
		return nil, err
	}

	// Watch for changes to Node
	err = c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return nil, err
	}

	return c, nil
}

var _ reconcile.Reconciler = &ReconcileNode{}

// ReconcileNode reconciles a Node object
type ReconcileNode struct {
	kubecli.NativelbClient
	Event             record.EventRecorder
	serviceController *service_controller.ServiceController
	scheme            *runtime.Scheme
	NodeMap           map[string]string
}

// Reconcile reads that state of the cluster for a Node object and makes changes based on the state read
// and what is in the Node.Spec
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
func (r *ReconcileNode) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	nativeClient := r.GetClient()
	// Fetch the Node instance
	nodeInstance := &corev1.Node{}
	err := nativeClient.Get(context.TODO(), request.NamespacedName, nodeInstance)
	if err != nil && !errors.IsNotFound(err) {
		log.Log.Errorf("Fail to reconcile node error message: %v", err)

		return reconcile.Result{}, err
	}

	if r.needToUpdateServices() {
		r.serviceController.UpdateAllServices()
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileNode) needToUpdateServices() bool {
	kubeClient := r.GetKubeClient()
	nodeList, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Log.Error("fail to get node list")
	}

	if len(nodeList.Items) != len(r.NodeMap) {
		r.NodeMap = loadNodes(kubeClient)
		return true
	}

	for _, nodeInstance := range nodeList.Items {
		if value, ok := r.NodeMap[nodeInstance.Name]; !ok {
			r.NodeMap = loadNodes(kubeClient)
			return true
		} else {
			for _, IpAddr := range nodeInstance.Status.Addresses {
				if IpAddr.Type == "InternalIP" && value != IpAddr.Address {
					r.NodeMap = loadNodes(kubeClient)
					return true
				}
			}
		}
	}

	return false
}
