package main

import (
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Metrics []MetricConfig `yaml:"metrics"`
}

type MetricConfig struct {
	MetricName   string            `yaml:"name"`
	MetricHelp   string            `yaml:"help"`
	MetricType   string            `yaml:"type"`
	MetricValue  float64           `yaml:"value"`
	MetricLabels map[string]string `yaml:"labels"`
}

type Collector struct {
	RingConfig ExporterRing
	Logger     log.Logger
	ConfigPath string
}

func (c Collector) Describe(_ chan<- *prometheus.Desc) {}

func (c Collector) Collect(ch chan<- prometheus.Metric) {
	if c.RingConfig.enabled {
		// If another replica is the leader, don't expose any metrics from this one.
		isLeaderNow, err := isLeader(c.RingConfig.client, c.RingConfig.lifecycler)
		if err != nil {
			level.Warn(c.Logger).Log("msg", "Failed to determine ring leader", "err", err)
			return
		}
		if !isLeaderNow {
			level.Debug(c.Logger).Log("msg", "not the ring leader")
			return
		}
		level.Debug(c.Logger).Log("msg", "ring leader")
	}
	var errVal float64
	config, err := load(c.ConfigPath)
	if err != nil {
		level.Error(c.Logger).Log("msg", "Failed to open config file", "err", err)
		errVal = 1.0
	} else {
		level.Info(c.Logger).Log("msg", "reading config file", "file", c.ConfigPath)
		for _, m := range config.Metrics {
			level.Debug(c.Logger).Log("name", m.MetricName)

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					m.MetricName,
					m.MetricHelp,
					nil, m.MetricLabels,
				),
				prometheus.GaugeValue, m.MetricValue,
			)
		}
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"threshold_exporter_scrape_error",
			"1 if there was an error opening or reading a file, 0 otherwise",
			nil, nil,
		),
		prometheus.GaugeValue, errVal,
	)
}

func load(path string) (Config, error) {
	var metrics Config
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return metrics, err
	}
	err = yaml.Unmarshal(yamlFile, &metrics)
	if err != nil {
		return metrics, err
	}
	return metrics, nil
}
