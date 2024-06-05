package k8schaos

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/smartcontractkit/chainlink-testing-framework/grafana"
)

type SingleLineGrafanaAnnotator struct {
	client       *grafana.Client
	dashboardUID string
	logger       zerolog.Logger
}

func NewSingleLineGrafanaAnnotator(grafanaURL, grafanaToken, dashboardUID string, logger zerolog.Logger) *SingleLineGrafanaAnnotator {
	return &SingleLineGrafanaAnnotator{
		client:       grafana.NewGrafanaClient(grafanaURL, grafanaToken),
		dashboardUID: dashboardUID,
		logger:       logger,
	}
}

func (l SingleLineGrafanaAnnotator) OnChaosCreated(chaos K8sChaos) {
}

func (l SingleLineGrafanaAnnotator) OnChaosCreationFailed(chaos K8sChaos, reason error) {
}

func (l SingleLineGrafanaAnnotator) OnChaosStarted(chaos K8sChaos) {
	experiment, _ := chaos.GetExperimentStatus()
	duration, _ := chaos.GetChaosDuration()

	var sb strings.Builder
	sb.WriteString("<body>")
	sb.WriteString(fmt.Sprintf("<h4>%s Started</h4>", chaos.GetChaosTypeStr()))
	sb.WriteString(fmt.Sprintf("<div>Name: %s</div>", chaos.Object.GetName()))
	if chaos.Description != "" {
		sb.WriteString(fmt.Sprintf("<div>Description: %s</div>", chaos.Description))
	}
	sb.WriteString(fmt.Sprintf("<div>Start Time: %s</div>", chaos.GetStartTime().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("<div>Duration: %s</div>", duration.String()))

	spec := chaos.GetChaosSpec()
	specBytes, err := json.MarshalIndent(spec, "", "  ")
	if err == nil && len(specBytes) > 0 {
		sb.WriteString("<br>")
		sb.WriteString("<h5>Spec:</h5>")
		sb.WriteString(string(specBytes))
		sb.WriteString("<br>")
	} else {
		l.logger.Warn().Msgf("could not get chaos spec: %s", err)
	}

	if len(experiment.Records) > 0 {
		sb.WriteString("<br>")
		sb.WriteString("<h5>Records:</h5>")
		sb.WriteString("<ul>")
		for _, record := range experiment.Records {
			sb.WriteString(fmt.Sprintf("<li>%s: %s</li>", record.Id, record.Phase))
		}
		sb.WriteString("</ul>")
	}

	sb.WriteString("</body>")

	a := grafana.PostAnnotation{
		DashboardUID: l.dashboardUID,
		Time:         Ptr[time.Time](chaos.GetStartTime()),
		Text:         sb.String(),
	}
	_, _, err = l.client.PostAnnotation(a)
	if err != nil {
		l.logger.Warn().Msgf("could not annotate on Grafana: %s", err)
	}
}

func (l SingleLineGrafanaAnnotator) OnChaosPaused(chaos K8sChaos) {
}

func (l SingleLineGrafanaAnnotator) OnChaosEnded(chaos K8sChaos) {
	experiment, _ := chaos.GetExperimentStatus()
	duration, _ := chaos.GetChaosDuration()

	var sb strings.Builder
	sb.WriteString("<body>")
	sb.WriteString(fmt.Sprintf("<h4>%s Ended</h4>", chaos.GetChaosTypeStr()))
	sb.WriteString(fmt.Sprintf("<div>Name: %s</div>", chaos.Object.GetName()))
	if chaos.Description != "" {
		sb.WriteString(fmt.Sprintf("<div>Description: %s</div>", chaos.Description))
	}
	sb.WriteString(fmt.Sprintf("<div>Start Time: %s</div>", chaos.GetStartTime().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("<div>End Time: %s</div>", chaos.GetEndTime().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("<div>Duration: %s</div>", duration.String()))

	spec := chaos.GetChaosSpec()
	specBytes, err := json.MarshalIndent(spec, "", "  ")
	if err == nil && len(specBytes) > 0 {
		sb.WriteString("<br>")
		sb.WriteString("<h5>Spec:</h5>")
		sb.WriteString(string(specBytes))
		sb.WriteString("<br>")
	} else {
		l.logger.Warn().Msgf("could not get chaos spec: %s", err)
	}

	if len(experiment.Records) > 0 {
		sb.WriteString("<br>")
		sb.WriteString("<h5>Records:</h5>")
		sb.WriteString("<ul>")
		for _, record := range experiment.Records {
			sb.WriteString(fmt.Sprintf("<li>%s: %s</li>", record.Id, record.Phase))
		}
		sb.WriteString("</ul>")
	}

	sb.WriteString("</body>")

	a := grafana.PostAnnotation{
		DashboardUID: l.dashboardUID,
		Time:         Ptr[time.Time](chaos.GetEndTime()),
		Text:         sb.String(),
	}
	_, _, err = l.client.PostAnnotation(a)
	if err != nil {
		l.logger.Warn().Msgf("could not annotate on Grafana: %s", err)
	}
}

func (l SingleLineGrafanaAnnotator) OnChaosStatusUnknown(chaos K8sChaos) {
}

func (l SingleLineGrafanaAnnotator) OnScheduleCreated(s Schedule) {
	var sb strings.Builder
	sb.WriteString("<body>")
	sb.WriteString(fmt.Sprintf("<h4>%s Schedule Created</h4>", s.Object.Spec.Type))
	sb.WriteString(fmt.Sprintf("<div>Name: %s</div>", s.Object.ObjectMeta.Name))
	sb.WriteString(fmt.Sprintf("<div>Schedule: %s</div>", s.Object.Spec.Schedule))
	if s.Description != "" {
		sb.WriteString(fmt.Sprintf("<div>Description: %s</div>", s.Description))
	}
	sb.WriteString(fmt.Sprintf("<div>Start Time: %s</div>", s.startTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("<div>Duration: %s</div>", s.Duration.String()))

	spec := s.Object.Spec.ScheduleItem
	specBytes, err := json.MarshalIndent(spec, "", "  ")
	if err == nil && len(specBytes) > 0 {
		sb.WriteString("<br>")
		sb.WriteString("<h5>Schedule Spec:</h5>")
		sb.WriteString(string(specBytes))
		sb.WriteString("<br>")
	} else {
		l.logger.Warn().Msgf("could not get chaos spec: %s", err)
	}
	sb.WriteString("</body>")

	a := grafana.PostAnnotation{
		DashboardUID: l.dashboardUID,
		Time:         Ptr[time.Time](s.startTime),
		Text:         sb.String(),
	}
	_, _, err = l.client.PostAnnotation(a)
	if err != nil {
		l.logger.Warn().Msgf("could not annotate on Grafana: %s", err)
	}
}

func (l SingleLineGrafanaAnnotator) OnScheduleDeleted(s Schedule) {
	var sb strings.Builder
	sb.WriteString("<body>")
	sb.WriteString(fmt.Sprintf("<h4>%s Schedule Ended</h4>", s.Object.Spec.Type))
	sb.WriteString(fmt.Sprintf("<div>Name: %s</div>", s.Object.ObjectMeta.Name))
	sb.WriteString(fmt.Sprintf("<div>Schedule: %s</div>", s.Object.Spec.Schedule))
	if s.Description != "" {
		sb.WriteString(fmt.Sprintf("<div>Description: %s</div>", s.Description))
	}
	sb.WriteString(fmt.Sprintf("<div>Start Time: %s</div>", s.startTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("<div>End Time: %s</div>", s.endTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("<div>Duration: %s</div>", s.Duration.String()))

	spec := s.Object.Spec.ScheduleItem
	specBytes, err := json.MarshalIndent(spec, "", "  ")
	if err == nil && len(specBytes) > 0 {
		sb.WriteString("<br>")
		sb.WriteString("<h5>Schedule Spec:</h5>")
		sb.WriteString(string(specBytes))
		sb.WriteString("<br>")
	} else {
		l.logger.Warn().Msgf("could not get chaos spec: %s", err)
	}
	sb.WriteString("</body>")

	a := grafana.PostAnnotation{
		DashboardUID: l.dashboardUID,
		Time:         Ptr[time.Time](s.endTime),
		Text:         sb.String(),
	}
	_, _, err = l.client.PostAnnotation(a)
	if err != nil {
		l.logger.Warn().Msgf("could not annotate on Grafana: %s", err)
	}
}
