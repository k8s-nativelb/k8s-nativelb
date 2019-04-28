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
)

// +k8s:openapi-gen=true
type AgentSpec struct {
	HostName    string `json:"hostName"`
	IPAddress   string `json:"ipAddress"`
	Port        int32  `json:"port"`
	Cluster     string `json:"cluster"`
	Operational bool   `json:"operational"`
}

// +k8s:openapi-gen=true
type AgentStatus struct {
	LoadBalancer     *LoadBalancer `json:"loadBalancer,omitempty"`
	KeepAlivedPid    int32         `json:"keepalivedPid,omitempty"`
	HaproxyPid       int32         `json:"haproxyPid,omitempty"`
	NginxPid         int32         `json:"nginxPid,omitempty"`
	Version          string        `json:"version,omitempty"`
	ConnectionStatus string        `json:"connectionStatus,omitempty"`
	OperationStatus  string        `json:"operationStatus,omitempty"`
	LastUpdate       metav1.Time   `json:"lastUpdate,omitempty"`
}

// +k8s:openapi-gen=true
type LoadBalancer struct {
	Haproxy    *Haproxy    `json:"haproxy,omitempty"`
	Nginx      *Nginx      `json:"nginx,omitempty"`
	Keepalived *Keepalived `json:"keepalived,omitempty"`
}

type Keepalived struct {
	Pid             uint64            `json:"pid"`
	InstancesStatus map[string]string `json:"instancesStatus"`
}

// +k8s:openapi-gen=true
type Haproxy struct {
	Version                    string `json:"version"`
	ReleaseDate                string `json:"releaseDate"`
	Nbproc                     uint64 `json:"nbproc"`
	ProcessNum                 uint64 `json:"process_num"`
	Pid                        uint64 `json:"pid"`
	Uptime                     string `json:"uptime"`
	UptimeSec                  uint64 `json:"uptime_sec"`
	MemMaxMB                   uint64 `json:"memmax_MB"`
	UlimitN                    uint64 `json:"ulimit-n"`
	Maxsock                    uint64 `json:"maxsock"`
	Maxconn                    uint64 `json:"maxconn"`
	HardMaxconn                uint64 `json:"hardMaxconn"`
	CurrConns                  uint64 `json:"currConns"`
	CumConns                   uint64 `json:"cumConns"`
	CumReq                     uint64 `json:"cumReq"`
	MaxSslConns                uint64 `json:"maxSslConns"`
	CurrSslConns               uint64 `json:"currSslConns"`
	CumSslConns                uint64 `json:"cumSslConns"`
	Maxpipes                   uint64 `json:"maxpipes"`
	PipesUsed                  uint64 `json:"pipesUsed"`
	PipesFree                  uint64 `json:"pipesFree"`
	ConnRate                   uint64 `json:"connRate"`
	ConnRateLimit              uint64 `json:"connRateLimit"`
	MaxConnRate                uint64 `json:"maxConnRate"`
	SessRate                   uint64 `json:"sessRate"`
	SessRateLimit              uint64 `json:"sessRateLimit"`
	MaxSessRate                uint64 `json:"maxSessRate"`
	SslRate                    uint64 `json:"sslRate"`
	SslRateLimit               uint64 `json:"sslRateLimit"`
	MaxSslRate                 uint64 `json:"maxSslRate"`
	SslFrontendKeyRate         uint64 `json:"sslFrontendKeyRate"`
	SslFrontendMaxKeyRate      uint64 `json:"sslFrontendMaxKeyRate"`
	SslFrontendSessionReusePct uint64 `json:"sslFrontendSessionReuse_pct"`
	SslBackendKeyRate          uint64 `json:"sslBackendKeyRate"`
	SslBackendMaxKeyRate       uint64 `json:"sslBackendMaxKeyRate"`
	SslCacheLookups            uint64 `json:"sslCacheLookups"`
	SslCacheMisses             uint64 `json:"sslCacheMisses"`
	CompressBpsIn              uint64 `json:"compressBpsIn"`
	CompressBpsOut             uint64 `json:"compressBpsOut"`
	CompressBpsRateLim         uint64 `json:"compressBpsRateLim"`
	Tasks                      uint64 `json:"tasks"`
	RunQueue                   uint64 `json:"run_queue"`
	IdlePct                    uint64 `json:"idle_pct"`
}

type Nginx struct {
	Pid               uint64 `json:"Pid"`
	ActiveConnections uint64 `json:"activeConnections"`
	Reading           uint64 `json:"reading"`
	Writing           uint64 `json:"writing"`
	Waiting           uint64 `json:"waiting"`
	Version           string `json:"version"`
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
