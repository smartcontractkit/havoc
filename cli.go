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

	ErrNoSelection       = "no selection, exiting"
	ErrInvalidNamespace  = "first argument must be a valid k8s namespace"
	ErrAutocompleteError = "autocomplete file walk errored"
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

func experimentTypeCompleter(dir string) (func(d prompt.Document) []prompt.Suggest, error) {
	s := make([]prompt.Suggest, 0)
	err := filepath.Walk(
		dir,
		func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				s = append(s, prompt.Suggest{
					Text:        info.Name(),
					Description: info.Name(),
				})
			}
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
					cfg, err := ReadConfig(cliCtx.String("config"))
					if err != nil {
						return err
					}
					m, err := NewController(cfg)
					if err != nil {
						return err
					}
					return m.GenerateSpecs(ns)
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
					cfg, err := ReadConfig(cliCtx.String("config"))
					if err != nil {
						return err
					}
					m, err := NewController(cfg)
					if err != nil {
						return err
					}
					cc, err := experimentTypeCompleter(m.cfg.Havoc.Dir)
					if err != nil {
						return err
					}
					expType := prompt.Input("Choose experiment type >> ", cc)
					if expType == "" {
						return errors.New(ErrNoSelection)
					}
					c, err := experimentCompleter(m.cfg.Havoc.Dir, expType)
					if err != nil {
						return errors.Wrap(err, ErrAutocompleteError)
					}
					expName := prompt.Input("Choose experiment name >> ", c)
					if expName == "" {
						return errors.New(ErrNoSelection)
					}

					data, err := os.ReadFile(fmt.Sprintf("%s/%s/%s", cfg.Havoc.Dir, expType, expName))
					if err != nil {
						return err
					}
					nexp := &NamedExperiment{
						Name:     expName,
						Type:     expType,
						Manifest: string(data),
					}
					return m.ApplyAndAnnotate(nexp)
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
					m, err := NewController(cfg)
					if err != nil {
						return err
					}
					if cfg.Havoc.Dir == DefaultExperimentsDir {
						cfg.Havoc.Dir = "havoc-monkey-temp-dir"
						err = m.GenerateSpecs(ns)
						if err != nil {
							return err
						}
					}
					return m.Run()
				},
			},
		},
	}
	return app.Run(args)
}
