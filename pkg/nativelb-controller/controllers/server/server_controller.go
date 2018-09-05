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

package server_controller

import (
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	//pb "github.com/k8s-nativelb/pkg/proto"
	"github.com/k8s-nativelb/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/client-go/tools/record"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/api/errors"

	"context"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/backend"
	pb "github.com/k8s-nativelb/pkg/proto"
)

type ServerController struct {
	Controller        controller.Controller
	Reconcile *Reconcile
	backendController *backend_controller.BackendController
	serverStatsChannel chan pb.ServerStats
}

func NewServerController(mgr manager.Manager,backendController *backend_controller.BackendController,serverStatsChannel chan pb.ServerStats) (*ServerController, error) {
	reconcileInstance := newReconciler(mgr)
	controllerInstance, err := newController(mgr, reconcileInstance)
	if err != nil {
		return nil, err
	}

	serverController := &ServerController{Controller: controllerInstance,
		Reconcile: reconcileInstance,backendController:backendController,serverStatsChannel:serverStatsChannel}

	return serverController, nil
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, ) *Reconcile {
	return &Reconcile{Client: mgr.GetClient(),
		scheme:     mgr.GetScheme(),
		Event:      mgr.GetRecorder(v1.EventRecorderName)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func newController(mgr manager.Manager, r reconcile.Reconciler) (controller.Controller, error) {
	// Create a new controller
	c, err := controller.New("server-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return nil, err
	}

	// Watch for changes to Provider
	err = c.Watch(&source.Kind{Type: &v1.Server{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return nil, err
	}

	return c, nil
}

var _ reconcile.Reconciler = &Reconcile{}

// ReconcileProvider reconciles a Provider object
type Reconcile struct {
	client.Client
	Event      record.EventRecorder
	scheme     *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Agent object and makes changes based on the state read
// and what is in the Agent.Spec
// +kubebuilder:rbac:groups=k8s.native-lb,resources=server,verbs=get;list;watch;create;update;patch;delete
func (r *Reconcile) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Provider instance
	instance := &v1.Server{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
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
	return reconcile.Result{}, nil
}