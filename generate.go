package havoc

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

const (
	ErrParsingTemplate = "failed to parse Go text template"

	ErrExperimentTimeout = "waiting for experiment to finish timed out"
	ErrExperimentApply   = "error applying experiment manifest"
)

var (
	RecommendedExperimentTypes = []string{
		ChaosTypeFailure,
		ChaosTypeGroupFailure,
		ChaosTypeLatency,
		ChaosTypeGroupLatency,
		ChaosTypeStressMemory,
		ChaosTypeStressCPU,
		ChaosTypePartitionExternal,
	}
)

// MarshalTemplate Helper to marshal templates
func MarshalTemplate(jobSpec interface{}, name, templateString string) (string, error) {
	var buf bytes.Buffer
	tmpl, err := template.New(name).Parse(templateString)
	if err != nil {
		return "", errors.Wrap(err, ErrParsingTemplate)
	}
	err = tmpl.Execute(&buf, jobSpec)
	if err != nil {
		return "", err
	}
	return buf.String(), err
}

type NetworkChaosExperiment struct {
	ExperimentName  string
	Namespace       string
	WaitLabel       string
	Duration        string
	Latency         string
	PodName         string
	GroupPercentage string
	Selector        string
}

func (m NetworkChaosExperiment) String() (string, error) {
	tpl := `
kind: NetworkChaos
apiVersion: chaos-mesh.org/v1alpha1
metadata:
  name: {{ .ExperimentName }}
  namespace: {{ .Namespace }}
  labels:
    waitLabel: {{ .WaitLabel }}
spec:
  selector:
    namespaces:
      - {{ .Namespace }}
    {{- if .Selector}}
    labelSelectors:
      {{ .Selector }}
	{{- else}}
    fieldSelectors:
      metadata.name: {{ .PodName }}	
	{{- end}}
  {{- if .Selector }}
  mode: fixed-percent
  value: '{{ .GroupPercentage }}'
  {{- else }}
  mode: one
  {{- end }}
  action: delay
  duration: {{ .Duration }}
  delay:
    latency: {{ .Latency }}
  direction: from
  target:
    selector:
      namespaces:
        - {{ .Namespace }}
      {{- if .Selector}}
      labelSelectors:
        {{ .Selector }}
	  {{- else}}
      fieldSelectors:
        metadata.name: {{ .PodName }}	
	  {{- end}}
    mode: all
`
	return MarshalTemplate(
		m,
		uuid.NewString(),
		tpl,
	)
}

type NetworkChaosExternalPartitionExperiment struct {
	ExperimentName string
	Namespace      string
	WaitLabel      string
	Duration       string
	PodName        string
	ExternalURL    string
}

func (m NetworkChaosExternalPartitionExperiment) String() (string, error) {
	tpl := `
kind: NetworkChaos
apiVersion: chaos-mesh.org/v1alpha1
metadata:
  name: {{ .ExperimentName }}
  namespace: {{ .Namespace }}
  labels:
    waitLabel: {{ .WaitLabel }}
spec:
  selector:
    namespaces:
      - {{ .Namespace }}
  mode: all
  action: partition
  duration: {{ .Duration }}
  direction: to
  target:
    selector:
      namespaces:
        - {{ .Namespace }}
    mode: all
  externalTargets:
    - {{ .ExternalURL }}
`
	return MarshalTemplate(
		m,
		uuid.NewString(),
		tpl,
	)
}

type PodFailureExperiment struct {
	ExperimentName  string
	Namespace       string
	WaitLabel       string
	Duration        string
	PodName         string
	GroupPercentage string
	Selector        string
}

