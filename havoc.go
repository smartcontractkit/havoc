package havoc

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

const (
	ChaosTypeFailure           = "failure"
	ChaosTypeGroupFailure      = "group-failure"
	ChaosTypeLatency           = "latency"
	ChaosTypeGroupLatency      = "group-latency"
	ChaosTypeStressMemory      = "memory"
	ChaosTypeStressCPU         = "cpu"
	ChaosTypePartitionExternal = "external"
)

var (
	ExperimentsToCRDs = map[string]string{
		ChaosTypeFailure:           "podchaos.chaos-mesh.org",
		ChaosTypeGroupFailure:      "podchaos.chaos-mesh.org",
		ChaosTypeLatency:           "networkchaos.chaos-mesh.org",
		ChaosTypeGroupLatency:      "networkchaos.chaos-mesh.org",
		ChaosTypeStressMemory:      "stresschaos.chaos-mesh.org",
		ChaosTypeStressCPU:         "stresschaos.chaos-mesh.org",
		ChaosTypePartitionExternal: "networkchaos.chaos-mesh.org",
	}
)

var L zerolog.Logger

func SetGlobalLogger(l zerolog.Logger) {
	L = l.With().Str("Component", "havoc").Logger()
}

func InitDefaultLogging() {
	lvl, err := zerolog.ParseLevel(os.Getenv("HAVOC_LOG_LEVEL"))
	if err != nil {
		panic(err)
	}
	if lvl.String() == "" {
		lvl = zerolog.InfoLevel
	}
	L = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(lvl)
}

type ChaosSpecs struct {
	PodFailures                map[string]string
	PodFailureGroups           map[string]string
	PodLatencies               map[string]string
	PodLatencyGroups           map[string]string
	PodStressMemory            map[string]string
	PodStressCPU               map[string]string
	NamespacePartitionExternal map[string]string
}

func (m *ChaosSpecs) Dump(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.Mkdir(dir, os.ModePerm); err != nil {
		return err
	}
	if err := os.Mkdir(fmt.Sprintf("%s/%s", dir, ChaosTypeFailure), os.ModePerm); err != nil {
		return err
	}
	if err := os.Mkdir(fmt.Sprintf("%s/%s", dir, ChaosTypeGroupFailure), os.ModePerm); err != nil {
		return err
	}
	if err := os.Mkdir(fmt.Sprintf("%s/%s", dir, ChaosTypeLatency), os.ModePerm); err != nil {
		return err
	}
	if err := os.Mkdir(fmt.Sprintf("%s/%s", dir, ChaosTypeGroupLatency), os.ModePerm); err != nil {
		return err
	}
	if err := os.Mkdir(fmt.Sprintf("%s/%s", dir, ChaosTypeStressMemory), os.ModePerm); err != nil {
		return err
	}
	if err := os.Mkdir(fmt.Sprintf("%s/%s", dir, ChaosTypeStressCPU), os.ModePerm); err != nil {
		return err
	}
	if err := os.Mkdir(fmt.Sprintf("%s/%s", dir, ChaosTypePartitionExternal), os.ModePerm); err != nil {
		return err
	}
	for expName, expBody := range m.PodFailures {
		if err := os.WriteFile(
			fmt.Sprintf("%s/%s/%s.yaml", dir, ChaosTypeFailure, expName),
			[]byte(expBody),
			os.ModePerm,
		); err != nil {
			return err
		}
	}
	for expName, expBody := range m.PodFailureGroups {
		if err := os.WriteFile(
			fmt.Sprintf("%s/%s/%s.yaml", dir, ChaosTypeGroupFailure, expName),
			[]byte(expBody),
			os.ModePerm,
		); err != nil {
			return err
		}
	}
	for expName, expBody := range m.PodLatencies {
		if err := os.WriteFile(
			fmt.Sprintf("%s/%s/%s.yaml", dir, ChaosTypeLatency, expName),
			[]byte(expBody),
			os.ModePerm,
		); err != nil {
			return err
		}
	}
	for expName, expBody := range m.PodLatencyGroups {
		if err := os.WriteFile(
			fmt.Sprintf("%s/%s/%s.yaml", dir, ChaosTypeGroupLatency, expName),
			[]byte(expBody),
			os.ModePerm,
		); err != nil {
			return err
		}
	}
	for expName, expBody := range m.PodStressMemory {
		if err := os.WriteFile(
			fmt.Sprintf("%s/%s/%s.yaml", dir, ChaosTypeStressMemory, expName),
			[]byte(expBody),
			os.ModePerm,
		); err != nil {
			return err
		}
	}
	for expName, expBody := range m.PodStressCPU {
		if err := os.WriteFile(
			fmt.Sprintf("%s/%s/%s.yaml", dir, ChaosTypeStressCPU, expName),
			[]byte(expBody),
			os.ModePerm,
		); err != nil {
			return err
		}
	}
	for expName, expBody := range m.NamespacePartitionExternal {
		if err := os.WriteFile(
			fmt.Sprintf("%s/%s/%s.yaml", dir, ChaosTypePartitionExternal, expName),
			[]byte(expBody),
			os.ModePerm,
		); err != nil {
			return err
		}
	}
	return nil
}
