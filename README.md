## Havoc

Havoc is a tool that introspects your k8s namespace and generates a `ChaosMesh` CRDs suite for you

You can use havoc as a CLI to quickly test hypothesis or run it in "monkey" mode with your load tests and have Grafana annotations

### Goals

- Make chaos testing easy by generating most of the things automatically just by looking at your namespace
- Easy integration with Grafana to understand how chaos affects your services
- Be easy to use both programmatically and as a CLI

### How it works
Havoc generates groups of experiments based just on your pods and labels found in namespace

Single pod experiments:

- PodFailure
- NetworkChaos (Pod latency)
- Stress (Memory)
- Stress (CPU)
- External service failure (Network partition)

Group experiments:

- Group failure
- Group latency
- Group CPU
- Group memory

You can generate default chaos suite by [configuring](havoc.toml) havoc then set `dir` param and add your custom experiments, then run monkey to test your services

### Install

Please use GitHub releases of this repo
Download latest [release](https://github.com/smartcontractkit/havoc/releases)

You need `kubectl` to available on your machine

If you wish Grafana integration, please set env variables (optional)
```
HAVOC_LOG_LEVEL={warn,info,debug,trace}
GRAFANA_URL="..."
GRAFANA_TOKEN="..."
DASHBOARD_NAME="..."
```

### Manual usage

Generate default experiments for your namespace

```
havoc -c havoc.toml generate [namespace]
```

Check this [section](havoc.toml) for `ignore_pods` and `ignore_group_labels`, default settings should be reasonable, however, you can tweak them

This will create `havoc-experiments` dir, then you can choose from recommended experiments

```
havoc -c havoc.toml apply
```

### Monkey mode
You can run havoc as an automated sequential or randomized suite
```
havoc -c havoc.toml run [namespace]
```
See `[havoc.monkey]` config [here](havoc.toml)

### Programmatic usage

See how you can use recommended experiments from code in [examples](examples)

### Custom experiments

Havoc is just a generator and a module that reads your `dir = $mydir` from config

If you wish to add custom experiments written by hand create your custom directory and add experiments

Experiments will be executed in lexicographic order, however, for custom experiments there are 2 simple rules:
- directory names must be in `["external", "failure", "latency", "cpu", "memory", "group-failure", "group-latency"]`
- `metadata.name` should be equal to your experiment filename

When you are using `run` monkey command, if directory is not empty havoc won't automatically generate experiments, so you can extend generated experiments with your custom modifications

### Developing

We are using [nix](https://nixos.org/)

Enter the shell

```
nix develop
```

### Why not to use ChaosMesh UI/API instead of CRDs?

`ChaosMesh` UI/API is great, but it has some downsides:
- No OpenAPI spec, hard to integrate
- No dynamic generation for a namespace, you need to rely on labels that might change
- Writing chaos experiments is tedious, in most cases you just copy-paste a lot, or you can forget something
- Workflows validation is broken
- Can't mix chaos experiments and API calls
- No straightforward integration with load testing tools, it's easy to run an experiment, but it's hard to validate it right away without additional code
- Can't check chaos experiments statuses through API and fail the test, need to use k8s
- Experiments created from YAML and UI and not always compatible

### Disclaimer
This software represents an educational example of how we break a Chainlink system, product, or service and is provided to demonstrate how to interact with Chainlink’s systems.
This code is provided “AS IS” and “AS AVAILABLE” without warranties of any kind, it has not been audited, and it may be missing key checks or error handling to make the usage of the system, product or service more clear.
Do not use the code in this example in a production environment without completing your own audits and application of best practices.
Neither Chainlink Labs, the Chainlink Foundation, nor Chainlink node operators are responsible for unintended outputs that are generated due to errors in code.
