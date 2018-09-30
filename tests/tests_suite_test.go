package tests_test

import (
	"fmt"
	. "github.com/k8s-nativelb/tests"
	"k8s.io/apimachinery/pkg/api/errors"
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
	testClient.CreateTestNamespace()
})

var _ = AfterSuite(func() {
	testClient, err := NewTestClient()
	PanicOnError(err)
	fmt.Printf("Waiting for namespace %s to be removed, this can take a while ...\n",TestNamespace )
	EventuallyWithOffset(1, func() bool { return errors.IsNotFound(testClient.DeleteTestNamespace()) }, 30*time.Second, 1*time.Second).
		Should(BeTrue())
})