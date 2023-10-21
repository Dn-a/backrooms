package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Backrooms configuration yaml")
}

var _ = Describe("Backrooms", func() {
	var config *Configurations
	var err error

	BeforeEach(func() {
		config, err = GetConfig()
		Expect(err).NotTo(HaveOccurred())
	})

	/* AfterEach(func() {

	}) */

	It("has port", func() {
		Expect(config.Port).NotTo(BeNil())
	})

	It("has default url", func() {
		Expect(config.DefaultUrl).NotTo(BeNil())
	})

	It("has updated on", func() {
		Expect(config.UpdatedOn).NotTo(BeNil())
	})

})
