package havoc

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"sort"
)

const (
	ErrNoNamespace = "no namespace found"
)

type ManifestPart struct {
	Kind                string
	Name                string
	LabelSelectors      []*ActionablePodInfo
	AnnotationSelectors []*ActionablePodInfo
	FlattenedManifest   map[string]interface{}
}

// PodsListResponse pod list response from kubectl in JSON
type PodsListResponse struct {
	Items []*PodResponse `json:"items"`
}

// PodResponse pod info response from kubectl in JSON
type PodResponse struct {
	Metadata struct {
		Name   string            `json:"name"`
		Labels map[string]string `json:"labels"`
	}
}

// ActionablePodInfo info about pod and labels for which we can generate a chaos experiment
type ActionablePodInfo struct {
	PodName string
	Labels  []string
}

// groupLabels generates an array of labels which are present on more than one pod
// returns these labels and counts of how many pods are in the group
func (m *Controller) groupLabels(cfg *Config, input []string) ([]string, map[string]int) {
	counts := make(map[string]int)
	seen := make(map[string]bool)
	for _, ld := range input {
		counts[ld]++
	}
	groupLabels := make([]string, 0)
	for _, ld := range input {
		if counts[ld] > 1 && !seen[ld] {
			seen[ld] = true
			if len(cfg.Havoc.IgnoreGroupLabels) > 0 && !sliceContainsSubString(ld, cfg.Havoc.IgnoreGroupLabels) {
				groupLabels = append(groupLabels, ld)
			}
		}
	}
	return groupLabels, counts
}

// processPodInfo parses pods call response and returns:
// pods with all associated labels
// group labels and count of pods affected
func (m *Controller) processPodInfo(cfg *Config, mfp *PodsListResponse) ([]*ActionablePodInfo, []string) {
	L.Info().Msg("Processing pods info")
	allPodsDataWithLabels := make([]*ActionablePodInfo, 0)
	onlyLabels := make([]string, 0)
	for _, p := range mfp.Items {
		api := &ActionablePodInfo{
			PodName: p.Metadata.Name,
		}
		for labelKey, labelValue := range p.Metadata.Labels {
			api.Labels = append(api.Labels, fmt.Sprintf("'%s': '%s'", labelKey, labelValue))
		}
		allPodsDataWithLabels = append(allPodsDataWithLabels, api)
		onlyLabels = append(onlyLabels, api.Labels...)
	}
	gl, glCounts := m.groupLabels(cfg, onlyLabels)
	sort.Slice(gl, func(i, j int) bool {
		return gl[i] < gl[j]
	})
	sort.Slice(allPodsDataWithLabels, func(i, j int) bool {
		return allPodsDataWithLabels[i].PodName < allPodsDataWithLabels[j].PodName
	})
	for _, groupLabel := range gl {
		L.Info().
			Int("PodsSelected", glCounts[groupLabel]).
			Str("Label", groupLabel).
			Msg("Group Label")
	}
	return allPodsDataWithLabels, gl
}

// GetPodsInfo gets info about all the pods in the namespace
func (m *Controller) GetPodsInfo(namespace string) (*PodsListResponse, error) {
	if _, err := ExecCmd(fmt.Sprintf("kubectl get ns %s", namespace)); err != nil {
		return nil, errors.Wrap(errors.New(ErrNoNamespace), namespace)
	}
	out, err := ExecCmd(fmt.Sprintf("kubectl get pods -n %s -o json", namespace))
	if err != nil {
		return nil, err
	}
	var pr *PodsListResponse
	if err := json.Unmarshal([]byte(out), &pr); err != nil {
		return nil, err
	}
	if len(m.cfg.Havoc.IgnoredPods) > 0 {
		validItems := make([]*PodResponse, 0)
		for _, pi := range pr.Items {
			if !sliceContainsSubString(pi.Metadata.Name, m.cfg.Havoc.IgnoredPods) {
				validItems = append(validItems, pi)
			}
		}
		pr.Items = validItems
	}
	return pr, nil
}