func (m PodFailureExperiment) String() (string, error) {
	tpl := `
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: {{ .ExperimentName }}
  namespace: {{ .Namespace }}
  labels:
    waitLabel: {{ .WaitLabel }}
spec:
  action: pod-failure
  {{- if .Selector }}
  mode: fixed-percent
  value: '{{ .GroupPercentage }}'
  {{- else }}
  mode: one
  {{- end }}
  duration: {{ .Duration }}
  selector:
    {{- if .Selector}}
    labelSelectors:
      {{ .Selector }}
	{{- else}}
    fieldSelectors:
      metadata.name: {{ .PodName }}	
	{{- end}}
`
	return MarshalTemplate(
		m,
		uuid.NewString(),
		tpl,
	)
}

type PodStressCPUExperiment struct {
	ExperimentName string
	Namespace      string
	WaitLabel      string
	Workers        int
	Load           int
	Duration       string
	PodName        string
	Selector       string
}

func (m PodStressCPUExperiment) String() (string, error) {
	tpl := `
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: {{ .ExperimentName }}
  namespace: {{ .Namespace }}
  labels:
    waitLabel: {{ .WaitLabel }}
spec:
  mode: one
  duration: {{ .Duration }}
  selector:
    {{- if .Selector}}
    labelSelectors:
      {{ .Selector }}
	{{- else}}
    fieldSelectors:
      metadata.name: {{ .PodName }}	
	{{- end}}
  stressors:
    cpu:
      workers: {{ .Workers }}
      load: {{ .Load }}
`
	return MarshalTemplate(
		m,
		uuid.NewString(),
		tpl,
	)
}

type PodStressMemoryExperiment struct {
	ExperimentName string
	Namespace      string
	WaitLabel      string
	Workers        int
	Memory         string
	Duration       string
	PodName        string
	Selector       string
}

func (m PodStressMemoryExperiment) String() (string, error) {
	tpl := `
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata:
  name: {{ .ExperimentName }}
  namespace: {{ .Namespace }}
  labels:
    waitLabel: {{ .WaitLabel }}
spec:
  mode: one
  duration: {{ .Duration }}
  selector:
    {{- if .Selector}}
    labelSelectors:
      {{ .Selector }}
	{{- else}}
    fieldSelectors:
      metadata.name: {{ .PodName }}	
	{{- end}}
  stressors:
    memory:
      workers: {{ .Workers }}
      size: {{ .Memory }}
`
	return MarshalTemplate(
		m,
		uuid.NewString(),
		tpl,
	)
}

type NamedExperiment struct {
	Name     string
	Type     string
	Manifest string
}

func (m *Controller) ReadExperimentsFromDir(expTypes []string, dir string) ([]*NamedExperiment, error) {
	expData := make([]*NamedExperiment, 0)
	for _, expType := range expTypes {
		targetDir := fmt.Sprintf("%s/%s", dir, expType)
		if _, err := os.Stat(targetDir); err != nil {
			log.Warn().
				Str("Dir", targetDir).
				Msg("Experiments dir not found, skipping")
			return nil, nil
		}
		err := filepath.Walk(
			fmt.Sprintf("%s/%s", dir, expType),
			func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				expData = append(expData, &NamedExperiment{
					Name:     info.Name(),
					Type:     expType,
					Manifest: string(data),
				})
				return err
			})
		if err != nil {
			return nil, err
		}
	}
	return expData, nil
}

func (m *Controller) generatePodFailureExperiments(namespace string, podInfo []*ActionablePodInfo, cfg *Config) (map[string]string, error) {
	ma := make(map[string]string)
	for _, pi := range podInfo {
		ph, err := PodFailureExperiment{
			Namespace:      namespace,
			ExperimentName: fmt.Sprintf("%s-%s", ChaosTypeFailure, pi.PodName),
			WaitLabel:      pi.PodName,
			Duration:       cfg.Havoc.Failure.Duration,
			PodName:        pi.PodName,
		}.String()
		if err != nil {
			return nil, err
		}
		ma[pi.PodName] = ph
	}
	return ma, nil
}

