package main

import (
	"fmt"
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

	It("File not fonud", func() {
		var err error
		fileName := "noFile.yml"
		config, err = GetConfig(fileName)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(fmt.Sprintf("File %v not found!", fileName)))
	})

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
