package havoc

import (
	"github.com/smartcontractkit/havoc"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestUsingRecommendedExperiments(t *testing.T) {
	cfg := &havoc.Config{
		Havoc: &havoc.Havoc{
			Dir: "my-experiments",
			Failure: &havoc.Failure{
				Duration: "1m",
			},
			Latency: &havoc.Latency{
				Duration: "1m",
				Latency:  "300ms",
			},
			StressMemory: &havoc.StressMemory{
				Duration: "1m",
				Workers:  4,
				Memory:   "512MB",
			},
			StressCPU: &havoc.StressCPU{
				Duration: "1m",
				Workers:  1,
				Load:     100,
			},
		},
	}
	c, err := havoc.NewController(cfg)
	require.NoError(t, err)
	err = c.GenerateSpecs("skudasov-crib")
	require.NoError(t, err)
	/*
		your test logic here
	*/
	err = c.ApplyChaosFile("failure", "app-node-3.yaml", true)
	require.NoError(t, err)
	err = c.ApplyChaosFile("latency", "app-node-3.yaml", true)
	require.NoError(t, err)
	/*
		your verification logic here
	*/
}

func TestGenerating(t *testing.T) {
	cfg := &havoc.Config{
		Havoc: &havoc.Havoc{
			Dir:         "my-experiments",
			IgnoredPods: []string{"geth", "mockserver", "-db-"},
			Failure: &havoc.Failure{
				Duration:        "5s",
				GroupPercentage: []string{"30"},
			},
			Latency: &havoc.Latency{
				Duration:        "5s",
				GroupPercentage: []string{"30"},
				Latency:         "300ms",
			},
			StressMemory: &havoc.StressMemory{
				Duration: "5s",
				Workers:  4,
				Memory:   "512MB",
			},
			StressCPU: &havoc.StressCPU{
				Duration: "5s",
				Workers:  1,
				Load:     100,
			},
		},
	}
	c, err := havoc.NewController(cfg)
	require.NoError(t, err)
	err = c.GenerateSpecs("skudasov-crib")
	require.NoError(t, err)
}

func TestCommonIntegrationWithLoadTool(t *testing.T) {
	//havoc.SetGlobalLogger(your zerolog here...)
	cfg := &havoc.Config{
		Havoc: &havoc.Havoc{
			Dir: "my-experiments",
			Failure: &havoc.Failure{
				Duration:        "5s",
				GroupPercentage: []string{"30"},
			},
			Latency: &havoc.Latency{
				Duration:        "5s",
				GroupPercentage: []string{"30"},
				Latency:         "300ms",
			},
			StressMemory: &havoc.StressMemory{
				Duration: "5s",
				Workers:  4,
				Memory:   "512MB",
			},
			StressCPU: &havoc.StressCPU{
				Duration: "5s",
				Workers:  1,
				Load:     100,
			},
			Monkey: &havoc.Monkey{
				Duration:                "7m",
				Cooldown:                "10s",
				Mode:                    "seq",
				MaxSimultaneousFailures: 2,
			},
			Grafana: &havoc.Grafana{
				URL:           os.Getenv("GRAFANA_URL"),
				Token:         os.Getenv("GRAFANA_TOKEN"),
				DashboardName: os.Getenv("DASHBOARD_NAME"),
			},
		},
	}

	m, err := havoc.NewController(cfg)
	require.NoError(t, err)
	err = m.GenerateSpecs("skudasov-crib")
	require.NoError(t, err)
	go func() {
		err = m.Run()
		require.NoError(t, err)
	}()
	time.Sleep(1 * time.Minute)
	m.Stop()
}
