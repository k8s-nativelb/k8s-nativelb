package tests_test

import (
	. "github.com/k8s-nativelb/tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Internal Cluster", func() {
	testClient, err := NewTestClient()
	PanicOnError(err)

	var nginxDeployment *v1.Deployment

	BeforeEach(func() {
		nginxDeployment = CreateNginxDeployment()
		nginxDeployment, err := testClient.KubeClient.AppsV1().Deployments(TestNamespace).Create(nginxDeployment)
		Expect(err).NotTo(HaveOccurred())

		err = WaitForDeploymentToBeReady(testClient, nginxDeployment)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := DeleteNginxDeployment(testClient, nginxDeployment)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create Service", func() {
		It("Should not create a farm for clusterIP service", func() {
			ports := []core.ServicePort{{Name: "port1", Protocol: "TCP", Port: 8080}}
			clusterIpService := &core.Service{ObjectMeta: metav1.ObjectMeta{Name: "nginx-service", Namespace: TestNamespace},
				Spec: core.ServiceSpec{Selector: SelectorLabel, Ports: ports}}

			clusterIpService, err := testClient.KubeClient.Core().Services(TestNamespace).Create(clusterIpService)
			Expect(err).NotTo(HaveOccurred())
			err = WaitForClusterIpService(testClient, clusterIpService)
			Expect(err).NotTo(HaveOccurred())

			_, err = testClient.NativelbClient.Farm().Get(FarmName(NginxDeploymentName))
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())

			err = testClient.KubeClient.Core().Services(TestNamespace).Delete(clusterIpService.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should create a farm representing a Loadbalancer service", func() {
			ports := []core.ServicePort{{Name: "port1", Protocol: "TCP", Port: 8080}}
			clusterIpService := &core.Service{ObjectMeta: metav1.ObjectMeta{Name: "nginx-loadbalancer", Namespace: TestNamespace},
				Spec: core.ServiceSpec{Selector: SelectorLabel, Ports: ports, Type: core.ServiceTypeLoadBalancer}}

			clusterIpService, err := testClient.KubeClient.Core().Services(TestNamespace).Create(clusterIpService)
			Expect(err).NotTo(HaveOccurred())
			err = WaitForServiceToBySynced(testClient, clusterIpService)
			Expect(err).NotTo(HaveOccurred())

			clusterIpService, err = testClient.KubeClient.Core().Services(TestNamespace).Get(clusterIpService.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			By("Checking ingress ip address exist")
			Expect(len(clusterIpService.Status.LoadBalancer.Ingress)).To(Equal(1))

			serviceIpAddr := clusterIpService.Status.LoadBalancer.Ingress[0].IP
			farm, err := testClient.NativelbClient.Farm().Get(FarmName("nginx-loadbalancer"))
			Expect(err).NotTo(HaveOccurred())
			Expect(farm.Status.IpAdress).To(Equal(serviceIpAddr))
		})
	})
})
