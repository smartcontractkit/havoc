package havoc

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"os"
	"sort"
	"strings"
)

const (
	ErrNoNamespace    = "no namespace found"
	ErrEmptyNamespace = "no pods found inside namespace, namespace is empty or check your filter"
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
	} `json:"metadata"`
}

// ActionablePodInfo info about pod and labels for which we can generate a chaos experiment
type ActionablePodInfo struct {
	PodName  string
	Labels   []string
	HasGroup bool
}

// splitLabelsIntoGroups splits labels into groups:
// - all discovered groups that have more than 1 pod
// - network partitioning groups (can include groups with only 1 pod)
func (m *Controller) splitLabelsIntoGroups(cfg *Config, onlyPodLabels []string, labelsToPods map[string][]string) (
	[]string,
	[]string,
	map[string][]string,
	[][]string,
) {
	counts := make(map[string]int)
	seen := make(map[string]bool)
	for _, ld := range onlyPodLabels {
		counts[ld]++
	}
	groupLabels := make([]string, 0)
	groupPodNames := make([]string, 0)
	groupLabelToPods := make(map[string][]string)
	groupNetworkPartitionLabels := make([]string, 0)
	for _, ld := range onlyPodLabels {
		if counts[ld] > 1 && !seen[ld] {
			if !sliceContainsSubString(ld, cfg.Havoc.IgnoreGroupLabels) && !strings.Contains(ld, m.cfg.Havoc.NetworkPartition.Label) {
				groupLabels = append(groupLabels, ld)
				groupLabelToPods[ld] = labelsToPods[ld]
				groupPodNames = append(groupPodNames, groupLabelToPods[ld]...)
				L.Info().
					Str("Label", ld).
					Int("Count", counts[ld]).
					Strs("Pods", labelsToPods[ld]).
					Msg("New group found")
			}
		}
		if strings.Contains(ld, m.cfg.Havoc.NetworkPartition.Label) && !seen[ld] {
			L.Info().
				Str("Label", ld).
				Int("Count", counts[ld]).
				Strs("Pods", labelsToPods[ld]).
				Msg("New group found")
			groupNetworkPartitionLabels = append(groupNetworkPartitionLabels, ld)
		}
		seen[ld] = true
	}
	sort.Slice(groupLabels, func(i, j int) bool {
		return groupLabels[i] < groupLabels[j]
	})
	return groupLabels, groupPodNames, groupLabelToPods, uniquePairs(groupNetworkPartitionLabels)
}

func uniquePairs(strings []string) [][]string {
	var pairs [][]string
	for i := 0; i < len(strings); i++ {
		for j := i + 1; j < len(strings); j++ {
			pair := []string{strings[i], strings[j]}
			pairs = append(pairs, pair)
		}
	}
	return pairs
}

// processPodInfo parses pods call response and returns:
// pods with all associated labels
// group labels and count of pods affected
// TODO: refactor with samber/lo, too complex. Filter, Map, GroupBy..
func (m *Controller) processPodInfo(cfg *Config, mfp *PodsListResponse) ([]*ActionablePodInfo, []string, [][]string, error) {
	L.Info().Msg("Processing pods info")
	validItems := make([]*PodResponse, 0)
	for _, pi := range mfp.Items {
		if !sliceContainsSubString(pi.Metadata.Name, m.cfg.Havoc.IgnoredPods) {
			validItems = append(validItems, pi)
		}
	}
	mfp.Items = validItems
	if len(mfp.Items) == 0 {
		return nil, nil, nil, errors.New(ErrEmptyNamespace)
	}

	allPodsDataWithLabels := make([]*ActionablePodInfo, 0)
	onlyLabels := make([]string, 0)
	labelsToPods := make(map[string][]string)
	for _, p := range mfp.Items {
		api := &ActionablePodInfo{
			PodName: p.Metadata.Name,
		}
		for labelKey, labelValue := range p.Metadata.Labels {
			l := fmt.Sprintf("'%s': '%s'", labelKey, labelValue)
			api.Labels = append(api.Labels, l)
			if !sliceContains(l, m.cfg.Havoc.IgnoreGroupLabels) {
				if labelsToPods[l] == nil {
					labelsToPods[l] = make([]string, 0)
				}
				labelsToPods[l] = append(labelsToPods[l], p.Metadata.Name)
			}
		}
		allPodsDataWithLabels = append(allPodsDataWithLabels, api)
		onlyLabels = append(onlyLabels, api.Labels...)
	}
	groupLabels, groupPodNames, _, networkGroupLabels := m.splitLabelsIntoGroups(cfg, onlyLabels, labelsToPods)

	podsWithoutGroup := make([]*ActionablePodInfo, 0)
	for _, p := range allPodsDataWithLabels {
		if !sliceContains(p.PodName, groupPodNames) {
			L.Info().Str("Pod", p.PodName).Msg("Pod doesn't have a group")
			podsWithoutGroup = append(podsWithoutGroup, p)
		}
	}

	sort.Slice(allPodsDataWithLabels, func(i, j int) bool { return allPodsDataWithLabels[i].PodName < allPodsDataWithLabels[j].PodName })
	return podsWithoutGroup, groupLabels, networkGroupLabels, nil
}

// GetPodsInfo gets info about all the pods in the namespace
func (m *Controller) GetPodsInfo(namespace string) (*PodsListResponse, error) {
	if _, err := ExecCmd(fmt.Sprintf("kubectl get ns %s", namespace)); err != nil {
		return nil, errors.Wrap(errors.New(ErrNoNamespace), namespace)
	}
	var cmdBuilder strings.Builder
	cmdBuilder.Write([]byte(fmt.Sprintf("kubectl get pods -n %s ", namespace)))
	if m.cfg.Havoc.NamespaceLabelFilter != "" {
		cmdBuilder.Write([]byte(fmt.Sprintf("-l %s ", m.cfg.Havoc.NamespaceLabelFilter)))
	}
	cmdBuilder.Write([]byte("-o json"))
	out, err := ExecCmd(cmdBuilder.String())
	if err != nil {
		return nil, err
	}
	if err := dumpPodInfo(out); err != nil {
		return nil, err
	}
	var pr *PodsListResponse
	if err := json.Unmarshal([]byte(out), &pr); err != nil {
		return nil, err
	}
	return pr, nil
}

func dumpPodInfo(out string) error {
	if L.GetLevel() == zerolog.DebugLevel {
		var plr *PodsListResponse
		if err := json.Unmarshal([]byte(out), &plr); err != nil {
			return err
		}
		d, err := json.Marshal(plr)
		if err != nil {
			return err
		}
		_ = os.WriteFile("pods_dump.json", d, os.ModePerm)
		return nil
	}
	return nil
}
