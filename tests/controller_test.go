package tests_test

import (
	. "github.com/k8s-nativelb/tests"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Controller", func() {
	testClient, err := NewTestClient()
	PanicOnError(err)

	Describe("Test Namespace", func() {
		It("Should exist", func() {
			_, err := testClient.GetTestNamespace()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
