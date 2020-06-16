package main

import (
	"testing"
)

func TestRunner(t *testing.T) {
	config, err := readConfig("../../config-example.json")
	if err != nil {
		t.Error(err)
	}

	if len(config.Buckets) <= 0 {
		t.Error("buckets are empty")
	}

	runner, err := Runner(config)
	if err != nil {
		t.Error(err)
	}

	if _, err := runner.AllStatus(); err != nil {
		t.Error(err)
	}
}
