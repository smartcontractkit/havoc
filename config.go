package havoc

import (
	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	"os"
	"strings"
)

const (
	ErrReadSethConfig      = "failed to read TOML config for havoc"
	ErrUnmarshalSethConfig = "failed to unmarshal TOML config for havoc"

	ErrFailureGroupIsNil      = "failure group must be specified in config"
	ErrLatencyGroupIsNil      = "latency group must be specified in config"
	ErrStressCPUGroupIsNil    = "stress cpu group must be specified in config"
	ErrStressMemoryGroupIsNil = "stress memory group must be specified in config"
	ErrFormat                 = "format error"
)

const (
	DefaultExperimentsDir         = "havoc-experiments"
	DefaultGroupPercentage        = "30"
	DefaultPodFailureDuration     = "1m"
	DefaultNetworkLatencyDuration = "1m"
	DefaultStressMemoryDuration   = "1m"
	DefaultStressMemoryWorkers    = 1
	DefaultStressMemoryAmount     = "512MB"
	DefaultStressCPUDuration      = "1m"
	DefaultStressCPUWorkers       = 1
	DefaultStressCPULoad          = 100
	DefaultNetworkLatency         = "500ms"
)

type Config struct {
	Havoc *Havoc `toml:"havoc"`
}

type Havoc struct {
	Dir               string           `toml:"dir"`
	IgnoredPods       []string         `toml:"ignore_pods"`
	IgnoreGroupLabels []string         `toml:"ignore_group_labels"`
	Failure           *Failure         `toml:"failure"`
	Latency           *Latency         `toml:"latency"`
	StressMemory      *StressMemory    `toml:"stress_memory"`
	StressCPU         *StressCPU       `toml:"stress_cpu"`
	ExternalTargets   *ExternalTargets `toml:"external_targets"`
	Monkey            *Monkey          `toml:"monkey"`
	Grafana           *Grafana         `toml:"grafana"`
}

func DefaultConfig() *Config {
	return &Config{
		Havoc: &Havoc{
			Dir: DefaultExperimentsDir,
			Failure: &Failure{
				Duration:        DefaultPodFailureDuration,
				GroupPercentage: DefaultGroupPercentage,
			},
			Latency: &Latency{
				Duration:        DefaultNetworkLatencyDuration,
				Latency:         DefaultNetworkLatency,
				GroupPercentage: DefaultGroupPercentage,
			},
			StressMemory: &StressMemory{
				Duration: DefaultStressMemoryDuration,
				Workers:  DefaultStressMemoryWorkers,
				Memory:   DefaultStressMemoryAmount,
			},
			StressCPU: &StressCPU{
				Duration: DefaultStressCPUDuration,
				Workers:  DefaultStressCPUWorkers,
				Load:     DefaultStressCPULoad,
			},
			Monkey: &Monkey{
				Duration:                "999h",
				Mode:                    "seq",
				Cooldown:                "30s",
				MaxSimultaneousFailures: 1,
			},
			Grafana: &Grafana{
				URL:           os.Getenv("GRAFANA_URL"),
				Token:         os.Getenv("GRAFANA_TOKEN"),
				DashboardName: os.Getenv("DASHBOARD_NAME"),
			},
		},
	}
}

