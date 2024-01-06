package havoc

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	DefaultCMDTimeout = "3m"

	ErrInvalidChaosType   = "invalid chaos type, valid types are: failure, latency, memory, cpu, group-failure, group-latency"
	ErrNoSelection        = "no selection, exiting"
	ErrInvalidNamespace   = "first argument must be a valid k8s namespace"
	ErrAutocompleteError  = "autocomplete file walk errored"
	ErrInvalidMonkeyCreds = "in order to run monkey you need to set GRAFANA_URL/GRAFANA_TOKEN/DASHBOARD_NAME vars"
)

func experimentCompleter(dir string, expType string) (func(d prompt.Document) []prompt.Suggest, error) {
	s := make([]prompt.Suggest, 0)
	err := filepath.Walk(
		fmt.Sprintf("%s/%s", dir, expType),
		func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			s = append(s, prompt.Suggest{
				Text:        info.Name(),
				Description: info.Name(),
			})
			return nil
		})
	if err != nil {
		return nil, err
	}
	return func(d prompt.Document) []prompt.Suggest {
		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}, nil
}

func RunCLI(args []string) error {
	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "havoc",
		Version:              "v0.0.1",
		Usage:                "Automatic chaos experiments CLI",
		UsageText:            `Utility to generate and apply chaos experiments for a namespace`,
		Before: func(cCtx *cli.Context) error {
			InitDefaultLogging()
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Aliases: []string{"c"}},
			&cli.StringFlag{Name: "dir", Aliases: []string{"d"}},
		},
		Commands: []*cli.Command{
			{
				Name:     "generate",
				HelpName: "generate",
				Aliases:  []string{"g"},
				Description: `generates chaos experiments:
havoc generate [namespace]
or use custom config
havoc -c havoc.toml generate [namespace]
you can also specify a directory where to put manifests
havoc -c havoc.toml -d custom_experiments [namespace]
`,
				Action: func(cliCtx *cli.Context) error {
					ns := cliCtx.Args().Get(0)
					if ns == "" {
						return errors.New(ErrInvalidNamespace)
					}
					dir := cliCtx.String("dir")
					if dir == "" {
						dir = DefaultExperimentsDir
					} else {
						if _, err := os.Stat(dir); err != nil {
							return err
						}
					}
					cfg, err := ReadConfig(cliCtx.String("config"))
					if err != nil {
						return err
					}
					return GenerateSpecs(ns, dir, cfg)
				},
			},
			{
				Name:     "apply",
				HelpName: "apply",
				Aliases:  []string{"a"},
				Description: `applies an experiment from a file:
examples:
havoc apply failure
havoc apply latency
havoc apply memory
havoc apply cpu
`,
				Action: func(cliCtx *cli.Context) error {
					chaosType := cliCtx.Args().Get(0)
					if !sliceContains(chaosType, RecommendedExperimentTypes) {
						return errors.New(ErrInvalidChaosType)
					}
					dir := cliCtx.String("dir")
					if dir == "" {
						dir = DefaultExperimentsDir
					} else {
						if _, err := os.Stat(dir); err != nil {
							return err
						}
					}
					c, err := experimentCompleter(dir, chaosType)
					if err != nil {
						return errors.Wrap(err, ErrAutocompleteError)
					}
					expName := prompt.Input(">> ", c)
					if expName == "" {
						return errors.New(ErrNoSelection)
					}
					return ApplyChaosFile(dir, chaosType, expName, true)
				},
			},
			{
				Name:     "run",
				HelpName: "run",
				Aliases:  []string{"r"},
				Description: `starts a chaos monkey
examples:
havoc run -c havoc.toml [namespace]
`,
				Action: func(cliCtx *cli.Context) error {
					ns := cliCtx.Args().Get(0)
					cfgPath := cliCtx.String("config")
					cfg, err := ReadConfig(cfgPath)
					if err != nil {
						return err
					}
					if cfg.Havoc.Monkey.Dir == "" {
						cfg.Havoc.Monkey.Dir = "havoc-monkey-temp-dir"
						err = GenerateSpecs(
							ns,
							cfg.Havoc.Monkey.Dir,
							cfg,
						)
						if err != nil {
							return err
						}
					}
					m, err := NewMonkey(cfg)
					if err != nil {
						return err
					}
					return m.Run(nil)
				},
			},
		},
	}
	return app.Run(args)
}
