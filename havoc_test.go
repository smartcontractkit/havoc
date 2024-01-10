package havoc

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	SnapshotDir        = "testdata/experiments-snapshot"
	TestExperimentsDir = "testdata/experiments-test"
)

func TestSmokeParsingGenerating(t *testing.T) {
	InitDefaultLogging()
	d, err := os.ReadFile("testdata/deployment_crib_1.yaml")
	require.NoError(t, err)
	var depls *PodsListResponse
	err = json.Unmarshal(d, &depls)
	require.NoError(t, err)
	m, err := NewController(nil)
	require.NoError(t, err)
	// namespace doesn't matter here, we are checking that generated files are valid
	_, _, err = m.generateSpecs("my-test-namespace", depls)
	require.NoError(t, err)
	snapshotData, err := m.ReadExperimentsFromDir(RecommendedExperimentTypes, SnapshotDir)
	require.NoError(t, err)
	generatedData, err := m.ReadExperimentsFromDir(RecommendedExperimentTypes, TestExperimentsDir)
	require.NoError(t, err)
	for i := range snapshotData {
		require.Equal(t, snapshotData[i], generatedData[i])
	}
}
