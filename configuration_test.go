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

var _ = Describe("Configuration", func() {
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

	It("should return stripped path array", func() {
		url := "http://domain.com/res1/res2"
		path := getPathAsArray(url)
		Expect(path).NotTo(BeEmpty())
		Expect(path[0]).To(Equal("res1"))
		Expect(path[1]).To(Equal("res2"))
	})
	It("should return path array", func() {
		url := "/res1/res2"
		path := getPathAsArray(url)
		Expect(path).NotTo(BeEmpty())
		Expect(path[0]).To(Equal("res1"))
		Expect(path[1]).To(Equal("res2"))
	})

	It("should generate Matchers", func() {
		config = &Configurations{Resources: make(map[string]Resource)}
		config.Resources["res1"] = Resource{Name: "res1", Matchers: "/res1/**", Url: "http://localhost"}
		config.RequestMatchers = generateRequestMatchers(config)

		matchers := (*config.RequestMatchers)

		Expect(matchers).NotTo(BeNil())
		Expect((matchers)["res1"]).NotTo(BeNil())
	})

	It("should match", func() {
		config = &Configurations{Resources: make(map[string]Resource)}
		config.Resources["res1"] = Resource{Name: "res1", Matchers: "/res1/**", Url: "http://localhost"}
		config.RequestMatchers = generateRequestMatchers(config)

		url := "http://localhost/res1/res2"
		res, _ := config.Matchers(url)

		Expect(res).NotTo(BeNil())
		Expect(res.Url).To(Equal("http://localhost"))
	})

})
