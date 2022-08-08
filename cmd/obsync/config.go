package main

import (
	"github.com/mingcheng/obsync"
	"gopkg.in/yaml.v3"
	"os"
)

// Config is the main configuration struct for program
type Config struct {
	Log struct {
		Debug bool   `json:"debug" yaml:"debug"`
		Path  string `json:"path" yaml:"path"`
	} `json:"log" yaml:"log"`
	RunnerConfigs []obsync.RunnerConfig `json:"targets" yaml:"targets"`
}

// NewConfig to create a new configuration form specified file path
func NewConfig(configPath string) (config Config, err error) {
	var data []byte

	// read config and initial obs client
	data, err = os.ReadFile(configPath)
	if err != nil {
		return
	}

	// unmarshal config into initialized objects
	if err = yaml.Unmarshal(data, &config); err != nil {
		return
	}

	return
}
