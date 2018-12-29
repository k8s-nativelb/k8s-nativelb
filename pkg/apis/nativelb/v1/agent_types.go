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
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// +k8s:openapi-gen=true
type AgentSpec struct {
	HostName  string `json:"hostName"`
	IPAddress string `json:"ipAddress"`
	Port      int32  `json:"port"`
	Cluster   string `json:"cluster"`
}

// +k8s:openapi-gen=true
type AgentStatus struct {
	LBPid            int           `json:"lbPid,omitempty"`
	KeepAlivedPid    int           `json:"KeepAlivedPid,omitempty"`
	StartTime        string        `json:"startTime,omitempty"`
	Time             string        `json:"time,omitempty"`
	Uptime           time.Duration `json:"uptime,omitempty"`
	Version          string        `json:"version,omitempty"`
	ConnectionStatus string        `json:"connectionStatus,omitempty"`
	LastUpdate       metav1.Time   `json:"lastUpdate,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Agent is the Schema for the Agents API
// +k8s:openapi-gen=true
type Agent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentSpec   `json:"spec"`
	Status AgentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AgentList contains a list of Agent
type AgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Agent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Agent{}, &AgentList{})
}

func (a *Agent) GetUrl() string {
	return fmt.Sprintf("%s:%d", a.Spec.IPAddress, a.Spec.Port)
}
