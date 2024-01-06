package havoc

import (
	"github.com/smartcontractkit/havoc"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestUsingRecommendedExperiments(t *testing.T) {
	myExperimentsDir := "my-experiments"
	// config can be nil, then default will be used, that's just an example
	err := havoc.GenerateSpecs(
		"my-namespace",
		myExperimentsDir,
		&havoc.Config{
			Havoc: &havoc.Havoc{
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
		},
	)
	/*
		your test logic here
	*/
	require.NoError(t, err)
	err = havoc.ApplyChaosFile(myExperimentsDir, "failure", "app-node-3.yaml", true)
	require.NoError(t, err)
	err = havoc.ApplyChaosFile(myExperimentsDir, "latency", "app-node-3.yaml", true)
	require.NoError(t, err)
	/*
		your verification logic here
	*/
}

func TestGenerating(t *testing.T) {
	cfg := &havoc.Config{
		Havoc: &havoc.Havoc{
			IgnoredPods: []string{"geth", "mockserver", "-db-"},
			Failure: &havoc.Failure{
				Duration:        "5s",
				GroupPercentage: "0.3",
			},
			Latency: &havoc.Latency{
				Duration:        "5s",
				GroupPercentage: "0.3",
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
	err := havoc.GenerateSpecs(
		"skudasov-crib",
		"test-generating-dir",
		cfg,
	)
	require.NoError(t, err)
}

func TestUsingSequentialMonkey(t *testing.T) {
	//havoc.SetGlobalLogger(your zerolog here...)
	myExperimentsDir := "sequential-monkey"
	cfg := &havoc.Config{
		Havoc: &havoc.Havoc{
			Failure: &havoc.Failure{
				Duration:        "5s",
				GroupPercentage: "0.3",
			},
			Latency: &havoc.Latency{
				Duration:        "5s",
				GroupPercentage: "0.3",
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
				Dir:                     myExperimentsDir,
				Mode:                    "seq",
				MaxSimultaneousFailures: 2,
				GrafanaURL:              os.Getenv("GRAFANA_URL"),
				GrafanaToken:            os.Getenv("GRAFANA_TOKEN"),
				DashboardName:           os.Getenv("DASHBOARD_NAME"),
			},
		},
	}

	err := havoc.GenerateSpecs(
		"skudasov-crib",
		myExperimentsDir,
		cfg,
	)
	require.NoError(t, err)
	m, err := havoc.NewMonkey(cfg)
	require.NoError(t, err)
	err = m.Run(nil)
	require.NoError(t, err)
}
