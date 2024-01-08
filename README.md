## Havoc CLI

Havoc is a tool that introspects your k8s namespace and generates a chaos suite for you

You can use havoc as a CLI to quickly test hypothesis or run it in "monkey" mode with your load tests

### Goals

- Make chaos testing easy by generating most of the things automatically
- Easy integration with Grafana to understand how chaos affects your services
- Be easy to use both programmatically and as a CLI

### How it works
Havoc generates groups of experiments based just on your pods found in namespace

Single pod experiments:

- PodFailure
- NetworkChaos (Pod latency)
- Stress (Memory)
- Stress (CPU)
- External service failure (Network partition)

Group experiments:

- Pods failure
- Pods latencies

You can generate default chaos suite by [configuring](havoc.toml) havoc then set `dir` param and add your custom experiments, then run monkey to test your services

### Setup

We are using [nix](https://nixos.org/)

Enter the shell

```
nix develop
```

### Install as a binary

Please use GitHub releases of this repo

### Manual usage

Generate default experiments for your namespace

```
havoc generate [namespace]
or with a custom config
havoc -c havoc.toml generate [namespace]
```

This will create `experiments` dir, then you can choose from recommended experiments

```
havoc apply {failure, latency, memory, cpu, external, group-failure, group-latency}
```

### Monkey mode
You can run havoc as an automated sequential or randomized suite
```
havoc -c havoc.toml run [namespace]
```
See `[havoc.monkey]` config [here](havoc.toml)

### Programmatic usage

See how you can use recommended experiments from code in [examples](examples)

### Environment variables
You need to set the last three in order to run monkey mode
```
HAVOC_LOG_LEVEL={warn,info,debug,trace}
GRAFANA_URL="..."
GRAFANA_TOKEN="..."
DASHBOARD_NAME="..."
```
