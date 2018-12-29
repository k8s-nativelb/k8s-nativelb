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

package daemonset_controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/client-go/tools/record"

	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"github.com/k8s-nativelb/pkg/kubecli"
	"github.com/k8s-nativelb/pkg/log"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/agent"
	"github.com/k8s-nativelb/pkg/nativelb-controller/controllers/service"
)

type DaemonsetController struct {
	Controller    controller.Controller
	ReconcileNode reconcile.Reconciler
}

func NewDaemonsetController(nativelbClient kubecli.NativelbClient, agentController *agent_controller.AgentController) (*DaemonsetController, error) {
	reconcileDaemonset := newReconciler(nativelbClient, agentController)

	daemonsetInstance, err := newDaemonsetControllerController(nativelbClient, reconcileDaemonset)
	if err != nil {
		return nil, err
	}
	daemonsetController := &DaemonsetController{Controller: daemonsetInstance,
		ReconcileNode: reconcileDaemonset}

	return daemonsetController, nil

}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(nativelbClient kubecli.NativelbClient, agentController *agent_controller.AgentController) *ReconcileDaemonset {

	return &ReconcileDaemonset{NativelbClient: nativelbClient,
		scheme:          nativelbClient.GetScheme(),
		Event:           nativelbClient.GetRecorder(v1.EventRecorderName),
		agentController: agentController}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func newDaemonsetControllerController(nativelbClient kubecli.NativelbClient, r reconcile.Reconciler) (controller.Controller, error) {
	// Create a new controller
	c, err := controller.New("daemonset-controller", nativelbClient.GetManager(), controller.Options{Reconciler: r})
	if err != nil {
		return nil, err
	}

	// Watch for changes to Node
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return nil, err
	}

	return c, nil
}

var _ reconcile.Reconciler = &ReconcileDaemonset{}

// ReconcileNode reconciles a Node object
type ReconcileDaemonset struct {
	kubecli.NativelbClient
	Event             record.EventRecorder
	serviceController *service_controller.ServiceController
	scheme            *runtime.Scheme
	agentController   *agent_controller.AgentController
}

// Reconcile reads that state of the cluster for a daemonset object and makes changes based on the state
// +kubebuilder:rbac:groups=apps,resources=DaemonSet,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=Pod,verbs=get;list
func (r *ReconcileDaemonset) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	if request.Namespace != v1.ControllerNamespace {
		return reconcile.Result{}, nil
	}

	nativeClient := r.GetClient()
	// Fetch the Node instance
	daemonsetInstance := &appsv1.DaemonSet{}
	err := nativeClient.Get(context.TODO(), request.NamespacedName, daemonsetInstance)
	if err != nil && !errors.IsNotFound(err) {
		log.Log.Errorf("Fail to reconcile node error message: %v", err)

		return reconcile.Result{}, err
	}

	if daemonsetInstance.Labels == nil {
		return reconcile.Result{}, nil
	} else if _, ok := daemonsetInstance.Labels[v1.ClusterLabel]; !ok {
		return reconcile.Result{}, nil
	}

	labelSelector := labels.Set{}
	labelSelector[v1.DaemonsetLabel] = daemonsetInstance.Name
	pods := &corev1.PodList{}
	err = r.GetClient().List(context.TODO(), &client.ListOptions{LabelSelector: labelSelector.AsSelector()}, pods)
	if err != nil {
		log.Log.Reason(err).Errorf("failed to find pods for daemonset %s error %v", daemonsetInstance.Name, err)
	}

	agents, err := r.Agent().List(&client.ListOptions{LabelSelector: labelSelector.AsSelector()})
	if err != nil {
		log.Log.Reason(err).Errorf("failed to find pods for daemonset %s error %v", daemonsetInstance.Name, err)
	}

	agentMap := map[string]*v1.Agent{}
	for _, agent := range agents.Items {
		agentMap[agent.Spec.IPAddress] = &agent
	}

	for _, pod := range pods.Items {
		for i := 1; i < 5 && pod.Status.PodIP == ""; i++ {
			log.Log.Infof("pod %s doesn't have ip address yet retry", pod.Name)
			time.Sleep(2 * time.Second)
			podObject, err := r.GetKubeClient().CoreV1().Pods(v1.ControllerNamespace).Get(pod.Name, metav1.GetOptions{})
			if err != nil {
				log.Log.Reason(err).Errorf("failed to get pod %s", pod.Name)
				return reconcile.Result{}, fmt.Errorf("failed to get pod %s", pod.Name)
			}

			pod = *podObject
		}
		if pod.Status.PodIP == "" {
			return reconcile.Result{}, fmt.Errorf("failed to get ip address of the pod %s", pod.Name)
		}
		if value, ok := agentMap[pod.Status.PodIP]; !ok {
			err := r.agentController.CreateOrUpdateAgentFromPod(daemonsetInstance.Labels[v1.ClusterLabel], &pod, nil)
			if err != nil {
				log.Log.Reason(err).Errorf("failed to create agent object for pod %s error %v", pod.Name, err)
			}
		} else {
			err := r.agentController.CreateOrUpdateAgentFromPod(daemonsetInstance.Labels[v1.ClusterLabel], &pod, value)
			if err != nil {
				log.Log.Reason(err).Errorf("failed to update agent object for pod %s error %v", pod.Name, err)
			}
		}
		delete(agentMap, pod.Status.PodIP)
	}

	return reconcile.Result{}, nil
}
