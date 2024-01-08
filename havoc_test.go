package havoc

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

var (
	// We are not testing with real k8s, namespace is just a placeholder that should match in snapshots/results
	Namespace      = "cl-cluster"
	TestDataDir    = "testdata"
	SnapshotDir    = filepath.Join(TestDataDir, "snapshot")
	ResultsDir     = filepath.Join(TestDataDir, "results")
	DeploymentsDir = filepath.Join(TestDataDir, "deployments")
	ConfigsDir     = filepath.Join(TestDataDir, "configs")
)

func init() {
	InitDefaultLogging()
}

func setup(t *testing.T, podsInfoPath string, configPath string, resultsDir string) (*Controller, *PodsListResponse) {
	d, err := os.ReadFile(filepath.Join(DeploymentsDir, podsInfoPath))
	require.NoError(t, err)
	var plr *PodsListResponse
	err = json.Unmarshal(d, &plr)
	require.NoError(t, err)
	var cfg *Config
	if configPath != "" {
		cfg, err = ReadConfig(filepath.Join(ConfigsDir, configPath))
		require.NoError(t, err)
	} else {
		cfg = DefaultConfig()
		cfg.Havoc.Dir = filepath.Join(ResultsDir, resultsDir)
	}
	m, err := NewController(cfg)
	require.NoError(t, err)
	return m, plr
}

func TestSmokeParsingGenerating(t *testing.T) {
	type test struct {
		name         string
		podsDumpName string
		configName   string
		snapshotDir  string
		resultsDir   string
	}
	tests := []test{
		{
			name:         "can generate for 1 pod without groups",
			podsDumpName: "deployment_single_pod.json",
			configName:   "",
			snapshotDir:  "single_pod",
			resultsDir:   "single_pod",
		},
		{
			name:         "can generate for an arbitrary component group",
			podsDumpName: "deployment_single_group.json",
			configName:   "",
			snapshotDir:  "single_group",
			resultsDir:   "single_group",
		},
		{
			name:         "can generate for several component groups",
			podsDumpName: "deployment_two_groups.json",
			configName:   "",
			snapshotDir:  "two_groups",
			resultsDir:   "two_groups",
		},
		{
			name:         "different experiments should be generated for two components groups and some standalone pods",
			podsDumpName: "deployment_two_groups_and_standalone.json",
			configName:   "",
			snapshotDir:  "two_groups_plus_standalone_no_labels",
			resultsDir:   "two_groups_plus_standalone_no_labels",
		},
		{
			name:         "must count only groups of 2+ pods, even with common keys in labels",
			podsDumpName: "ignoring_one_pod_group.json",
			configName:   "",
			snapshotDir:  "ignoring_one_pod_group",
			resultsDir:   "ignoring_one_pod_group",
		},
		{
			name:         "can generate 2 component groups, standalones and 4 network groups",
			podsDumpName: "deployment_crib_1.json",
			configName:   "crib.toml",
			snapshotDir:  "default",
			resultsDir:   "default",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, plr := setup(t, tc.podsDumpName, tc.configName, tc.resultsDir)
			_, _, err := m.generateSpecs(Namespace, plr)
			require.NoError(t, err)
			snapshotData, err := m.ReadExperimentsFromDir(RecommendedExperimentTypes, filepath.Join(SnapshotDir, tc.snapshotDir))
			require.NoError(t, err)
			generatedData, err := m.ReadExperimentsFromDir(RecommendedExperimentTypes, filepath.Join(ResultsDir, tc.resultsDir))
			require.NoError(t, err)
			require.Equal(t, len(snapshotData), len(generatedData))
			for i := range snapshotData {
				require.Equal(t, snapshotData[i], generatedData[i])
			}
		})
	}
}

// That's just an easy way to enter debug with arbitrary config, run it manually
func TestManualGenerate(t *testing.T) {
	cfg, err := ReadConfig("havoc.toml")
	require.NoError(t, err)
	m, err := NewController(cfg)
	require.NoError(t, err)
	err = m.GenerateSpecs("cl-cluster")
	require.NoError(t, err)
}
