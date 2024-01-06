package havoc

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"math/rand"
	"time"
)

const (
	MonkeyModeSeq    = "seq"
	MonkeyModeRandom = "rand"

	ErrInvalidMode = "monkey mode is invalid, should be either \"seq\" or \"rand\""
)

type ExperimentAction struct {
	Name           string
	ExperimentType string
	ExperimentSpec string
	TimeStart      int64
	TimeEnd        int64
}

type ExperimentAnnotationBody struct {
	DashboardUID string   `json:"dashboardUID"`
	Time         int64    `json:"time"`
	TimeEnd      int64    `json:"timeEnd"`
	Tags         []string `json:"tags"`
	Text         string   `json:"text"`
}

type ChaosMonkey struct {
	client            *resty.Client
	cfg               *Config
	experimentActions []*ExperimentAction
}

func NewMonkey(cfg *Config) (*ChaosMonkey, error) {
	InitDefaultLogging()
	if cfg == nil {
		cfg = DefaultConfig()
	}
	c := resty.New()
	c.SetBaseURL(cfg.Havoc.Monkey.GrafanaURL)
	c.SetAuthScheme("Bearer")
	c.SetAuthToken(cfg.Havoc.Monkey.GrafanaToken)
	return &ChaosMonkey{
		client:            c,
		cfg:               cfg,
		experimentActions: make([]*ExperimentAction, 0),
	}, nil
}

// annotateAndSend sends annotation marker to Grafana dashboard
func (m *ChaosMonkey) annotateAndSend(a *ExperimentAction) error {
	start := a.TimeStart * 1e3
	end := a.TimeEnd * 1e3
	specBody := fmt.Sprintf("<pre>%s</pre>", a.ExperimentSpec)
	aa := &ExperimentAnnotationBody{
		DashboardUID: m.cfg.Havoc.Monkey.DashboardName,
		Time:         start,
		TimeEnd:      end,
		Tags:         []string{"havoc", a.ExperimentType},
		Text: fmt.Sprintf(
			"ChaosExperimentFile: %s\n%s",
			a.Name,
			specBody,
		),
	}
	_, err := m.client.R().
		SetBody(aa).
		Post(fmt.Sprintf("%s/api/annotations", m.cfg.Havoc.Monkey.GrafanaURL))
	if err != nil {
		return err
	}
	L.Info().
		Str("Name", a.Name).
		Int64("Start", a.TimeStart).
		Int64("End", a.TimeEnd).
		Msg("Annotated experiment")
	return nil
}

func (m *ChaosMonkey) ApplyAndAnnotate(exp *NamedExperiment) error {
	ea := &ExperimentAction{
		Name:           exp.Name,
		ExperimentType: exp.Type,
		ExperimentSpec: exp.Manifest,
		TimeStart:      time.Now().Unix(),
	}
	if err := ApplyChaosFile(m.cfg.Havoc.Monkey.Dir, exp.Type, exp.Name, true); err != nil {
		return err
	}
	ea.TimeEnd = time.Now().Unix()
	return m.annotateAndSend(ea)
}

func (m *ChaosMonkey) Run(ctx context.Context) error {
	if m.cfg.Havoc.Monkey.GrafanaURL == "" || m.cfg.Havoc.Monkey.GrafanaToken == "" || m.cfg.Havoc.Monkey.DashboardName == "" {
		return errors.New(ErrInvalidMonkeyCreds)
	}
	dur, err := time.ParseDuration(m.cfg.Havoc.Monkey.Duration)
	if err != nil {
		return err
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dur)
		defer cancel()
	}
	switch m.cfg.Havoc.Monkey.Mode {
	case MonkeyModeSeq:
		for _, expType := range RecommendedExperimentTypes {
			experiments, err := ReadExperimentsFromDir([]string{expType}, m.cfg.Havoc.Monkey.Dir)
			if err != nil {
				return err
			}
			for _, exp := range experiments {
				if err := m.ApplyAndAnnotate(exp); err != nil {
					return err
				}
				cdDuration, err := time.ParseDuration(m.cfg.Havoc.Monkey.Cooldown)
				if err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					L.Info().Msg("Monkey has finished by timeout")
					return nil
				default:
				}
				L.Info().
					Dur("Duration", cdDuration).
					Msg("Cooldown between experiments")
				time.Sleep(cdDuration)
				L.Info().Msg("Monkey has finished all scheduled experiments")
			}
		}
	case MonkeyModeRandom:
		allExperiments := make([]*NamedExperiment, 0)
		r := rand.New(rand.NewSource(time.Now().Unix()))
		for _, expType := range RecommendedExperimentTypes {
			experiments, err := ReadExperimentsFromDir([]string{expType}, m.cfg.Havoc.Monkey.Dir)
			if err != nil {
				return err
			}
			allExperiments = append(allExperiments, experiments...)
		}
		for {
			select {
			case <-ctx.Done():
				L.Info().Msg("Monkey has finished by timeout")
				return nil
			default:
				exp := pickExperiment(r, allExperiments)
				if err := m.ApplyAndAnnotate(exp); err != nil {
					return err
				}
				cdDuration, err := time.ParseDuration(m.cfg.Havoc.Monkey.Cooldown)
				if err != nil {
					return err
				}
				L.Info().
					Dur("Duration", cdDuration).
					Msg("Cooldown between experiments")
				time.Sleep(cdDuration)
			}
		}
	default:
		return errors.New(ErrInvalidMode)
	}
	return nil
}

func pickExperiment(r *rand.Rand, s []*NamedExperiment) *NamedExperiment {
	return s[r.Intn(len(s))]
}
