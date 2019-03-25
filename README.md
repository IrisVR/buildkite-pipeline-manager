# Buildkite Pipeline Manager

Configures [Buildkite](https://buildkite.com) pipelines based on a YAML configuation file, e.g.:
```
pipelines:
- name: buildkite-pipeline-manager-build
  repository: git@github.com:irisvr/buildkite-pipeline-manager
  steps: 
  - type: "script"
    name: ":pipeline:"
    command: "buildkite-agent pipeline upload .buildkite/build-pipeline.yaml"
    agent_query_rules:
    - queue=linux
- name: buildkite-pipeline-manager-deploy
  repository: git@github.com:irisvr/buildkite-pipeline-manager
  provider_settings:
    trigger_mode: "none"
  steps: 
  - type: "script"
    name: ":pipeline:"
    command: "buildkite-agent pipeline upload .buildkite/build-pipeline.yaml"
    agent_query_rules:
    - queue=linux
```

In this example, two pipelines are created.  
1. `buildkite-pipeline-manager-build` will load the `.buildkite/build-pipeline.yaml` on an agent where `queue=linux`, and will trigger automatically if the GitHub webhook is fired (Buildkite default)
2. `buildkite-pipeline-manager-deploy` is almost identical but uses the `.buildkite/build-pipeline.yaml` pipeline definition and has webhook triggering disabled.

## Usage
```
--org string          Buildkite organisation slug (required)
--token string        Buildkite API token (required)
--config string       Configuration file (default ".buildkite/autoconf.yaml")
--log-format string   Logging format, one of text, json or fluentd (default "fluentd")
--log-level string    Logging level (default "debug")
```

## Configuration Format

The configuration for each pipeline as as per the [Buildkite API specification](https://buildkite.com/docs/apis/rest-api/pipelines#create-a-pipeline).