package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Env                  string        `yaml:"env"`
	UdpPortRange         string        `yaml:"udp_port_range"`
	HttpAddr             string        `yaml:"http_addr"`
	FpmStatusURL         string        `yaml:"fpm_status_url"`
	HttpClientTimeout    time.Duration `yaml:"http_client_timeout"`
	LoadFpmStatusTimeout time.Duration `yaml:"load_fpm_status_timeout"`
	StuckProcessDuration time.Duration `yaml:"stuck_process_duration"`
	Buffer               int           `yaml:"buffer"`
	PacketsSize          int           `yaml:"packets_size"`
	AppName              string        `yaml:"app_name"`
	LayoutTime           string
	UdpPortStart         int
	UdpPortEnd           int
	UdpPortRangeCount    int

	verbosity int
}

func (c *Config) SetVerbosity(isVerbose, isVeryVerbose, isVeryVeryVerbose bool) {
	if isVeryVeryVerbose {
		c.verbosity = 3
	}
	if isVeryVerbose {
		c.verbosity = 2
	}
	if isVerbose {
		c.verbosity = 1
	}
}

func (c *Config) IsVerboseByLevel(verboseLevel string) bool {
	return c.verbosity >= len(verboseLevel)
}

func LoadFromFile(filePath string) (*Config, error) {
	configBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	cfg := Config{}
	if err = yaml.Unmarshal(configBytes, &cfg); err != nil {
		return nil, err
	}

	cfg.LayoutTime = "2006-01-02T15:04:05.000000-07:00"
	portRange := strings.Split(cfg.UdpPortRange, "-")
	cfg.UdpPortStart, _ = strconv.Atoi(portRange[0])
	cfg.UdpPortEnd, _ = strconv.Atoi(portRange[1])
	cfg.UdpPortRangeCount = cfg.UdpPortEnd - cfg.UdpPortStart + 1

	return &cfg, nil
}
