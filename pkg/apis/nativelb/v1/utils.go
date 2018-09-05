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
	"google.golang.org/grpc"
	"time"

	"fmt"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

const (
	EventRecorderName   = "nativelb"
	ControllerNamespace = "nativelb"

	BackendConnectionTimeout = "2s"
	BackendIdleTimeout = "10m"
	ClientIdleTimeout = "10m"
	MaxConnections = 10000

	/*
	weight
	Select backend from discovery pool with probability based on backends weights.

	iphash
	Target backend will be calculated using hash function of client ip address mod backends count. Note if backends pool changes (for example, due discovery), client may be proxied to a different backend.

	iphash1
	(since 0.6.0) Target backend will be calculated using hash function of client ip address in a way that if some backend goes down, only clients of that backend will be proxied to another backends. If some new backends are added, behavior is tha same as iphash has.

	leastconn
	gobetween will select backends with least connections to it.

	roundrobin
	It's most simple balancing strategy, and each new connection will be proxies to next backend in the backends pool successively.

	leastbandwidth
	(since 0.3.0) Backends with least sum of rx/tx per second traffic will be selected for next request. Note that rx/tx per second values are calculated with 2 seconds interval so changes in bandwidth won't be instantly applied.
	 */
	Balance = "roundrobin"


	// UDP default configuration
	UDPMaxRequests = 0
	UDPMaxResponses = 0

	// HealthCheck default configuration
	HealthCheckFails = 1
	HealthCheckPasses = 1
	HealthCheckInterval = "2s"
	HealthCheckPingTimeoutDuration ="500ms"

	NativeLBAnnotationKey = "k8s.nativelb/cluster"
	NativeLBDefaultLabel  = "k8s.nativelb.default"

	ClusterConnectionStatusFail    = "Failed"
	ClusterConnectionStatusSuccess = "Synced"

	FarmStatusLabel        = "native-lb-farm-status"
	FarmStatusLabelSynced  = "Synced"
	FarmStatusLabelFailed  = "Failed"
	FarmStatusLabelDeleted = "Deleted"

	ServiceStatusLabel       = "native-lb-service-status"
	ServiceStatusLabelSynced = "Synced"
	ServiceStatusLabelFailed = "Failed"

	DefaultPriority = 1
	DefaultWeight = 1
)

var (
	GrpcDial    = grpc.WithInsecure()
	GrpcTimeout = 30 * time.Second
)

func configServer(port *corev1.ServicePort,isInternal bool,nodelist []string) (*Server) {
	serverSpec := ServerSpec{Bind: "",
		Protocol:strings.ToLower(fmt.Sprintf("%s",port.Protocol)),
		BackendConnectionTimeout:BackendConnectionTimeout,
		BackendIdleTimeout:BackendIdleTimeout,
		ClientIdleTimeout:ClientIdleTimeout,
		MaxConnections:MaxConnections,
		Balance:Balance,
		UDP:DefaultUdpSpec(),
		Discovery:DefaultDiscovery(port,isInternal,nodelist),HealthCheck:DefaultHealthCheck()}

	serverStatus := ServerStatus{ActiveConnections:0,RxSecond:0,RxTotal:0,TxSecond:0,TxTotal:0}

	return &Server{Spec:serverSpec,Status:serverStatus}
}


func DefaultUdpSpec() (UDP) {
	return UDP{MaxRequests:UDPMaxRequests,MaxResponses:UDPMaxResponses}
}

func DefaultHealthCheck() (HealthCheck){
	return HealthCheck{Kind: "ping",
					   Fails:HealthCheckFails,
					   Passes:HealthCheckPasses,
					   Interval:HealthCheckInterval,
	                   PingTimeoutDuration:HealthCheckPingTimeoutDuration}
}


func DefaultDiscovery(port *corev1.ServicePort,isInternal bool,backendServers []string) (Discovery) {
	return Discovery{Kind:"exec",Backends:CreateBackends(port,isInternal ,backendServers)}
}

func CreateBackends(port *corev1.ServicePort,isInternal bool,backendServers []string) ([]Backend) {
	backends := make([]Backend, len(backendServers))
	backendPort := ""

	if isInternal {
		backendPort = fmt.Sprintf("%d",port.TargetPort.IntVal)
	} else {
		backendPort = fmt.Sprintf("%d",port.NodePort)
	}

	for idx := range backendServers {
		backendSpec := BackendSpec{Host:backendServers[idx],Port:backendPort,Priority:DefaultPriority,Weight:DefaultWeight}
		backends[idx] = Backend{Spec:backendSpec,Status:DefaultBackendStatus()}
	}

	return backends
}

func DefaultBackendStatus() (BackendStatus) {
	return BackendStatus{TxSecond:0,RxSecond:0,ActiveConnections:0,Live:false,RefusedConnections:0,Rx:0,TotalConnections:0,Tx:0}
}