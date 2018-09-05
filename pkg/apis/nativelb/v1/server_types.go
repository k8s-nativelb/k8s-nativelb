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

// ServerSpec defines the desired state of Server
// +k8s:openapi-gen=true
type ServerSpec struct {
	Bind                     string      `json:"bind"`
	Protocol                 string      `json:"protocol"`
	UDP                      UDP         `json:"udp,omitempty"`
	Balance                  string      `json:"balance"`
	MaxConnections           int      `json:"maxConnections"`
	ClientIdleTimeout        string      `json:"clientIdleTimeout"`
	BackendIdleTimeout       string      `json:"backendIdleTimeout"`
	BackendConnectionTimeout string      `json:"backendConnectionTimeout"`
	Discovery                Discovery   `json:"discovery"`
	HealthCheck              HealthCheck `json:"healthCheck,omitempty"`
}

// ServerStatus defines the observed state of Server
// +k8s:openapi-gen=true
type ServerStatus struct {
	ActiveConnections int       `json:"activeConnections,omitempty"`
	RxTotal           int       `json:"rxTotal,omitempty"`
	TxTotal           int       `json:"txTotal,omitempty"`
	RxSecond          int       `json:"rxSecond,omitempty"`
	TxSecond          int       `json:"txSecond,omitempty"`
}

// +k8s:openapi-gen=true
type Discovery struct {
	Kind        string        `json:"kind"`
	Backends []Backend `json:"backend"`
}

// +k8s:openapi-gen=true
type HealthCheck struct {
	Fails               int    `json:"fails,omitempty"`
	Passes              int    `json:"passes,omitempty"`
	Interval            string `json:"interval,omitempty"`
	Timeout             string `json:"timeout,omitempty"`
	Kind                string `json:"kind,omitempty"`
	PingTimeoutDuration string `json:"pingTimeoutDuration,omitempty"`
}

// +k8s:openapi-gen=true
type UDP struct {
	MaxRequests  int `json:"maxRequests,omitempty"`
	MaxResponses int `json:"maxResponses,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Server is the Schema for the Servers API
// +k8s:openapi-gen=true
type Server struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerSpec   `json:"spec"`
	Status ServerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServerList contains a list of Server
type ServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Server `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Server{}, &ServerList{})
}
