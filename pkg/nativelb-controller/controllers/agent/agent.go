package agent_controller

import (
	"fmt"
	"github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a AgentController) CreateOrUpdateAgentFromPod(clusterName string, pod *corev1.Pod, agent *v1.Agent) error {
	if _, ok := pod.Labels[v1.AgentPodPortLabel]; !ok {
		return fmt.Errorf("failed to find %s label on the pod", v1.AgentPodPortLabel)
	}

	pod.Labels[v1.ClusterLabel] = clusterName
	port, err := strconv.Atoi(pod.Labels[v1.AgentPodPortLabel])
	if err != nil {
		return fmt.Errorf("failed to convert port %s to string error %v", pod.Labels[v1.AgentPodPortLabel], err)
	}

	ownerRef := []metav1.OwnerReference{{Name: pod.Name, Kind: "Pod", UID: pod.UID, APIVersion: "v1"}}

	agentSpec := v1.AgentSpec{IPAddress: pod.Status.PodIP, Cluster: clusterName, HostName: pod.Name, Port: int32(port)}
	agentStatus := v1.AgentStatus{LastUpdate: metav1.Now(), ConnectionStatus: v1.AgentUnknowStatus}
	agentObject := &v1.Agent{ObjectMeta: metav1.ObjectMeta{Name: pod.Name, Namespace: v1.ControllerNamespace, Labels: pod.Labels, OwnerReferences: ownerRef}, Spec: agentSpec, Status: agentStatus}
	if agent == nil {
		_, err = a.Reconcile.Agent().Create(agentObject)
	} else {
		agentObject.Status = agent.Status
		_, err = a.Reconcile.Agent().Update(agentObject)
	}
	return err
}
