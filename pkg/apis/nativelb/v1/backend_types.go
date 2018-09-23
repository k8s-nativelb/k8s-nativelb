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

// backendSpec defines the desired state of backend
// +k8s:openapi-gen=true
type BackendSpec struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Priority int    `json:"priority,omitempty"`
	Weight   int    `json:"weight,omitempty"`
}

// backendStatus defines the observed state of backend
// +k8s:openapi-gen=true
type BackendStatus struct {
	Live               bool `json:"live"`
	TotalConnections   int  `json:"totalConnections"`
	ActiveConnections  int  `json:"activeConnections"`
	RefusedConnections int  `json:"refusedConnections"`
	Rx                 int  `json:"rx"`
	Tx                 int  `json:"tx"`
	RxSecond           int  `json:"rxSecond"`
	TxSecond           int  `json:"txSecond"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// backend is the Schema for the backends API
// +k8s:openapi-gen=true
type Backend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackendSpec   `json:"spec"`
	Status BackendStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// backendList contains a list of backend
type BackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backend `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Backend{}, &BackendList{})
}
