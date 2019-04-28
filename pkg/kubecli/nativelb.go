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
package kubecli

//go:generate mockgen -source $GOFILE -package=$GOPACKAGE -destination=generated_mock_$GOFILE

import (
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type NativelbClient interface {
	Cluster(string) ClusterInterface
	Farm(string) FarmInterface
	Server(string) ServerInterface
	Agent(string) AgentInterface
	GetManager() manager.Manager
	GetClient() client.Client
	GetScheme() *runtime.Scheme
	GetRecorder(name string) record.EventRecorder
	GetKubeClient() *kubernetes.Clientset
}

type nativeLB struct {
	Manager    manager.Manager
	Client     client.Client
	KubeConfig *kubernetes.Clientset
}

func (n *nativeLB) GetClient() client.Client {
	return n.Client
}

func (n *nativeLB) GetManager() manager.Manager {
	return n.Manager
}

func (n *nativeLB) GetScheme() *runtime.Scheme {
	return n.Manager.GetScheme()
}

func (n *nativeLB) GetRecorder(name string) record.EventRecorder {
	return n.Manager.GetRecorder(name)
}

func (n *nativeLB) GetKubeClient() *kubernetes.Clientset {
	return n.KubeConfig
}

type ClusterInterface interface {
	Get(name string) (*v1.Cluster, error)
	List(opts *client.ListOptions) (*v1.ClusterList, error)
	Create(instance *v1.Cluster) (*v1.Cluster, error)
	Update(*v1.Cluster) (*v1.Cluster, error)
	Delete(name string) error
}

type FarmInterface interface {
	Get(name string) (*v1.Farm, error)
	List(opts *client.ListOptions) (*v1.FarmList, error)
	Create(instance *v1.Farm) (*v1.Farm, error)
	Update(*v1.Farm) (*v1.Farm, error)
	Delete(name string) error
}

type ServerInterface interface {
	Get(name string) (*v1.Server, error)
	List(opts *client.ListOptions) (*v1.ServerList, error)
	Create(instance *v1.Server) (*v1.Server, error)
	Update(*v1.Server) (*v1.Server, error)
	Delete(name string) error
}

type AgentInterface interface {
	Get(name string) (*v1.Agent, error)
	List(opts *client.ListOptions) (*v1.AgentList, error)
	Create(instance *v1.Agent) (*v1.Agent, error)
	Update(*v1.Agent) (*v1.Agent, error)
	Delete(name string) error
}
