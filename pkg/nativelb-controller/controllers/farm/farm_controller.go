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

package farm_controller

import (
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"time"

	//pb "github.com/k8s-nativelb/pkg/proto"
	"github.com/k8s-nativelb/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	"context"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/cluster"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/server"
)

type FarmController struct {
	Controller        controller.Controller
	Reconcile *Reconcile
	serverController *server_controller.ServerController
	clusterController *cluster_controller.ClusterController
}

func NewFarmController(mgr manager.Manager,serverController *server_controller.ServerController,clusterController *cluster_controller.ClusterController) (*FarmController, error) {
	reconcileInstance := newReconciler(mgr)
	controllerInstance, err := newController(mgr, reconcileInstance)
	if err != nil {
		return nil, err
	}

	farmController := &FarmController{Controller: controllerInstance,
		Reconcile: reconcileInstance,serverController:serverController,clusterController:clusterController}

	return farmController, nil
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
	c, err := controller.New("farm-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return nil, err
	}

	// Watch for changes to Provider
	err = c.Watch(&source.Kind{Type: &v1.Farm{}}, &handler.EnqueueRequestForObject{})
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
// +kubebuilder:rbac:groups=k8s.native-lb,resources=farm,verbs=get;list;watch;create;update;patch;delete
func (r *Reconcile) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Provider instance
	instance := &v1.Farm{}
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

func (r *Reconcile) GetFarm(farmName string) (*v1.Farm, error) {
	farm := &v1.Farm{}
	retry := 5
	var err error

	for i := 0;i < retry; i++ {
		err = r.Get(context.TODO(), client.ObjectKey{Namespace: v1.ControllerNamespace, Name: farmName}, farm)
		if err != nil && !errors.IsNotFound(err) {
			return nil,err
		} else if err == nil {
			return farm, nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return nil, err
}