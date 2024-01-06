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

func TestSmokeParsingGeneratingEndToEnd(t *testing.T) {
	InitDefaultLogging()
	d, err := os.ReadFile("testdata/deployment_crib_1.yaml")
	require.NoError(t, err)
	var depls *PodsListResponse
	err = json.Unmarshal(d, &depls)
	require.NoError(t, err)
	_, _, err = generateSpecs("my-test-namespace", TestExperimentsDir, depls, nil)
	require.NoError(t, err)
	snapshotData, err := ReadExperimentsFromDir(RecommendedExperimentTypes, SnapshotDir)
	require.NoError(t, err)
	generatedData, err := ReadExperimentsFromDir(RecommendedExperimentTypes, TestExperimentsDir)
	require.NoError(t, err)
	for i := range snapshotData {
		require.Equal(t, snapshotData[i], generatedData[i])
	}
}
