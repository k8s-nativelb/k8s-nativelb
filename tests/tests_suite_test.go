package tests_test

import (
	"fmt"
	nativelbv1 "github.com/k8s-nativelb/pkg/apis/nativelb/v1"
	. "github.com/k8s-nativelb/tests"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var testClient *TestClient
var err error

func TestTests(t *testing.T) {
	RegisterFailHandler(NativeLBFailedFunction)
	RunSpecs(t, "Tests Suite")
}

var _ = BeforeSuite(func() {
	testClient, err = NewTestClient()
	PanicOnError(err)

	if _, err := testClient.KubeClient.CoreV1().Namespaces().Get(TestNamespace, metav1.GetOptions{}); err == nil {
		EventuallyWithOffset(1, func() bool { return errors.IsNotFound(testClient.DeleteTestNamespace()) },120*time.Second, 5*time.Second).
			Should(BeTrue())
	}

	testClient.CreateTestNamespace()
	testClient.CleanNativelbNamespace()
	StartClient(testClient)
})

var _ = AfterSuite(func() {
	fmt.Printf("Waiting for namespace %s to be removed, this can take a while ...\n", TestNamespace)
	EventuallyWithOffset(1, func() bool { return errors.IsNotFound(testClient.DeleteTestNamespace()) }, 120*time.Second, 5*time.Second).
		Should(BeTrue())
	testClient.CleanNativelbNamespace()
})

func NativeLBFailedFunction(message string, callerSkip ...int) {
	clusterList, err := testClient.NativelbClient.Cluster().List(&client.ListOptions{})
	if err != nil {
		fmt.Println(err)
		Fail(message, callerSkip...)
	}

	for _,clusterObject := range clusterList.Items {
		fmt.Println(clusterObject)
	}

	agentList, err := testClient.NativelbClient.Agent().List(&client.ListOptions{})
	if err != nil {
		fmt.Println(err)
		Fail(message, callerSkip...)
	}

	for _,agentObject := range agentList.Items {
		fmt.Println(agentObject)
	}

	podList,err := testClient.KubeClient.CoreV1().Pods(nativelbv1.ControllerNamespace).List(metav1.ListOptions{})
	var tailLines int64 = 30
	for _,podObject := range podList.Items {
		fmt.Println(podObject)
		logsRaw, err :=testClient.KubeClient.CoreV1().Pods(nativelbv1.ControllerNamespace).GetLogs(podObject.Name,&corev1.PodLogOptions{TailLines: &tailLines,
			Container: podObject.Spec.Containers[0].Name}).DoRaw()
		if err == nil {
			fmt.Printf(string(logsRaw))
		}
	}

	podList,err = testClient.KubeClient.CoreV1().Pods(TestNamespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println(err)
		Fail(message, callerSkip...)
	}

	for _,agentObject := range agentList.Items {
		fmt.Println(agentObject)
	}

	Fail(message, callerSkip...)
}