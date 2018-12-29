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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster
// +k8s:openapi-gen=true
type ClusterSpec struct {
	IpRange  string `json:"ipRange"`
	Default  bool   `json:"default,omitempty"`
	Internal bool   `json:"internal,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
// +k8s:openapi-gen=true
type ClusterStatus struct {
	Agents              map[string]*Agent              `json:"agents,omitempty"`
	AllocatedIps        map[string]string              `json:"allocatedIps,omitempty"`
	AllocatedNamespaces map[string]*AllocatedNamespace `json:"AllocatedNamespaces,omitempty"`
	ConnectionStatus    string                         `json:"connectionStatus"`
	LastUpdate          metav1.Time                    `json:"lastUpdate"`
}

type AllocatedNamespace struct {
	RouterID int32    `json:"routerID"`
	Farms    []string `json:"farms"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cluster is the Schema for the clusters API
// +k8s:openapi-gen=true
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