func (m *Controller) generatePodFailureGroupExperiments(namespace string, groupLabels []string, cfg *Config) (map[string]string, error) {
	ma := make(map[string]string)
	for _, label := range groupLabels {
		sanitizedLabel := sanitizeLabel(label)
		ph, err := PodFailureExperiment{
			Namespace:       namespace,
			ExperimentName:  fmt.Sprintf("%s-%s", ChaosTypeGroupFailure, sanitizedLabel),
			WaitLabel:       sanitizedLabel,
			Duration:        cfg.Havoc.Failure.Duration,
			GroupPercentage: cfg.Havoc.Failure.GroupPercentage,
			Selector:        label,
		}.String()
		if err != nil {
			return nil, err
		}
		ma[sanitizedLabel] = ph
	}
	return ma, nil
}

func (m *Controller) generatePodLatenciesExperiments(namespace string, lfd []*ActionablePodInfo, cfg *Config) (map[string]string, error) {
	ma := make(map[string]string)
	for _, mfp := range lfd {
		pl, err := NetworkChaosExperiment{
			Namespace:      namespace,
			ExperimentName: fmt.Sprintf("%s-%s", ChaosTypeLatency, mfp.PodName),
			WaitLabel:      mfp.PodName,
			Duration:       cfg.Havoc.Latency.Duration,
			Latency:        cfg.Havoc.Latency.Latency,
			PodName:        mfp.PodName,
		}.String()
		if err != nil {
			return nil, err
		}
		ma[mfp.PodName] = pl
	}
	return ma, nil
}

func (m *Controller) generatePodLatencyGroupExperiments(namespace string, groupLabels []string, cfg *Config) (map[string]string, error) {
	ma := make(map[string]string)
	for _, label := range groupLabels {
		sanitizedLabel := sanitizeLabel(label)
		ph, err := NetworkChaosExperiment{
			Namespace:       namespace,
			ExperimentName:  fmt.Sprintf("%s-%s", ChaosTypeGroupLatency, sanitizedLabel),
			WaitLabel:       sanitizedLabel,
			Duration:        cfg.Havoc.Latency.Duration,
			GroupPercentage: cfg.Havoc.Latency.GroupPercentage,
			Latency:         cfg.Havoc.Latency.Latency,
			Selector:        label,
		}.String()
		if err != nil {
			return nil, err
		}
		ma[sanitizedLabel] = ph
	}
	return ma, nil
}

func (m *Controller) generatePodStressMemoryExperiments(namespace string, lfd []*ActionablePodInfo, cfg *Config) (map[string]string, error) {
	ma := make(map[string]string)
	for _, mfp := range lfd {
		ph, err := PodStressMemoryExperiment{
			Namespace:      namespace,
			ExperimentName: fmt.Sprintf("%s-%s", ChaosTypeStressMemory, mfp.PodName),
			WaitLabel:      mfp.PodName,
			Duration:       cfg.Havoc.StressMemory.Duration,
			Workers:        cfg.Havoc.StressMemory.Workers,
			Memory:         cfg.Havoc.StressMemory.Memory,
			PodName:        mfp.PodName,
		}.String()
		if err != nil {
			return nil, err
		}
		ma[mfp.PodName] = ph
	}
	return ma, nil
}

func (m *Controller) generatePodStressCPUExperiments(namespace string, lfd []*ActionablePodInfo, cfg *Config) (map[string]string, error) {
	ma := make(map[string]string)
	for _, mfp := range lfd {
		ph, err := PodStressCPUExperiment{
			Namespace:      namespace,
			ExperimentName: fmt.Sprintf("%s-%s", ChaosTypeStressCPU, mfp.PodName),
			WaitLabel:      mfp.PodName,
			Duration:       cfg.Havoc.StressCPU.Duration,
			Workers:        cfg.Havoc.StressCPU.Workers,
			Load:           cfg.Havoc.StressCPU.Load,
			PodName:        mfp.PodName,
		}.String()
		if err != nil {
			return nil, err
		}
		ma[mfp.PodName] = ph
	}
	return ma, nil
}

