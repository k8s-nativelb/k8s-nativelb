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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	EventRecorderName   = "nativelb"
	ControllerNamespace = "nativelb"
	ClusterLabel        = "k8s.nativelb.io/cluster"

	NativeLBAnnotationKey = "k8s.nativelb.io/cluster"
	NativeLBDefaultLabel  = "k8s.nativelb.io/default"
	DaemonsetLabel        = "k8s.nativelb.io/daemonset"

	DaemonsetClusterLabel = "daemonset.nativelb.io/cluster"
	AgentPodPortLabel     = "daemonset.nativelb.io/port"

	ClusterConnectionStatusFail    = "Failed"
	ClusterConnectionStatusSuccess = "Synced"
	ClusterConnectionStatusPartial = "Partial"
	ClusterTypeNativeAgent         = "NativeAgent"
	ClusterTypeCustom              = "Custom"

	// TODO: need this is unused need to add this to farm labels
	FarmStatusLabel        = "native-lb-farm-status"
	FarmStatusLabelSynced  = "Synced"
	FarmStatusLabelSyncing = "Syncing"
	FarmStatusLabelFailed  = "Failed"
	FarmStatusLabelDeleted = "Deleted"

	ServiceStatusLabel        = "native-lb-service-status"
	ServiceStatusLabelSynced  = "Synced"
	ServiceStatusLabelSyncing = "Syncing"
	ServiceStatusLabelFailed  = "Failed"

	AgentAliveStatus  = "Alive"
	AgentDownStatus   = "Down"
	AgentUnknowStatus = "Unknow"

	NativeLBServerRef = "k8s.nativelb.server"
	NativeLBFarmRef   = "k8s.nativelb.farm"

	// ---------------------------------------------------------------------

	KeepaliveTime              = 10
	ResyncServicesInterval     = 10
	ResyncFailFarms            = 30
	ResyncCleanRemovedServices = 320

	BackendConnectionTimeout = "2s"
	BackendIdleTimeout       = "10m"
	ClientIdleTimeout        = "10m"
	MaxConnections           = 10000

	Balance = "roundrobin"

	// UDP default configuration
	UDPMaxRequests  = 0
	UDPMaxResponses = 0

	// HealthCheck default configuration
	HealthCheckFails               = 1
	HealthCheckPasses              = 1
	HealthCheckInterval            = "2s"
	HealthCheckPingTimeoutDuration = "500ms"

	DefaultPriority = 1
	DefaultWeight   = 1
)

var (
	GrpcDial    = grpc.WithInsecure()
	GrpcTimeout = 30 * time.Second
)

func configServer(servicePort *corev1.ServicePort, isInternal bool, ipAddr string, discovery Discovery, serverName string, farmName string) *Server {
	labelMap := make(map[string]string)
	labelMap[NativeLBFarmRef] = farmName
	var port int32
	if isInternal {
		port = servicePort.Port
	} else {
		port = servicePort.NodePort
	}

	serverSpec := ServerSpec{Bind: ipAddr,
		Port:                     port,
		Protocol:                 strings.ToLower(fmt.Sprintf("%s", servicePort.Protocol)),
		BackendConnectionTimeout: BackendConnectionTimeout,
		BackendIdleTimeout:       BackendIdleTimeout,
		ClientIdleTimeout:        ClientIdleTimeout,
		MaxConnections:           MaxConnections,
		Balance:                  Balance,
		UDP:                      DefaultUdpSpec(),
		Discovery:                discovery, HealthCheck: DefaultHealthCheck()}

	serverStatus := ServerStatus{ActiveConnections: 0, RxSecond: 0, RxTotal: 0, TxSecond: 0, TxTotal: 0}

	return &Server{ObjectMeta: metav1.ObjectMeta{Labels: labelMap, Name: serverName, Namespace: ControllerNamespace}, Spec: serverSpec, Status: serverStatus}
}

func DefaultUdpSpec() UDP {
	return UDP{MaxRequests: UDPMaxRequests, MaxResponses: UDPMaxResponses}
}

func DefaultHealthCheck() HealthCheck {
	return HealthCheck{Kind: "ping",
		Fails:               HealthCheckFails,
		Passes:              HealthCheckPasses,
		Interval:            HealthCheckInterval,
		PingTimeoutDuration: HealthCheckPingTimeoutDuration}
}

func DefaultDiscovery(backendServers []BackendSpec) Discovery {
	return Discovery{Kind: "exec", Backends: backendServers}
}

func CreateBackends(port *corev1.ServicePort, isInternal bool, backendServers []string, serverName string) ([]Backend, []BackendSpec) {
	backends := make([]Backend, len(backendServers))
	backendsSpec := make([]BackendSpec, len(backendServers))
	var backendPort int32

	if isInternal {
		backendPort = port.TargetPort.IntVal
	} else {
		backendPort = port.NodePort
	}

	for idx := range backendServers {
		backendSpec := BackendSpec{Host: backendServers[idx], Port: backendPort, Priority: DefaultPriority, Weight: DefaultWeight}
		backends[idx] = Backend{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{NativeLBServerRef: serverName}, Name: fmt.Sprintf("%s-%s", serverName, backendServers[idx]), Namespace: ControllerNamespace}, Spec: backendSpec, Status: DefaultBackendStatus()}
		backendsSpec[idx] = backendSpec
	}

	return backends, backendsSpec
}

func DefaultBackendStatus() BackendStatus {
	return BackendStatus{TxSecond: 0, RxSecond: 0, ActiveConnections: 0, Live: false, RefusedConnections: 0, Rx: 0, TotalConnections: 0, Tx: 0}
}
