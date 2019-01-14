package tests

import (
	"bytes"
	"fmt"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"time"

	nativelb "github.com/k8s-nativelb/pkg/apis/nativelb/v1"

	"k8s.io/apimachinery/pkg/api/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega"
)

var (
	ClientPod = &corev1.Pod{}
	InClusterAgentLabel = map[string]string{nativelb.ClusterLabel:"cluster-sample-cluster"}
)

const (
	TestNamespace       = "nativelb-tests-namespace"
	SampleClusterName = "cluster-sample-cluster"
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
								 Spec:nativelb.ClusterSpec{Default:false,Internal:isInternal,Subnet:ipRange}}


	return testClient.NativelbClient.Cluster().Create(cluster)
}

func DeleteCluster(testClient *TestClient,clusterName string) (error) {
	return testClient.NativelbClient.Cluster().Delete(clusterName)
}

func CreateNginxDeployment(deploymentName, port string,selectorLabel map[string]string) *appsv1.Deployment {
	replicas := int32(1)
	return &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: TestNamespace, Labels: selectorLabel},
		Spec: appsv1.DeploymentSpec{Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: selectorLabel},
			Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: selectorLabel},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "nginx", Image: "quay.io/k8s-nativelb/nativelb-nginx:latest",Command: []string{"nginx"},Env: []corev1.EnvVar{{Name:"NGINX_PORT",Value:port}}}}}}}}
}

func CreateUdpServerDeployment(deploymentName, port string,selectorLabel map[string]string) *appsv1.Deployment {
	replicas := int32(1)
	return &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: TestNamespace, Labels: selectorLabel},
		Spec: appsv1.DeploymentSpec{Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: selectorLabel},
			Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: selectorLabel},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "nginx", Image: "quay.io/k8s-nativelb/nativelb-nginx:latest",Command: []string{"/server", port},Env: []corev1.EnvVar{{Name:"NGINX_PORT",Value:port}}}}}}}}
}

func WaitForDeploymentToBeReady(testClient *TestClient, deploymentObject *appsv1.Deployment) {
	for i := 0; i < 10; i++ {
		deploymentObject, err := testClient.KubeClient.AppsV1().Deployments(TestNamespace).Get(deploymentObject.Name, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			Expect(err).ToNot(HaveOccurred())
		}

		if err == nil && *deploymentObject.Spec.Replicas == deploymentObject.Status.AvailableReplicas {
			return
		}

		time.Sleep(6 * time.Second)
	}

	Expect(fmt.Errorf("deployment not ready")).ToNot(HaveOccurred())
}

func WaitForClusterIpService(testClient *TestClient, serviceObject *corev1.Service) {
	for i := 0; i < 10; i++ {
		_, err := testClient.KubeClient.CoreV1().Services(TestNamespace).Get(serviceObject.Name, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			Expect(err).ToNot(HaveOccurred())
		}

		if err == nil {
			return
		}

		time.Sleep(1 * time.Second)
	}

	Expect(fmt.Errorf("service not created")).ToNot(HaveOccurred())
}

func WaitForServiceToBySynced(testClient *TestClient, serviceObject *corev1.Service) {
	for i := 0; i < 10; i++ {
		serviceObject, err := testClient.KubeClient.CoreV1().Services(TestNamespace).Get(serviceObject.Name, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			Expect(err).ToNot(HaveOccurred())
		}

		if err == nil && serviceObject.Status.LoadBalancer.Ingress != nil && len(serviceObject.Status.LoadBalancer.Ingress) == 1 {
			return
		}

		time.Sleep(3 * time.Second)
	}

	Expect(fmt.Errorf("service don't have sync label")).ToNot(HaveOccurred())
}
func DeleteNginxDeployment(testClient *TestClient, deploymentObject *appsv1.Deployment) {
	err := testClient.KubeClient.AppsV1().Deployments(TestNamespace).Delete(deploymentObject.Name, &metav1.DeleteOptions{})
	Expect(err).ToNot(HaveOccurred())

	for i := 0; i < 10; i++ {
		deploymentObject, err = testClient.KubeClient.AppsV1().Deployments(TestNamespace).Get(deploymentObject.Name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return
			}
			Expect(err).ToNot(HaveOccurred())
		}

		time.Sleep(3 * time.Second)
	}
	Expect(fmt.Errorf("Fail to remove nginx deployment")).ToNot(HaveOccurred())
}

func FarmName(serviceName string) string {
	return fmt.Sprintf("%s-%s", TestNamespace, serviceName)
}

func StartClient(testClient *TestClient) {
	var err error
	ClientPod = &corev1.Pod{ObjectMeta:metav1.ObjectMeta{Name:"test-client"},Spec:corev1.PodSpec{Containers: []corev1.Container{{Name: "client", Image: "quay.io/k8s-nativelb/nativelb-client:latest"}}}}
	ClientPod, err = testClient.KubeClient.CoreV1().Pods(TestNamespace).Create(ClientPod)
	Expect(err).ToNot(HaveOccurred())

	for i := 0; i < 20; i++ {
		ClientPod, err = testClient.KubeClient.CoreV1().Pods(TestNamespace).Get(ClientPod.Name, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			Expect(err).ToNot(HaveOccurred())
		}

		if len(ClientPod.Status.ContainerStatuses) == 1 && ClientPod.Status.ContainerStatuses[0].Ready {
			return
		}

		time.Sleep(6 * time.Second)
	}
}

func CurlFromClient(testClient *TestClient, url string) {
	var (
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)
	command := []string{"curl","-I",url}
	req := testClient.KubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(ClientPod.Name).
		Namespace(ClientPod.Namespace).
		SubResource("exec").
		Param("container", "client")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: "client",
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(testClient.NativelbClient.GetManager().GetConfig(), "POST", req.URL())
	Expect(err).ToNot(HaveOccurred())

	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
		Tty:    false,
	})
	Expect(err).ToNot(HaveOccurred())

	Expect(stdoutBuf.String()).To(ContainSubstring("HTTP/1.1 200 OK"))

	err = testClient.KubeClient.CoreV1().Pods(TestNamespace).Delete(ClientPod.Name,&metav1.DeleteOptions{})
	Expect(err).ToNot(HaveOccurred())

}

func UdpDialFromClient(testClient *TestClient, url string) {
	var (
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)
	command := []string{"/client",url}
	req := testClient.KubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(ClientPod.Name).
		Namespace(ClientPod.Namespace).
		SubResource("exec").
		Param("container", "client")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: "client",
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(testClient.NativelbClient.GetManager().GetConfig(), "POST", req.URL())
	Expect(err).ToNot(HaveOccurred())

	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
		Tty:    false,
	})
	Expect(err).ToNot(HaveOccurred())

	Expect(stdoutBuf.String()).ToNot(ContainSubstring("failed"))

	err = testClient.KubeClient.CoreV1().Pods(TestNamespace).Delete(ClientPod.Name,&metav1.DeleteOptions{})
	Expect(err).ToNot(HaveOccurred())
}