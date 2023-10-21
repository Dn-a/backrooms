package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Resource struct {
	Name     string `yaml:"name"`
	Matchers string `yaml:"matchers"`
	Type     string `default:"proxy" yaml:"type"`
	Url      string `yaml:"url"`
}

type Configurations struct {
	Port            string              `yaml:"port"`
	DefaultUrl      string              `yaml:"default-url"`
	Resources       map[string]Resource `yaml:"resources"`
	UpdatedOn       string
	RequestMatchers *map[string]interface{}
}

var conf *Configurations

func GetConfig() (*Configurations, error) {
	info, er := os.Stat(CONFIG_FILE_NANE)
	if er != nil {
		log.Printf("FileInfo: %v", er)
	}
	modTime := info.ModTime().String()

	if conf == nil {
		conf = &Configurations{UpdatedOn: modTime}
	}

	if len(conf.Resources) == 0 || conf.UpdatedOn != modTime {
		conf.UpdatedOn = modTime
		readFile, e := os.ReadFile(CONFIG_FILE_NANE)

		if e != nil {
			log.Printf("File err: %v\n", e)
		}

		log.Printf("%v lastUpdatedOn: %v", CONFIG_FILE_NANE, modTime)
		log.Printf("Unmarshal %v", CONFIG_FILE_NANE)

		err := yaml.Unmarshal(readFile, conf)

		// refresh requestMatchers
		conf.RequestMatchers = generateRequestMatchers(conf)

		if err != nil {
			log.Printf("err: %v\n", err)
		}
	}

	//log.Printf("%+v\n", modTime)

	return conf, nil
}
