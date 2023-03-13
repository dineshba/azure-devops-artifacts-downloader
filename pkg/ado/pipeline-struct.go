package ado

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type AzurePipeline struct {
	Resources struct {
		Pipelines []PipelineResource `yaml:"pipelines"`
	} `yaml:"resources"`
}

type PipelineResource struct {
	Pipeline string `yaml:"pipeline"`
	Source   string `yaml:"source"`
	Branch   string `yaml:"branch"`
	Version  string `yaml:"version"`
}

func GetAzurePipelineConfig(fileName string) (*AzurePipeline, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fileName, err)
	}

	var config AzurePipeline
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %s: %w", fileName, err)
	}

	return &config, nil
}
