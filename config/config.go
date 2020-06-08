// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period         time.Duration `config:"period"`
	LastCheck      int           `config:"lastCheck"`
	Op5host        string        `config:"op5Host"`
	Op5connect     string        `config:"op5Connect"`
	Query          string        `config:"query"`
	Columns        []string      `config:"columns"`
	Filter         []string      `config:"filter"`
	Metrics        bool          `config:"metrics"`
	MetricsAllow   []string      `config:"metrics_allow"`
	MetricsBlock   []string      `config:"metrics_block"`
	MetricsValOnly bool          `config:"metrics_value_only"`
}

var DefaultConfig = Config{
	Period:         30 * time.Second,
	LastCheck:      30,
	Op5host:        "/opt/monitor/var/rw/live",
	Op5connect:     "unix",
	Query:          "services",
	Columns:        []string{"host_name", "display_name", "state", "plugin_output", "perf_data"},
	Filter:         nil,
	Metrics:        true,
	MetricsAllow:   nil,
	MetricsBlock:   nil,
	MetricsValOnly: true,
}

