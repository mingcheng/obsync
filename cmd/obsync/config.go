package main

import (
	"github.com/mingcheng/obsync"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Config struct {
	Log struct {
		Debug bool   `json:"debug" yaml:"debug"`
		Path  string `json:"path" yaml:"path"`
	} `json:"log" yaml:"log"`
	RunnerConfigs []obsync.RunnerConfig `json:"targets" yaml:"targets"`
}

func NewConfig(configPath string) (config Config, err error) {
	var data []byte

	// read config and initial obs client
	data, err = ioutil.ReadFile(configPath)
	if err != nil {
		return
	}

	// unmarshal config into initialized objects
	if err = yaml.Unmarshal(data, &config); err != nil {
		return
	}

	return
}
