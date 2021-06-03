package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/mingcheng/obsync/bucket"
)

type Config struct {
	// Debug      bool            `json:"debug"`
	// Standalone bool            `json:"standalone"`
	// Interval   uint            `json:"interval"`
	// Force   bool            `json:"force"`
	Root    string          `json:"root"`
	Buckets []bucket.Config `json:"buckets"`
}

func (c *Config) Dump() (config string, err error) {
	result, err := json.Marshal(c)
	if err != nil {
		return "", nil
	}

	return string(result), nil
}

func NewConfig(configPath string) (*Config, error) {
	var (
		err    error
		data   []byte
		config Config
	)

	// read config and initial obs client
	data, err = ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
