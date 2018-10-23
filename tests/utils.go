package tests

import (
	"fmt"
	nativelb "github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	"k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
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

func CreateNginxDeployment() *v1.Deployment {

	replicas := int32(1)
	return &v1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: NginxDeploymentName, Namespace: TestNamespace, Labels: SelectorLabel},
		Spec: v1.DeploymentSpec{Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: SelectorLabel},
			Template: v12.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: SelectorLabel},
				Spec: v12.PodSpec{Containers: []v12.Container{{Name: "nginx", Image: "docker.io/nginx:latest"}}}}}}
}

func WaitForDeploymentToBeReady(testClient *TestClient, deploymentObject *v1.Deployment) error {
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

func WaitForClusterIpService(testClient *TestClient, serviceObject *v12.Service) error {
	for i := 0; i < 10; i++ {
		_, err := testClient.KubeClient.Core().Services(TestNamespace).Get(serviceObject.Name, metav1.GetOptions{})
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

func WaitForServiceToBySynced(testClient *TestClient, serviceObject *v12.Service) error {
	for i := 0; i < 10; i++ {
		serviceObject, err := testClient.KubeClient.Core().Services(TestNamespace).Get(serviceObject.Name, metav1.GetOptions{})
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
func DeleteNginxDeployment(testClient *TestClient, deploymentObject *v1.Deployment) error {
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
