package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type NamedValue struct {
	Name  string
	Value string
}

type Config struct {
	Listener string
	Pattern  string
	Labels   prometheus.Labels
}

func copyLabels(in prometheus.Labels) prometheus.Labels {
	out := prometheus.Labels{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func Parse() (*Config, error) {
	listener := "0.0.0.0:2198"
	pattern := "[b]idule"
	labels := "application:prometheus-fd"
	flag.StringVar(&listener, "listener", listener, "listener address")
	flag.StringVar(&pattern, "pattern", pattern, "ps query pattern")
	flag.StringVar(&labels, "labels", labels, "a list of key:value labels separated by commas")
	flag.Parse()
	l := strings.Split(labels, ",")
	prometheusLabels := prometheus.Labels{}
	for _, label := range l {
		keyValue := strings.Split(label, ":")
		if len(keyValue) != 2 {
			return nil, fmt.Errorf("unrecognized label: %s", keyValue)
		}
		prometheusLabels[keyValue[0]] = keyValue[1]
	}
	return &Config{
		Listener: listener,
		Pattern:  pattern,
		Labels:   prometheusLabels,
	}, nil
}
