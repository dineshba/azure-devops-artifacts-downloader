package downloader

import (
	"ado-ad/pkg/ado"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v6/build"
)

var regexForGettingNameOfPipelineFromSource, _ = regexp.Compile("[a-zA-Z.-]*$")

func NewDownloader(adoContext *ado.ADOContext) (*Downloader, error) {
	client, err := NewClient(adoContext)
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}
	return &Downloader{
		adoContext: adoContext,
		client:     client,
	}, nil
}

type Downloader struct {
	adoContext *ado.ADOContext
	client     *Client
}

type DownloadState struct {
	State map[string]int
}

func (d *Downloader) DownloadResources(pipelineConfig *ado.AzurePipeline, currentState *DownloadState) (*DownloadState, error) {
	downloadState := &DownloadState{State: map[string]int{}}
	for _, pipeline := range pipelineConfig.Resources.Pipelines {
		fmt.Printf("Processing: %s, %s\n", pipeline.Pipeline, pipeline.Source)
		buildId, err := d.downloadResourcesForPipeline(pipeline, currentState.State[pipeline.Pipeline])
		if err != nil {
			return nil, err
		}
		fmt.Printf("---- Downloaded for %s ----\n\n", pipeline.Source)
		downloadState.State[pipeline.Pipeline] = *buildId
	}
	return downloadState, nil
}

func (d *Downloader) downloadResourcesForPipeline(pipeline ado.PipelineResource, downloadedBuildNumber int) (*int, error) {
	name := regexForGettingNameOfPipelineFromSource.FindString(pipeline.Source)
	path := fmt.Sprintf("\\%s", strings.TrimSuffix(pipeline.Source, "\\"+name))

	defs, err := d.client.GetBuildDefinitions(context.TODO())
	if err != nil {
		return nil, err
	}
	// for _, def := range defs {
	// 	fmt.Printf("%s, %d, %s\n", *def.Name, *def.Id, *def.Path)
	// }
	// fmt.Printf("Found %d definitions\n", len(defs))
	definitionId := getDefinitionIdFor(path, name, defs)

	if definitionId == 0 {
		return nil, fmt.Errorf("definition not found for pipeline %s/%s", path, name)
	}

	buildId, artifacts, err := d.client.GetArtifactsForDefinition(context.TODO(), definitionId, pipeline, downloadedBuildNumber)
	if err != nil {
		return nil, err
	}

	if *buildId == downloadedBuildNumber {
		fmt.Printf("Skipping artifact download and unzipping for pipeline %s\n\n", pipeline.Pipeline)
		return buildId, nil
	}

	fmt.Printf("Found %d artifacts\n", len(artifacts))

	for _, artifact := range artifacts {
		err = d.client.DownloadAritfact(context.TODO(), *artifact.Resource.DownloadUrl, *artifact.Name, pipeline.Pipeline)
		if err != nil {
			return nil, fmt.Errorf("failed to download artifact %s: %w", *artifact.Name, err)
		}
	}
	err = UnzipArtifact(pipeline.Pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to unzip artifact for pipeline %s: %w", pipeline.Pipeline, err)
	}
	err = CleanupZipFiles(pipeline.Pipeline)
	if err != nil {
		return nil, fmt.Errorf("error cleaning up zip files: %w", err)
	}
	return buildId, nil
}

func getDefinitionIdFor(path, name string, defs []build.BuildDefinitionReference) int {
	for _, def := range defs {
		if *def.Path == path && *def.Name == name {
			// fmt.Printf("%+v\n\n", def)
			fmt.Printf("Found definition for pipeline %s, %d, %s\n", *def.Name, *def.Id, *def.Path)
			return *def.Id
		}
	}
	return 0
}
