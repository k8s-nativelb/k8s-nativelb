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
	"github.com/k8s-nativelb/pkg/kubecli"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/client-go/tools/record"

	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/service"
)

type EndPointController struct {
	Controller    controller.Controller
	ReconcileNode reconcile.Reconciler
}

func NewEndPointController(nativelbClient kubecli.NativelbClient, serviceController *service_controller.ServiceController) (*EndPointController, error) {
	reconcileEndPoint := newReconciler(nativelbClient, serviceController)

	controllerInstance, err := newEndPointController(nativelbClient, reconcileEndPoint)
	if err != nil {
		return nil, err
	}
	endpointController := &EndPointController{Controller: controllerInstance,
		ReconcileNode: reconcileEndPoint}

	return endpointController, nil

}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(nativelbClient kubecli.NativelbClient, serviceController *service_controller.ServiceController) *ReconcileEndPoint {
	return &ReconcileEndPoint{NativelbClient: nativelbClient,
		serviceController: serviceController,
		scheme:            nativelbClient.GetScheme(),
		Event:             nativelbClient.GetRecorder(v1.EventRecorderName)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func newEndPointController(nativelbClient kubecli.NativelbClient, r reconcile.Reconciler) (controller.Controller, error) {
	// Create a new controller
	c, err := controller.New("endpoint-controller", nativelbClient.GetManager(), controller.Options{Reconciler: r})
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
	kubecli.NativelbClient
	Event             record.EventRecorder
	serviceController *service_controller.ServiceController
	scheme            *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Endpoints object and makes changes based on the state read
// and what is in the Endpoints.Spec
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
func (r *ReconcileEndPoint) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	nativeClient := r.GetClient()
	// Fetch the Node instance
	endpoint := &corev1.Endpoints{}
	err := nativeClient.Get(context.TODO(), request.NamespacedName, endpoint)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			return reconcile.Result{}, nil
		}

		log.Log.Errorf("Fail to reconcile endpoint error message: %v", err)
		return reconcile.Result{}, err
	}

	if len(endpoint.Subsets) > 0 {
		service := &corev1.Service{}
		err := nativeClient.Get(context.TODO(), client.ObjectKey{Namespace: endpoint.Namespace, Name: endpoint.Name}, service)
		if err == nil && service.Spec.Type == "LoadBalancer" {
			if status, ok := service.Labels[v1.ServiceStatusLabel]; ok && status == v1.ServiceStatusLabelSynced {
				log.Log.V(2).Infof("Endpoint event for service name: %s on namespace %s", service.Name, service.Namespace)
				err = r.serviceController.ReconcileService.UpdateEndpoints(service, endpoint)
				if err != nil {
					return reconcile.Result{Requeue: true}, nil
				}
			}
		}
	}
	return reconcile.Result{}, nil
}
