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

package endpoint

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/service"
)

type EndPointController struct {
	Controller    controller.Controller
	ReconcileNode reconcile.Reconciler
}

func NewEndPointController(mgr manager.Manager, kubeClient *kubernetes.Clientset, serviceController *service_controller.ServiceController) (*EndPointController, error) {
	reconcileEndPoint := newReconciler(mgr, kubeClient, serviceController)

	controllerInstance, err := newEndPointController(mgr, reconcileEndPoint)
	if err != nil {
		return nil, err
	}
	endpointController := &EndPointController{Controller: controllerInstance,
		ReconcileNode: reconcileEndPoint}

	return endpointController, nil

}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, kubeClient *kubernetes.Clientset, serviceController *service_controller.ServiceController) *ReconcileEndPoint {
	return &ReconcileEndPoint{Client: mgr.GetClient(),
		kubeClient:        kubeClient,
		serviceController: serviceController,
		scheme:            mgr.GetScheme(),
		Event:             mgr.GetRecorder(v1.EventRecorderName)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func newEndPointController(mgr manager.Manager, r reconcile.Reconciler) (controller.Controller, error) {
	// Create a new controller
	c, err := controller.New("endpoint-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return nil, err
	}

	// Watch for changes to Node
	err = c.Watch(&source.Kind{Type: &corev1.Endpoints{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return nil, err
	}

	return c, nil
}

var _ reconcile.Reconciler = &ReconcileEndPoint{}

// ReconcileNode reconciles a Endpoints object
type ReconcileEndPoint struct {
	client.Client
	kubeClient        *kubernetes.Clientset
	Event             record.EventRecorder
	serviceController *service_controller.ServiceController
	scheme            *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Endpoints object and makes changes based on the state read
// and what is in the Endpoints.Spec
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
func (r *ReconcileEndPoint) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Node instance
	endpoint := &corev1.Endpoints{}
	err := r.Get(context.TODO(), request.NamespacedName, endpoint)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			return reconcile.Result{}, nil
		}

		log.Log.Errorf("Fail to reconcile endpoint error message: %s", err.Error())
		return reconcile.Result{}, err
	}

	if len(endpoint.Subsets) > 0 {
		service := &corev1.Service{}
		err := r.Get(context.Background(), client.ObjectKey{Namespace: endpoint.Namespace, Name: endpoint.Name}, service)
		if err == nil && service.Spec.Type == "LoadBalancer" {
			r.serviceController.ReconcileService.UpdateEndpoints(service, endpoint)
		}
	}
	return reconcile.Result{}, nil
}
