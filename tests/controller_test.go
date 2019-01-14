package tests_test

import (
	"fmt"
	. "github.com/k8s-nativelb/tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"


	"k8s.io/apimachinery/pkg/api/errors"
	core "k8s.io/api/core/v1"
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

	Describe("Create Service", func() {
		It("Should not create a farm for clusterIP service", func() {
			nginxDeploymentName := "nginx1"
			selectorLabel := map[string]string{"app":nginxDeploymentName}
			deployment := CreateNginxDeployment(nginxDeploymentName,"8080",selectorLabel)
			deployment,err = testClient.KubeClient.AppsV1().Deployments(TestNamespace).Create(deployment)
			Expect(err).ToNot(HaveOccurred())

			WaitForDeploymentToBeReady(testClient,deployment)

			ports := []core.ServicePort{{Name: "port1", Protocol: "TCP", Port: 8080}}
			clusterIpService := &core.Service{ObjectMeta: metav1.ObjectMeta{Name: "nginx-service", Namespace: TestNamespace},
				Spec: core.ServiceSpec{Selector: selectorLabel, Ports: ports}}

			clusterIpService, err := testClient.KubeClient.CoreV1().Services(TestNamespace).Create(clusterIpService)
			Expect(err).NotTo(HaveOccurred())

			WaitForClusterIpService(testClient, clusterIpService)

			_, err = testClient.NativelbClient.Farm().Get(FarmName(nginxDeploymentName))
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())

			err = testClient.KubeClient.CoreV1().Services(TestNamespace).Delete(clusterIpService.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())

			DeleteNginxDeployment(testClient,deployment)

		})

		It("Should create a farm representing a TCP Loadbalancer service", func() {
			nginxDeploymentName := "nginx1"
			selectorLabel := map[string]string{"app":nginxDeploymentName}
			deployment := CreateNginxDeployment(nginxDeploymentName,"8080",selectorLabel)
			deployment,err = testClient.KubeClient.AppsV1().Deployments(TestNamespace).Create(deployment)
			Expect(err).ToNot(HaveOccurred())

			WaitForDeploymentToBeReady(testClient,deployment)

			ports := []core.ServicePort{{Name: "port1", Protocol: "TCP", Port: 8080}}
			clusterIpService := &core.Service{ObjectMeta: metav1.ObjectMeta{Name: "nginx-loadbalancer", Namespace: TestNamespace,Annotations:InClusterAgentLabel},
				Spec: core.ServiceSpec{Selector: selectorLabel, Ports: ports, Type: core.ServiceTypeLoadBalancer}}

			clusterIpService, err = testClient.KubeClient.CoreV1().Services(TestNamespace).Create(clusterIpService)
			Expect(err).NotTo(HaveOccurred())

			WaitForServiceToBySynced(testClient, clusterIpService)

			clusterIpService, err = testClient.KubeClient.CoreV1().Services(TestNamespace).Get(clusterIpService.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			By("Checking ingress ip address exist")
			Expect(len(clusterIpService.Status.LoadBalancer.Ingress)).To(Equal(1))

			serviceIpAddr := clusterIpService.Status.LoadBalancer.Ingress[0].IP
			farm, err := testClient.NativelbClient.Farm().Get(FarmName("nginx-loadbalancer"))
			Expect(err).NotTo(HaveOccurred())
			Expect(farm.Status.IpAdress).To(Equal(serviceIpAddr))

			CurlFromClient(testClient,fmt.Sprintf("http://%s:%d",serviceIpAddr,8080))

			err = testClient.KubeClient.CoreV1().Services(TestNamespace).Delete(clusterIpService.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())

			DeleteNginxDeployment(testClient,deployment)
		})

		It("Should create a farm representing a UDP Loadbalancer service", func() {
			nginxDeploymentName := "udp1"
			selectorLabel := map[string]string{"app":nginxDeploymentName}
			deployment := CreateUdpServerDeployment(nginxDeploymentName,"8080",selectorLabel)
			deployment,err = testClient.KubeClient.AppsV1().Deployments(TestNamespace).Create(deployment)
			Expect(err).ToNot(HaveOccurred())

			WaitForDeploymentToBeReady(testClient,deployment)

			ports := []core.ServicePort{{Name: "port1", Protocol: "UDP", Port: 8080}}
			clusterIpService := &core.Service{ObjectMeta: metav1.ObjectMeta{Name: "nginx-udp-loadbalancer", Namespace: TestNamespace,Annotations:InClusterAgentLabel},
				Spec: core.ServiceSpec{Selector: selectorLabel, Ports: ports, Type: core.ServiceTypeLoadBalancer}}

			clusterIpService, err = testClient.KubeClient.CoreV1().Services(TestNamespace).Create(clusterIpService)
			Expect(err).NotTo(HaveOccurred())

			WaitForServiceToBySynced(testClient, clusterIpService)

			clusterIpService, err = testClient.KubeClient.CoreV1().Services(TestNamespace).Get(clusterIpService.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			By("Checking ingress ip address exist")
			Expect(len(clusterIpService.Status.LoadBalancer.Ingress)).To(Equal(1))

			serviceIpAddr := clusterIpService.Status.LoadBalancer.Ingress[0].IP
			farm, err := testClient.NativelbClient.Farm().Get(FarmName("nginx-udp-loadbalancer"))
			Expect(err).NotTo(HaveOccurred())
			Expect(farm.Status.IpAdress).To(Equal(serviceIpAddr))

			UdpDialFromClient(testClient,fmt.Sprintf("%s:%d",serviceIpAddr,8080))

			err = testClient.KubeClient.CoreV1().Services(TestNamespace).Delete(clusterIpService.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())

			DeleteNginxDeployment(testClient,deployment)
		})
	})
})
