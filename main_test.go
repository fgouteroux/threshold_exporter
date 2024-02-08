package main

import (
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config, err := load("config_test.yaml")
	if err != nil {
		t.Fatal(err)
	}

	var expected []MetricConfig
	expected = append(expected, MetricConfig{
		MetricName:  "server_cpu_threshold_percent",
		MetricHelp:  "The used cpu threshold",
		MetricType:  "gauge",
		MetricValue: 0.05,
		MetricLabels: map[string]string{
			"priority": "SEV-5",
			"severity": "warning",
			"process":  "all",
		},
	})

	if !reflect.DeepEqual(config.Metrics, expected) {
		t.Fatalf("%+v != %+v", config.Metrics, expected)
	}
}
