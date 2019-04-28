package tests_test

import (
	"fmt"
	. "github.com/k8s-nativelb/tests"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/rand"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Controller", func() {
	Describe("Namespaces", func() {
		It("Test namespace should exist", func() {
			_, err := testClient.GetTestNamespace()
			Expect(err).NotTo(HaveOccurred())
		})

		It("NativeLB namespace should exist", func() {
			_, err := testClient.GetNativeLBNamespace()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	table.DescribeTable("Should not create a farm for clusterIP service", func(clusterLabel map[string]string) {
		nginxDeploymentName := "nginx-" + rand.String(12)
		selectorLabel := map[string]string{"app": nginxDeploymentName}
		deployment := CreateNginxDeployment(nginxDeploymentName, "8080", selectorLabel)
		deployment, err = testClient.KubeClient.AppsV1().Deployments(TestNamespace).Create(deployment)
		Expect(err).ToNot(HaveOccurred())

		WaitForDeploymentToBeReady(testClient, deployment)

		ports := []core.ServicePort{{Name: "port1", Protocol: "TCP", Port: 8080}}
		clusterIpService := &core.Service{ObjectMeta: metav1.ObjectMeta{Name: "nginx-service-" + rand.String(12), Namespace: TestNamespace, Annotations: clusterLabel},
			Spec: core.ServiceSpec{Selector: selectorLabel, Ports: ports}}

		clusterIpService, err := testClient.KubeClient.CoreV1().Services(TestNamespace).Create(clusterIpService)
		Expect(err).NotTo(HaveOccurred())

		WaitForClusterIpService(testClient, clusterIpService)

		_, err = testClient.NativelbClient.Farm(TestNamespace).Get(nginxDeploymentName)
		Expect(err).To(HaveOccurred())
		Expect(errors.IsNotFound(err)).To(BeTrue())

		err = testClient.KubeClient.CoreV1().Services(TestNamespace).Delete(clusterIpService.Name, &metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())

		DeleteNginxDeployment(testClient, deployment)
	},
		table.Entry("Daemon Cluster Agent", DaemonClusterAgentLabel),
		table.Entry("External Cluster Agent", ExternalClusterAgentLabel))

	table.DescribeTable("Should create a farm representing a TCP Loadbalancer service", func(clusterLabel map[string]string) {
		nginxDeploymentName := "nginx-" + rand.String(12)
		selectorLabel := map[string]string{"app": nginxDeploymentName}
		deployment := CreateNginxDeployment(nginxDeploymentName, "8081", selectorLabel)
		deployment, err = testClient.KubeClient.AppsV1().Deployments(TestNamespace).Create(deployment)
		Expect(err).ToNot(HaveOccurred())

		WaitForDeploymentToBeReady(testClient, deployment)

		ports := []core.ServicePort{{Name: "port1", Protocol: "TCP", Port: 8081}}
		clusterIpService := &core.Service{ObjectMeta: metav1.ObjectMeta{Name: "nginx-loadbalancer-" + rand.String(12), Namespace: TestNamespace, Annotations: clusterLabel},
			Spec: core.ServiceSpec{Selector: selectorLabel, Ports: ports, Type: core.ServiceTypeLoadBalancer}}

		clusterIpService, err = testClient.KubeClient.CoreV1().Services(TestNamespace).Create(clusterIpService)
		Expect(err).NotTo(HaveOccurred())

		WaitForServiceToBySynced(testClient, clusterIpService)

		clusterIpService, err = testClient.KubeClient.CoreV1().Services(TestNamespace).Get(clusterIpService.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		By("Checking ingress ip address exist")
		Expect(len(clusterIpService.Status.LoadBalancer.Ingress)).To(Equal(1))

		serviceIpAddr := clusterIpService.Status.LoadBalancer.Ingress[0].IP
		farm, err := testClient.NativelbClient.Farm(TestNamespace).Get(clusterIpService.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(farm.Status.IpAdress).To(Equal(serviceIpAddr))

		CurlFromClient(testClient, fmt.Sprintf("http://%s:%d", serviceIpAddr, 8081), true)

		err = testClient.KubeClient.CoreV1().Services(TestNamespace).Delete(clusterIpService.Name, &metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			_, err = testClient.NativelbClient.Farm(TestNamespace).Get(clusterIpService.Name)
			return err
		}, 30, 5).ShouldNot(BeNil())
		Expect(errors.IsNotFound(err)).To(BeTrue())

		CurlFromClient(testClient, fmt.Sprintf("http://%s:%d", serviceIpAddr, 8081), false)
		DeleteNginxDeployment(testClient, deployment)
	},
		table.Entry("Daemon Cluster Agent", DaemonClusterAgentLabel),
		table.Entry("External Cluster Agent", ExternalClusterAgentLabel))

	table.DescribeTable("Should create a farm representing a UDP Loadbalancer service", func(clusterLabel map[string]string) {
		nginxDeploymentName := "nginx-" + rand.String(12)
		selectorLabel := map[string]string{"app": nginxDeploymentName}
		deployment := CreateUdpServerDeployment(nginxDeploymentName, "8081", selectorLabel)
		deployment, err = testClient.KubeClient.AppsV1().Deployments(TestNamespace).Create(deployment)
		Expect(err).ToNot(HaveOccurred())

		WaitForDeploymentToBeReady(testClient, deployment)

		ports := []core.ServicePort{{Name: "port1", Protocol: "UDP", Port: 8081}}
		clusterIpService := &core.Service{ObjectMeta: metav1.ObjectMeta{Name: "nginx-udp-loadbalancer-" + rand.String(6), Namespace: TestNamespace, Annotations: clusterLabel},
			Spec: core.ServiceSpec{Selector: selectorLabel, Ports: ports, Type: core.ServiceTypeLoadBalancer}}

		clusterIpService, err = testClient.KubeClient.CoreV1().Services(TestNamespace).Create(clusterIpService)
		Expect(err).NotTo(HaveOccurred())

		WaitForServiceToBySynced(testClient, clusterIpService)

		clusterIpService, err = testClient.KubeClient.CoreV1().Services(TestNamespace).Get(clusterIpService.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		By("Checking ingress ip address exist")
		Expect(len(clusterIpService.Status.LoadBalancer.Ingress)).To(Equal(1))

		serviceIpAddr := clusterIpService.Status.LoadBalancer.Ingress[0].IP
		farm, err := testClient.NativelbClient.Farm(TestNamespace).Get(clusterIpService.Name)
		Expect(err).NotTo(HaveOccurred())
		Expect(farm.Status.IpAdress).To(Equal(serviceIpAddr))

		UdpDialFromClient(testClient, fmt.Sprintf("%s:%d", serviceIpAddr, 8081), true)

		err = testClient.KubeClient.CoreV1().Services(TestNamespace).Delete(clusterIpService.Name, &metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			_, err = testClient.NativelbClient.Farm(TestNamespace).Get(clusterIpService.Name)
			return err
		}, 30, 5).ShouldNot(BeNil())
		Expect(errors.IsNotFound(err)).To(BeTrue())

		UdpDialFromClient(testClient, fmt.Sprintf("%s:%d", serviceIpAddr, 8081), false)
		DeleteNginxDeployment(testClient, deployment)
	},
		table.Entry("Daemon Cluster Agent", DaemonClusterAgentLabel),
		table.Entry("External Cluster Agent", ExternalClusterAgentLabel))
})
