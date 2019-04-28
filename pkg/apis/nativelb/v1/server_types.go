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
	Bind        string                  `json:"bind"`
	Port        int32                   `json:"port"`
	Protocol    string                  `json:"protocol"`
	UDP         *UDP                    `json:"udp,omitempty"`
	TCP         *TCP                    `json:"tcp"`
	Balance     string                  `json:"balance"`
	Backends    map[string]*BackendSpec `json:"backend"`
	HealthCheck *HealthCheck            `json:"healthCheck,omitempty"`
}

// ServerStatus defines the observed state of Server
// +k8s:openapi-gen=true
type ServerStatus struct {
	Frontend *HaproxyStatus   `json:"frontEnd,omitempty"`
	Backend  *HaproxyStatus   `json:"backEnd,omitempty"`
	Backends []*HaproxyStatus `json:"backends,omitempty"`
}

// +k8s:openapi-gen=true
type HealthCheck struct {
	Fails               int32  `json:"fails,omitempty"`
	Passes              int32  `json:"passes,omitempty"`
	Interval            string `json:"interval,omitempty"`
	Timeout             string `json:"timeout,omitempty"`
	Kind                string `json:"kind,omitempty"`
	PingTimeoutDuration string `json:"pingTimeoutDuration,omitempty"`
}

// +k8s:openapi-gen=true
type UDP struct {
	MaxRequests  int32 `json:"maxRequests,omitempty"`
	MaxResponses int32 `json:"maxResponses,omitempty"`
}

type TCP struct {
	MaxConnections           int32  `json:"maxConnections"`
	ClientIdleTimeout        string `json:"clientIdleTimeout"`
	BackendIdleTimeout       string `json:"backendIdleTimeout"`
	BackendConnectionTimeout string `json:"backendConnectionTimeout"`
}

// +k8s:openapi-gen=true
type BackendSpec struct {
	Host     string `json:"host"`
	Port     int32  `json:"port"`
	Priority int32  `json:"priority,omitempty"`
	Weight   int32  `json:"weight,omitempty"`
}

// backendStatus defines the observed state of backend
// +k8s:openapi-gen=true
type BackendStatus struct {
	HaproxyStatus
}

type HaproxyStatus struct {
	PxName   string `json:"pxname,omitempty"`
	SvName   string `json:"svname,omitempty"`
	Qcur     uint64 `json:"qcur,omitempty"`
	Qmax     uint64 `json:"qmax,omitempty"`
	Scur     uint64 `json:"scur,omitempty"`
	Smax     uint64 `json:"smax,omitempty"`
	Slim     uint64 `json:"slim,omitempty"`
	Stot     uint64 `json:"stot,omitempty"`
	Bin      uint64 `json:"bin,omitempty"`
	Bout     uint64 `json:"bout,omitempty"`
	Dreq     uint64 `json:"dreq,omitempty"`
	Dresp    uint64 `json:"dresp,omitempty"`
	Ereq     uint64 `json:"ereq,omitempty"`
	Econ     uint64 `json:"econ,omitempty"`
	Eresp    uint64 `json:"eresp,omitempty"`
	Wretr    uint64 `json:"wretr,omitempty"`
	Wredis   uint64 `json:"wredis,omitempty"`
	Status   string `json:"status,omitempty"`
	Weight   uint64 `json:"weight,omitempty"`
	Act      uint64 `json:"act,omitempty"`
	Bck      uint64 `json:"bck,omitempty"`
	ChkFail  uint64 `json:"chkfail,omitempty"`
	ChkDown  uint64 `json:"chkdown,omitempty"`
	Lastchg  uint64 `json:"lastchg,omitempty"`
	Downtime uint64 `json:"downtime,omitempty"`
	Qlimit   uint64 `json:"qlimit,omitempty"`
	Pid      uint64 `json:"pid,omitempty"`
	Iid      uint64 `json:"iid,omitempty"`
	Sid      uint64 `json:"sid,omitempty"`
	Throttle uint64 `json:"throttle,omitempty"`
	Lbtot    uint64 `json:"lbtot,omitempty"`
	Tracked  uint64 `json:"tracked,omitempty"`
	Type     uint64 `json:"type,omitempty"`
	Rate     uint64 `json:"rate,omitempty"`
	RateLim  uint64 `json:"rateLim,omitempty"`
	RateMax  uint64 `json:"rateMax,omitempty"`

	//UNK     -> unknown
	//INI     -> initializing
	//SOCKERR -> socket error
	//L4OK    -> check passed on layer 4, no upper layers testing enabled
	//L4TOUT  -> layer 1-4 timeout
	//L4CON   -> layer 1-4 connection problem, for example
	//"Connection refused" (tcp rst) or "No route to host" (icmp)
	//L6OK    -> check passed on layer 6
	//L6TOUT  -> layer 6 (SSL) timeout
	//L6RSP   -> layer 6 invalid response - protocol error
	//L7OK    -> check passed on layer 7
	//L7OKC   -> check conditionally passed on layer 7, for example 404 with
	//disable-on-404
	//L7TOUT  -> layer 7 (HTTP/SMTP) timeout
	//L7RSP   -> layer 7 invalid response - protocol error
	//L7STS   -> layer 7 response error, for example HTTP 5xx
	CheckStatus string `json:"checkStatus,omitempty"`

	CheckCode     uint64 `json:"check_code,omitempty"`
	CheckDuration uint64 `json:"check_duration,omitempty"`
	Hrsp1xx       uint64 `json:"hrsp1xx,omitempty"`
	Hrsp2xx       uint64 `json:"hrsp2xx,omitempty"`
	Hrsp3xx       uint64 `json:"hrsp3xx,omitempty"`
	Hrsp4xx       uint64 `json:"hrsp4xx,omitempty"`
	Hrsp5xx       uint64 `json:"hrsp5xx,omitempty"`
	HrspOther     uint64 `json:"hrspOther,omitempty"`
	Hanafail      uint64 `json:"hanafail,omitempty"`
	ReqRate       uint64 `json:"reqRate,omitempty"`
	ReqRateMax    uint64 `json:"reqRate_max,omitempty"`
	ReqTot        uint64 `json:"reqTot,omitempty"`
	CliAbrt       uint64 `json:"cliAbrt,omitempty"`
	SrvAbrt       uint64 `json:"srvAbrt,omitempty"`
	CompIn        uint64 `json:"compIn,omitempty"`
	CompOut       uint64 `json:"compOut,omitempty"`
	CompByp       uint64 `json:"compByp,omitempty"`
	CompRsp       uint64 `json:"compRsp,omitempty"`
	LastSess      int64  `json:"lastsess,omitempty"`
	LastChk       string `json:"lastChk,omitempty"`
	LastAgt       uint64 `json:"lastAgt,omitempty"`
	Qtime         uint64 `json:"qtime,omitempty"`
	Ctime         uint64 `json:"ctime,omitempty"`
	Rtime         uint64 `json:"rtime,omitempty"`
	Ttime         uint64 `json:"ttime,omitempty"`
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
