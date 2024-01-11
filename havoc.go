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
	ExperimentsByType map[string]map[string]string
}

func (m *ChaosSpecs) Dump(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.Mkdir(dir, os.ModePerm); err != nil {
		return err
	}
	L.Info().Str("Dir", dir).Msg("Writing experiments to a dir")
	for expType := range m.ExperimentsByType {
		if err := os.Mkdir(fmt.Sprintf("%s/%s", dir, expType), os.ModePerm); err != nil {
			return err
		}
		for expName, expBody := range m.ExperimentsByType[expType] {
			if err := os.WriteFile(
				fmt.Sprintf("%s/%s/%s.yaml", dir, expType, expName),
				[]byte(expBody),
				os.ModePerm,
			); err != nil {
				return err
			}
		}
	}
	return nil
}
