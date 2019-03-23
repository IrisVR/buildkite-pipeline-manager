package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/gdhagger/go-buildkite/buildkite"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v1"
)

var apiToken string
var org string
var configFile string
var client *buildkite.Client

type pipeline struct {
	Name             string                    `yaml:"name"`
	Repository       string                    `yaml:"repository"`
	Steps            []*buildkite.Step         `yaml:"steps"`
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
		r.Steps[i] = *p.Steps[i]
	}
	return &r
}

func (p pipeline) asUpdatedPipeline(r *buildkite.Pipeline) *buildkite.Pipeline {
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
	flag.Parse()

	config, err := buildkite.NewTokenConfig(apiToken, true)
	if err != nil {
		log.Fatalf("client config failed: %s", err)
	}
	client = buildkite.NewClient(config.Client())
}

func main() {
	autoconfFile, _ := os.Open(configFile)
	defer autoconfFile.Close()
	autoconfBytes, _ := ioutil.ReadAll(autoconfFile)
	var autoconfData autoconf
	_ = yaml.Unmarshal(autoconfBytes, &autoconfData)
	for _, p := range autoconfData.Pipelines {
		existingPipe, _, _ := client.Pipelines.Get(org, p.Name)
		if existingPipe != nil {
			up := p.asUpdatedPipeline(existingPipe)
			_, err := client.Pipelines.Update(org, up)
			if err != nil {
				log.Error(err)
				upJSON, _ := json.Marshal(up)
				log.Error(string(upJSON))
			}
		} else {
			cp := p.asCreatePipeline()
			_, _, err := client.Pipelines.Create(org, cp)
			if err != nil {
				log.Error(err)
				cpJSON, _ := json.Marshal(cp)
				log.Error(string(cpJSON))
			}
		}
	}
}
