package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gdhagger/go-buildkite/buildkite"
	joonix "github.com/joonix/log"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v1"
)

var apiToken string
var org string
var configFile string
var logLevel string
var logFormat string
var client *buildkite.Client

type pipeline struct {
	Name             string                    `yaml:"name"`
	Repository       string                    `yaml:"repository"`
	Steps            []buildkite.Step          `yaml:"steps"`
	ProviderSettings *buildkite.GitHubSettings `yaml:"provider_settings"`
}

func (p pipeline) asCreatePipeline() *buildkite.CreatePipeline {
	r := buildkite.CreatePipeline{
		Name:             p.Name,
		Repository:       p.Repository,
		Steps:            make([]buildkite.Step, len(p.Steps)),
		ProviderSettings: p.ProviderSettings,
	}
	for i := range p.Steps {
		r.Steps[i] = p.Steps[i]
	}
	return &r
}

func (p pipeline) asUpdatedPipeline(r *buildkite.Pipeline) *buildkite.Pipeline {
	for i := range p.Steps {
		r.Steps[i] = &p.Steps[i]
	}
	r.Provider.Settings = p.ProviderSettings
	return r
}

type autoconf struct {
	Pipelines []*pipeline `yaml:"pipelines"`
}

func init() {
	flag.StringVar(&apiToken, "token", "", "Buildkite API token")
	flag.StringVar(&org, "org", "", "Buildkite organisation slug")
	flag.StringVar(&configFile, "config", ".buildkite/autoconf.yaml", "Configuration file")
	flag.StringVar(&logLevel, "log-level", log.DebugLevel.String(), "Logging level")
	flag.StringVar(&logFormat, "log-format", "fluentd", "Logging format")
	flag.Parse()

	switch logFormat {
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	case "fluentd":
		log.SetFormatter(&joonix.FluentdFormatter{})
	}

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.SetLevel(level)

	config, err := buildkite.NewTokenConfig(apiToken, true)
	if err != nil {
		log.Fatalf("client config failed: %s", err)
	}
	client = buildkite.NewClient(config.Client())
}

func main() {
	contextLog := log.WithFields(log.Fields{
		"filename": configFile,
	})
	contextLog.Debug("Opening config file")
	autoconfFile, _ := os.Open(configFile)
	defer autoconfFile.Close()
	autoconfBytes, _ := ioutil.ReadAll(autoconfFile)
	var autoconfData autoconf
	contextLog.Debug("Unmarshalling config YAML")
	_ = yaml.Unmarshal(autoconfBytes, &autoconfData)
	for _, p := range autoconfData.Pipelines {
		contextLog := contextLog.WithField("pipeline", p.Name)
		contextLog.Debug("Checking for existing pipeline")
		existingPipe, _, _ := client.Pipelines.Get(org, p.Name)
		if existingPipe != nil {
			up := p.asUpdatedPipeline(existingPipe)
			upJSON, _ := json.Marshal(up)
			contextLog.WithField("pipeline_data", string(upJSON)).Debug("Updating existing pipeline")
			_, err := client.Pipelines.Update(org, up)
			if err != nil {
				contextLog.Error(err)
				contextLog.Error(string(upJSON))
			}
		} else {
			cp := p.asCreatePipeline()
			cpJSON, _ := json.Marshal(cp)
			contextLog.WithField("pipeline_data", string(cpJSON)).Debug("Creating new pipeline")
			_, _, err := client.Pipelines.Create(org, cp)
			if err != nil {
				contextLog.Error(err)
				contextLog.Error(string(cpJSON))
			}
		}
	}
}
