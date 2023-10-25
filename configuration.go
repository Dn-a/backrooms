package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Resource struct {
	Name     string `yaml:"name"`
	Matchers string `yaml:"matchers"`
	Type     string `default:"proxy" yaml:"type"`
	Url      string `yaml:"url"`
}

type Configurations struct {
	UpdatedOn       string
	Port            string              `yaml:"port"`
	DefaultUrl      string              `yaml:"default-url"`
	Resources       map[string]Resource `yaml:"resources"`
	RequestMatchers *map[string]interface{}
}

var conf *Configurations

func GetConfig(customFileName ...string) (*Configurations, error) {

	var err error

	fileName := DEFAULT_FILE_NANE

	if len(customFileName) > 0 {
		fileName = customFileName[0]
	}

	info, fileErr := os.Stat(fileName)
	var modTime string

	if fileErr != nil {
		errorMessage := fmt.Sprintf("File %v not found!", fileName)
		err = errors.New(errorMessage)
		log.Println(errorMessage)
	} else {
		modTime = info.ModTime().String()
	}

	if conf == nil {
		conf = &Configurations{UpdatedOn: modTime}
	}

	if fileErr == nil && (len(conf.Resources) == 0 || conf.UpdatedOn != modTime) {
		conf.UpdatedOn = modTime
		readFile, readFileErr := os.ReadFile(fileName)

		if readFileErr != nil {
			log.Printf("File err: %v\n", readFileErr)
		}

		log.Printf("%v lastUpdatedOn: %v", fileName, modTime)
		log.Printf("Unmarshal %v", fileName)

		yamlErr := yaml.Unmarshal(readFile, conf)

		// refresh requestMatchers
		conf.RequestMatchers = generateRequestMatchers(conf)

		if yamlErr != nil {
			log.Printf("err: %v\n", yamlErr)
		}
	}

	return conf, err
}

func generateRequestMatchers(config *Configurations) *map[string]interface{} {
	m := make(map[string]interface{})

	for _, resource := range config.Resources {
		slicedMatchers := strings.Split(resource.Matchers, "/")[1:]
		if len(slicedMatchers) > 0 {
			resourceAddress := resource
			mainKey := slicedMatchers[0]
			slice := slicedMatchers[1:]
			if len(slice) == 0 {
				m[mainKey] = &resourceAddress
			} else {
				m[mainKey] = MatchersRecursion(&slice, &resourceAddress)
			}
		}
	}
	return &m
}

func MatchersRecursion(matchers *[]string, resource *Resource) *map[string]interface{} {
	mp := make(map[string]interface{})
	matcher := (*matchers)[0]

	if len(*matchers) == 1 {
		mp[matcher] = resource
		return &mp
	} else {
		slice := (*matchers)[1:]
		mp[matcher] = MatchersRecursion(&slice, resource)
		return &mp
	}
}

func (c *Configurations) Matchers(path string) (*Resource, bool) {
	check := false
	slicedPath := strings.Split(path, "/")[1:]

	//var queryString string

	log.Printf("sliced path: %v", slicedPath)

	var matchers *map[string]any = c.RequestMatchers
	var rsc *Resource

	for _, p := range slicedPath {
		//Remove query string
		if strings.Contains(p, "?") {
			p = strings.Split(p, "?")[0]
		}

		m, ok := (*matchers)[p]

		if ok {
			if r, rOk := m.(*Resource); rOk {
				rsc = r
				check = true
				break
			} else {
				log.Printf("Ok %v", m)
				matchers = m.(*map[string]interface{})
			}
		} else if m2, ok2 := (*matchers)[JOLLY]; ok2 {
			if r, rOk := m2.(*Resource); rOk {
				log.Printf("Jolly %v", m2)
				rsc = r
				check = true
				break
			} else {
				matchers = m2.(*map[string]interface{})
			}
		} else {
			break
		}
	}

	return rsc, check
}
