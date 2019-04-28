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

	AgentAliveStatus              = "Alive"
	AgentDownStatus               = "Down"
	AgentUnknowStatus             = "Unknow"
	AgentOperetionalStatusActive  = "Active"
	AgentOperetionalStatusDisable = "Disable"

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

func configServer(servicePort *corev1.ServicePort, isInternal bool, ipAddr string, serverName string, farmName string, backendsSpec map[string]*BackendSpec) *Server {
	labelMap := make(map[string]string)
	labelMap[NativeLBFarmRef] = farmName

	healthCheck := &HealthCheck{Kind: "ping",
		Fails:               HealthCheckFails,
		Passes:              HealthCheckPasses,
		Interval:            HealthCheckInterval,
		PingTimeoutDuration: HealthCheckPingTimeoutDuration}

	tcp := &TCP{BackendConnectionTimeout: BackendConnectionTimeout,
		BackendIdleTimeout: BackendIdleTimeout,
		ClientIdleTimeout:  ClientIdleTimeout,
		MaxConnections:     MaxConnections}

	udp := &UDP{MaxRequests: UDPMaxRequests, MaxResponses: UDPMaxResponses}

	serverSpec := ServerSpec{Bind: ipAddr,
		Port:        servicePort.Port,
		Protocol:    strings.ToLower(fmt.Sprintf("%s", servicePort.Protocol)),
		Balance:     Balance,
		UDP:         udp,
		TCP:         tcp,
		HealthCheck: healthCheck,
		Backends:    backendsSpec}

	serverStatus := ServerStatus{}

	return &Server{ObjectMeta: metav1.ObjectMeta{Labels: labelMap, Name: serverName, Namespace: ControllerNamespace}, Spec: serverSpec, Status: serverStatus}
}

func CreateBackendsSpec(port *corev1.ServicePort, isInternal bool, endpoints []string) map[string]*BackendSpec {
	backendsSpec := make(map[string]*BackendSpec)
	var backendPort int32

	if isInternal {
		backendPort = port.TargetPort.IntVal
	} else {
		backendPort = port.NodePort
	}

	for _, host := range endpoints {
		backendName := fmt.Sprintf("%s-%s-%d", host, port.Protocol, backendPort)
		backendSpec := &BackendSpec{Host: host, Port: backendPort, Priority: DefaultPriority, Weight: DefaultWeight}
		backendsSpec[backendName] = backendSpec
	}

	return backendsSpec
}

func DefaultBackendStatus() BackendStatus {
	return BackendStatus{}
}