func (c *Config) Validate() []error {
	errs := make([]error, 0)
	if c.Havoc.Dir == "" {
		errs = append(errs, errors.Wrap(errors.New(ErrFormat), "monkey.dir must not be empty"))
	}
	if c.Havoc.Failure == nil {
		errs = append(errs, errors.New(ErrFailureGroupIsNil))
	}
	if c.Havoc.Latency == nil {
		errs = append(errs, errors.New(ErrLatencyGroupIsNil))
	}
	if c.Havoc.StressCPU == nil {
		errs = append(errs, errors.New(ErrStressCPUGroupIsNil))
	}
	if c.Havoc.StressMemory == nil {
		errs = append(errs, errors.New(ErrStressMemoryGroupIsNil))
	}
	if c.Havoc.Failure != nil {
		if c.Havoc.Failure.Duration == "" {
			errs = append(errs, errors.Wrap(errors.New(ErrFormat), "failure.duration must be in Go duration format, 1d2h3m0s"))
		}
	}
	if c.Havoc.Latency != nil {
		if c.Havoc.Latency.Duration == "" {
			errs = append(errs, errors.Wrap(errors.New(ErrFormat), "latency.duration must be in Go duration format, 1d2h3m0s"))
		}
		if c.Havoc.Latency.Latency == "" {
			errs = append(errs, errors.Wrap(errors.New(ErrFormat), "latency.latency must be in milliseconds format, ex.: 300ms"))
		}
	}
	if c.Havoc.StressMemory != nil {
		if c.Havoc.StressMemory.Workers <= 0 {
			errs = append(errs, errors.Wrap(errors.New(ErrFormat), "stress_memory.workers must be set, ex.: \"4\""))
		}
		if c.Havoc.StressMemory.Memory == "" {
			errs = append(errs, errors.Wrap(errors.New(ErrFormat), "stress_memory.memory must be set, ex.: \"256MB\" or \"25%\""))
		}
	}
	if c.Havoc.StressCPU != nil {
		if c.Havoc.StressCPU.Workers <= 0 {
			errs = append(errs, errors.Wrap(errors.New(ErrFormat), "stress_cpu.workers must be set, ex.: \"1\""))
		}
		if c.Havoc.StressCPU.Load <= 0 {
			errs = append(errs, errors.Wrap(errors.New(ErrFormat), "stress_cpu.load must be set, ex.: \"100\""))
		}
	}
	if c.Havoc.Monkey != nil {
		if c.Havoc.Monkey.Mode == "" {
			errs = append(errs, errors.Wrap(errors.New(ErrFormat), "monkey.mode must be either \"seq\" or \"rand\""))
		}
		if c.Havoc.Monkey.Duration == "" {
			errs = append(errs, errors.Wrap(errors.New(ErrFormat), "monkey.duration must be in Go duration format, 1d2h3m0s"))
		}
		if c.Havoc.Monkey.MaxSimultaneousFailures < 0 {
			errs = append(errs, errors.Wrap(errors.New(ErrFormat), "monkey.max_simultaneous_failures must be > 0"))
		}
	}
	return errs
}

type Failure struct {
	Duration        string `toml:"duration"`
	GroupPercentage string `toml:"group_percentage"`
}

type Latency struct {
	Duration        string `toml:"duration"`
	GroupPercentage string `toml:"group_percentage"`
	Latency         string `toml:"latency"`
}

type StressMemory struct {
	Duration string `toml:"duration"`
	Workers  int    `toml:"workers"`
	Memory   string `toml:"memory"`
}

type StressCPU struct {
	Duration string `toml:"duration"`
	Workers  int    `toml:"workers"`
	Load     int    `toml:"load"`
}

type ExternalTargets struct {
	Duration string   `toml:"duration"`
	URLs     []string `toml:"urls"`
}

type Monkey struct {
	Duration                string `toml:"duration"`
	Cooldown                string `toml:"cooldown"`
	Mode                    string `toml:"mode"`
	MaxSimultaneousFailures int    `toml:"max_simultaneous_failures"`
}

type Grafana struct {
	URL           string `toml:"grafana_url"`
	Token         string `toml:"grafana_token"`
	DashboardName string `toml:"dashboard_name"`
}

func ReadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		L.Info().Msg("No config specified, using default configuration")
	} else {
		L.Debug().
			Str("Path", path).
			Msg("Reading config from path")
		d, err := os.ReadFile(path)
		if err != nil {
			return nil, errors.Wrap(err, ErrReadSethConfig)
		}
		err = toml.Unmarshal(d, &cfg)
		if err != nil {
			return nil, errors.Wrap(err, ErrUnmarshalSethConfig)
		}
	}
	L.Debug().
		Interface("Config", cfg).
		Msg("Configuration loaded")
	cfg.Havoc.Grafana.URL = os.Getenv("GRAFANA_URL")
	cfg.Havoc.Grafana.Token = os.Getenv("GRAFANA_TOKEN")
	cfg.Havoc.Grafana.DashboardName = os.Getenv("DASHBOARD_NAME")
	return cfg, nil
}

// nolint
func sliceContains(target string, array []string) bool {
	for _, element := range array {
		if element == target {
			return true
		}
	}
	return false
}

func sliceContainsSubString(target string, array []string) bool {
	for _, element := range array {
		if strings.Contains(target, element) {
			return true
		}
	}
	return false
}
