package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type ConfigYaml struct {
	Servers  []ServersConfig   `yaml:"servers"`
	Items    map[string]string `yaml:"items"`
	Group    string            `yaml:"group"`
	Interval int               `yaml:"interval"`
}

type ServersConfig struct {
	ID       string `yaml:"id"`
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func InitServersConfig() ConfigYaml {
	yamlFile, err := ioutil.ReadFile("./config/zabbix.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
	var conf ConfigYaml
	// serversMap := make(map[int]ServersConfig)
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		fmt.Println(err)
	}
	return conf

}
