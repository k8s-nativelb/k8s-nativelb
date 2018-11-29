package tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
})
