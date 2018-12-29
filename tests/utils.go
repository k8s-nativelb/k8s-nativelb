package tests

import (
	"fmt"
	"time"

	nativelb "github.com/k8s-nativelb/pkg/apis/nativelb/v1"

	"k8s.io/apimachinery/pkg/api/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega"
)

const (
	TestNamespace       = "nativelb-tests-namespace"
	NginxDeploymentName = "nginx-app"
	ExternalClusterName = "cluster-sample-external"
)

var (
	SelectorLabel = map[string]string{"app": "nginx"}
)

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func CreateMockAgent(testClient *TestClient,agentName,cluster string,port int32) (*nativelb.Agent,*corev1.Pod, error) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name:agentName,Namespace:nativelb.ControllerNamespace},
		Spec:corev1.PodSpec{Containers: []corev1.Container{{Name:"mock-agent",
			Env: []corev1.EnvVar{{Name:"CONTROL_IP",ValueFrom:&corev1.EnvVarSource{FieldRef:&corev1.ObjectFieldSelector{FieldPath:"status.podIP"}}},
								 {Name:"CONTROL_PORT",Value:fmt.Sprintf("%d",port)},
			                     {Name:"CLUSTER_NAME",Value:cluster}},
			                     Image:"registry:5000/k8s-nativelb/nativelb-mockagent:latest",
			ImagePullPolicy:"IfNotPresent"}}}}
	var err error
	pod, err = testClient.KubeClient.CoreV1().Pods(nativelb.ControllerNamespace).Create(pod)
	if err != nil {
		return nil,nil,err
	}

	Eventually(func() bool{
		pod, err = testClient.KubeClient.CoreV1().Pods(nativelb.ControllerNamespace).Get(pod.Name,metav1.GetOptions{})
		if err != nil {
			return false
		}

		if pod.Status.PodIP != "" {
			return true
		}

		return false
	}, 15*time.Second, 5*time.Second).Should(Equal(true))

	agentSpec := nativelb.AgentSpec{Port:port,Cluster:cluster,HostName:agentName,IPAddress:pod.Status.PodIP}
	agent := &nativelb.Agent{ObjectMeta: metav1.ObjectMeta{Namespace:nativelb.ControllerNamespace,Name:agentName},
							Spec:agentSpec}

	agent, err = testClient.NativelbClient.Agent().Create(agent)
	if err != nil {
		return nil,nil,err
	}

	return agent,pod,nil
}

func DeleteMockAgent(testClient *TestClient,agentName,podName string) (error) {
	err := testClient.KubeClient.CoreV1().Pods(nativelb.ControllerNamespace).Delete(podName,&metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = testClient.NativelbClient.Agent().Delete(agentName)
	if err != nil {
		return err
	}

	Eventually(func() bool{
		_,err := testClient.KubeClient.CoreV1().Pods(nativelb.ControllerNamespace).Get(podName,metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			return true
		}

		return false

	}, 15*time.Second, 5*time.Second).Should(Equal(true))

	return nil
}

func CreateCluster(testClient *TestClient,clusterName ,ipRange string,isInternal bool) (*nativelb.Cluster, error) {
	cluster := &nativelb.Cluster{ObjectMeta:metav1.ObjectMeta{Name:clusterName,Namespace:nativelb.ControllerNamespace},
								 Spec:nativelb.ClusterSpec{Default:false,Internal:isInternal,IpRange:ipRange}}


	return testClient.NativelbClient.Cluster().Create(cluster)
}

func DeleteCluster(testClient *TestClient,clusterName string) (error) {
	return testClient.NativelbClient.Cluster().Delete(clusterName)
}

func CreateNginxDeployment() *appsv1.Deployment {
	replicas := int32(1)
	return &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: NginxDeploymentName, Namespace: TestNamespace, Labels: SelectorLabel},
		Spec: appsv1.DeploymentSpec{Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: SelectorLabel},
			Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: SelectorLabel},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "nginx", Image: "docker.io/nginx:latest"}}}}}}
}

func WaitForDeploymentToBeReady(testClient *TestClient, deploymentObject *appsv1.Deployment) error {
	for i := 0; i < 10; i++ {
		deploymentObject, err := testClient.KubeClient.AppsV1().Deployments(TestNamespace).Get(deploymentObject.Name, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		if *deploymentObject.Spec.Replicas == deploymentObject.Status.AvailableReplicas {
			return nil
		}

		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("deployment not ready")
}

func WaitForClusterIpService(testClient *TestClient, serviceObject *corev1.Service) error {
	for i := 0; i < 10; i++ {
		_, err := testClient.KubeClient.CoreV1().Services(TestNamespace).Get(serviceObject.Name, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		if err == nil {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("service not created")
}

func WaitForServiceToBySynced(testClient *TestClient, serviceObject *corev1.Service) error {
	for i := 0; i < 10; i++ {
		serviceObject, err := testClient.KubeClient.CoreV1().Services(TestNamespace).Get(serviceObject.Name, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		if err == nil && serviceObject.Labels[nativelb.ServiceStatusLabel] == nativelb.ServiceStatusLabelSynced {
			return nil
		}

		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("service don't have sync label")
}
func DeleteNginxDeployment(testClient *TestClient, deploymentObject *appsv1.Deployment) error {
	err := testClient.KubeClient.AppsV1().Deployments(TestNamespace).Delete(deploymentObject.Name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		deploymentObject, err = testClient.KubeClient.AppsV1().Deployments(TestNamespace).Get(deploymentObject.Name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}

		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("Fail to remove nginx deployment")
}

func FarmName(serviceName string) string {
	return fmt.Sprintf("%s-%s", TestNamespace, serviceName)
}
