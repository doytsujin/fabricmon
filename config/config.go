// Copyright 2017-20 Daniel Swarbrick. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package config handles the configuration parsing for FabricMon.
package config

import (
	"fmt"
	"io"
	"time"

	"golang.org/x/sys/unix"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// FabricmonConf is the main configuration struct for FabricMon.
type FabricmonConf struct {
	PollInterval   time.Duration `yaml:"poll_interval"`
	ResetThreshold uint          `yaml:"counter_reset_threshold"`
	Mkey           uint64        `yaml:"m_key"`
	InfluxDB       []InfluxDBConf
	Logging        LoggingConf
	Topology       TopologyConf
}

func (conf *FabricmonConf) validate() error {
	if conf.ResetThreshold < 25 || conf.ResetThreshold > 100 {
		return fmt.Errorf("counter_reset_threshold must be between 25 and 100")
	}

	return nil
}

// InfluxDBConf holds the configuration values for a single InfluxDB instance.
type InfluxDBConf struct {
	URL             string
	Database        string
	Username        string
	Password        string
	RetentionPolicy string `yaml:"retention_policy"`
	Timeout         time.Duration
}

type LoggingConf struct {
	LogLevel LogLevel `yaml:"log_level"`
}

type TopologyConf struct {
	Enabled   bool
	OutputDir string `yaml:"output_dir"`
}

func (conf *TopologyConf) validate() error {
	if conf.Enabled {
		if err := unix.Access(conf.OutputDir, unix.W_OK); err != nil {
			return fmt.Errorf("topology output directory: %s", err)
		}
	}

	return nil
}

// LogLevel is a wrapper type for logrus.Level.
type LogLevel logrus.Level

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	return logrus.Level(l).String()
}

// UnmarshalText parses a byte slice value into a logrus.Level value.
func (l *LogLevel) UnmarshalText(text []byte) error {
	level, err := logrus.ParseLevel(string(text))

	if err == nil {
		*l = LogLevel(level)
	}

	return err
}

func ReadConfig(r io.Reader) (*FabricmonConf, error) {
	// Defaults
	conf := &FabricmonConf{
		PollInterval: time.Second * 10,
		Logging: LoggingConf{
			LogLevel: LogLevel(logrus.InfoLevel),
		},
	}

	dec := yaml.NewDecoder(r)
	if err := dec.Decode(conf); err != nil {
		return nil, err
	}

	if err := conf.validate(); err != nil {
		return nil, err
	}

	if err := conf.Topology.validate(); err != nil {
		return nil, err
	}

	return conf, nil
}