func (m *Controller) generateExternalTargetsPartitions(namespace string, cfg *Config) (map[string]string, error) {
	ma := make(map[string]string)
	if cfg.Havoc.ExternalTargets == nil {
		return nil, nil
	}
	for _, u := range cfg.Havoc.ExternalTargets.URLs {
		nsAndURLHash := fmt.Sprintf("%s-%s", namespace, urlHash(u))
		ph, err := NetworkChaosExternalPartitionExperiment{
			Namespace:      namespace,
			ExperimentName: fmt.Sprintf("%s-%s", ChaosTypePartitionExternal, nsAndURLHash),
			WaitLabel:      nsAndURLHash,
			Duration:       cfg.Havoc.ExternalTargets.Duration,
			ExternalURL:    fmt.Sprintf("'%s'", u),
		}.String()
		if err != nil {
			return nil, err
		}
		ma[nsAndURLHash] = ph
	}
	return ma, nil
}

func urlHash(url string) string {
	hasher := md5.New()
	hasher.Write([]byte(url))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

func sanitizeLabel(label string) string {
	sanitizedLabel := strings.Replace(label, "'", "", -1)
	sanitizedLabel = strings.Replace(sanitizedLabel, ": ", "-", -1)
	sanitizedLabel = strings.Replace(sanitizedLabel, ".", "-", -1)
	sanitizedLabel = strings.Replace(sanitizedLabel, "/", "-", -1)
	return sanitizedLabel
}

type EventJSONItemResponse struct {
	APIVersion     string    `json:"apiVersion"`
	Count          int       `json:"count"`
	EventTime      any       `json:"eventTime"`
	FirstTimestamp time.Time `json:"firstTimestamp"`
	InvolvedObject struct {
		APIVersion      string `json:"apiVersion"`
		Kind            string `json:"kind"`
		Name            string `json:"name"`
		Namespace       string `json:"namespace"`
		ResourceVersion string `json:"resourceVersion"`
		UID             string `json:"uid"`
	} `json:"involvedObject"`
	Kind          string    `json:"kind"`
	LastTimestamp time.Time `json:"lastTimestamp"`
	Message       string    `json:"message"`
	Metadata      struct {
		Annotations struct {
			ChaosMeshOrgType string `json:"chaos-mesh.org/type"`
		} `json:"annotations"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
		Name              string    `json:"name"`
		Namespace         string    `json:"namespace"`
		ResourceVersion   string    `json:"resourceVersion"`
		UID               string    `json:"uid"`
	} `json:"metadata"`
	Reason             string `json:"reason"`
	ReportingComponent string `json:"reportingComponent"`
	ReportingInstance  string `json:"reportingInstance"`
	Source             struct {
		Component string `json:"component"`
	} `json:"source"`
	Type string `json:"type"`
}

type EventsJSONResponse struct {
	APIVersion string                   `json:"apiVersion"`
	Items      []*EventJSONItemResponse `json:"items"`
	Kind       string                   `json:"kind"`
	Metadata   struct {
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
}

func eventsForLastMinutes(out string, timeOfApplication time.Time) error {
	var d *EventsJSONResponse
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		return err
	}
	L.Debug().Msg("Listing all experiment events")
	for _, i := range d.Items {
		if i.LastTimestamp.After(timeOfApplication) {
			L.Info().
				Time("Time", i.LastTimestamp).
				Str("Reason", i.Reason).
				Str("Message", i.Message).
				Send()
		}
	}
	return nil
}

func (m *Controller) ApplyChaosFile(chaosType string, expName string, wait bool) error {
	timeOfApplication := time.Now()
	var errDefer error
	data, err := os.ReadFile(filepath.Join(m.cfg.Havoc.Dir, chaosType, expName))
	if err != nil {
		return err
	}
	L.Info().
		Str("Dir", m.cfg.Havoc.Dir).
		Str("Type", chaosType).
		Str("Name", expName).
		Msg("Applying experiment manifest")
	fmt.Println(string(data))
	_, err = ExecCmd(fmt.Sprintf("kubectl apply -f %s/%s/%s", m.cfg.Havoc.Dir, chaosType, expName))
	if err != nil {
		return errors.Wrap(err, ErrExperimentApply)
	}
	chaosFilenameParts := strings.Split(expName, ".")
	if wait {
		// we delete only if we wait for experiments, otherwise we don't know if it's safe to delete
		// or we can't wait for experiment to end
		defer func() {
			expName = strings.Replace(expName, ".yaml", "", -1)
			var out string
			out, errDefer = ExecCmd(
				fmt.Sprintf("kubectl get events --field-selector involvedObject.name=%s-%s -o json",
					chaosType,
					expName,
				))
			errDefer = eventsForLastMinutes(out, timeOfApplication)
			_, errDefer = ExecCmd(fmt.Sprintf("kubectl delete %s %s-%s", ExperimentsToCRDs[chaosType], chaosType, expName))
		}()
		_, err = ExecCmd(
			fmt.Sprintf("kubectl wait %s -l waitLabel=%s --for condition=AllRecovered=True --timeout %s",
				ExperimentsToCRDs[chaosType],
				chaosFilenameParts[0],
				DefaultCMDTimeout,
			))
		if err != nil {
			return errors.Wrap(err, ErrExperimentTimeout)
		}
		L.Info().Msg("Chaos experiment successfully recovered")
	}
	return errDefer
}

// GenerateSpecs generates specs from namespace, should be used programmatically in tests
func (m *Controller) GenerateSpecs(ns string) error {
	podsInfo, err := m.GetPodsInfo(ns)
	if err != nil {
		return err
	}
	_, _, err = m.generateSpecs(ns, podsInfo)
	return err
}

func (m *Controller) generateSpecs(namespace string, podListResponse *PodsListResponse) (*ChaosSpecs, []*ActionablePodInfo, error) {
	L.Trace().
		Interface("PodListResponse", podListResponse).
		Msg("Deployments manifest from the cluster")
	podInfo, groupLabels := m.processPodInfo(m.cfg, podListResponse)
	L.Info().Msg("Generating chaos experiments")
	podFailures, err := m.generatePodFailureExperiments(namespace, podInfo, m.cfg)
	if err != nil {
		return nil, nil, err
	}
	podLatencies, err := m.generatePodLatenciesExperiments(namespace, podInfo, m.cfg)
	if err != nil {
		return nil, nil, err
	}
	podMem, err := m.generatePodStressMemoryExperiments(namespace, podInfo, m.cfg)
	if err != nil {
		return nil, nil, err
	}
	podCPU, err := m.generatePodStressCPUExperiments(namespace, podInfo, m.cfg)
	if err != nil {
		return nil, nil, err
	}
	partExt, err := m.generateExternalTargetsPartitions(namespace, m.cfg)
	if err != nil {
		return nil, nil, err
	}
	L.Info().Msg("Generating group chaos experiments")
	podFailureGroups, err := m.generatePodFailureGroupExperiments(namespace, groupLabels, m.cfg)
	if err != nil {
		return nil, nil, err
	}
	podLatenciesGroup, err := m.generatePodLatencyGroupExperiments(namespace, groupLabels, m.cfg)
	if err != nil {
		return nil, nil, err
	}
	csp := &ChaosSpecs{
		PodFailures:                podFailures,
		PodFailureGroups:           podFailureGroups,
		PodLatencies:               podLatencies,
		PodLatencyGroups:           podLatenciesGroup,
		PodStressMemory:            podMem,
		PodStressCPU:               podCPU,
		NamespacePartitionExternal: partExt,
	}
	return csp, podInfo, csp.Dump(m.cfg.Havoc.Dir)
}
