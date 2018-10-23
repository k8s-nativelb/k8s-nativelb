package tests_test

import (
	"fmt"
	. "github.com/k8s-nativelb/tests"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests Suite")
}

var _ = BeforeSuite(func() {
	testClient, err := NewTestClient()
	PanicOnError(err)

	if _, err := testClient.KubeClient.Core().Namespaces().Get(TestNamespace, metav1.GetOptions{}); err == nil {
		EventuallyWithOffset(1, func() bool { return errors.IsNotFound(testClient.DeleteTestNamespace()) }, 30*time.Second, 1*time.Second).
			Should(BeTrue())
	}

	testClient.CreateTestNamespace()
	testClient.CleanNativelbNamespace()
})

var _ = AfterSuite(func() {
	testClient, err := NewTestClient()
	PanicOnError(err)
	fmt.Printf("Waiting for namespace %s to be removed, this can take a while ...\n", TestNamespace)
	EventuallyWithOffset(1, func() bool { return errors.IsNotFound(testClient.DeleteTestNamespace()) }, 30*time.Second, 1*time.Second).
		Should(BeTrue())
	testClient.CleanNativelbNamespace()
})
