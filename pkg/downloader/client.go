package downloader

import (
	"ado-ad/pkg/ado"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v6"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6/build"
)

type Client struct {
	projectName string
	buildClient build.Client
	token       string
}

func NewClient(adoContext *ado.ADOContext) (*Client, error) {
	connection := azuredevops.NewPatConnection(adoContext.OrganizationUrl, adoContext.GetAccessToken())

	newBuildClient, err := build.NewClient(context.TODO(), connection)
	if err != nil {
		return nil, fmt.Errorf("error creating build client: %s", err.Error())
	}

	return &Client{
		projectName: adoContext.ProjectName,
		buildClient: newBuildClient,
		token:       adoContext.GetAccessToken(),
	}, nil
}

func (c *Client) GetBuildDefinitions(ctx context.Context) ([]build.BuildDefinitionReference, error) {
	definitions, err := c.buildClient.GetDefinitions(ctx, build.GetDefinitionsArgs{
		Project: &c.projectName,
	})
	if err != nil {
		return nil, err
	}

	return definitions.Value, nil
}

func (c *Client) GetArtifactsForDefinition(ctx context.Context, definitionId int, pipeline ado.PipelineResource, downloadedBuildNumber int) (*int, []build.BuildArtifact, error) {
	definitions := []int{definitionId}
	buildArgs := build.GetBuildsArgs{
		Project:      &c.projectName,
		Definitions:  &definitions,
		StatusFilter: &build.BuildStatusValues.Completed,
		// ResultFilter: &build.BuildResultValues.PartiallySucceeded,
	}
	logMessage := ""
	if pipeline.Version != "" {
		buildArgs.BuildNumber = &pipeline.Version
		logMessage = fmt.Sprintf("version %s", pipeline.Version)
	} else {
		if pipeline.Branch != "" {
			branchName := addPrefixRefsHeadIfNotPresent(pipeline.Branch)
			buildArgs.BranchName = &branchName
			logMessage = fmt.Sprintf("branch %s", branchName)
		} else {
			logMessage = "latest build on all branches"
		}
	}
	builds, err := c.buildClient.GetBuilds(ctx, buildArgs)

	if err != nil {
		return nil, nil, fmt.Errorf("error getting builds: %s", err.Error())
	}

	if len(builds.Value) == 0 {
		if pipeline.Version != "" {
			return nil, nil, fmt.Errorf("no build found for given version %s", pipeline.Version)
		}
		if pipeline.Branch != "" {
			return nil, nil, fmt.Errorf("no build found for branch %s", pipeline.Branch)
		}
		return nil, nil, fmt.Errorf("no successful build found for definition %s", pipeline.Pipeline)
	}
	matchingBuild := builds.Value[0]

	if downloadedBuildNumber == *matchingBuild.Id {
		fmt.Printf("Build %d already downloaded for pipeline %s\n", *matchingBuild.Id, pipeline.Pipeline)
		return matchingBuild.Id, nil, nil
	}
	if downloadedBuildNumber != 0 {
		fmt.Printf("Previously downloaded build %d for pipeline %s is outdated, so downloading build %d\n", downloadedBuildNumber, pipeline.Pipeline, *matchingBuild.Id)
	} else {
		fmt.Printf("Found build %d for definition %d with %s\n", *matchingBuild.Id, definitionId, logMessage)
	}

	artifacts, err := c.buildClient.GetArtifacts(ctx, build.GetArtifactsArgs{
		Project: &c.projectName,
		BuildId: matchingBuild.Id,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("error getting artifacts: %s", err.Error())
	}

	return matchingBuild.Id, *artifacts, nil
}

func addPrefixRefsHeadIfNotPresent(branchName string) string {
	return fmt.Sprintf("refs/heads/%s", strings.TrimPrefix(branchName, "refs/heads/"))
}

func (c *Client) DownloadAritfact(ctx context.Context, downloadUrl string, name, folderName string) error {
	client := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	request, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		return fmt.Errorf("error creating download artifact request: %s", err.Error())
	}

	request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+c.token)))
	fmt.Printf("Downloading artifact %s\n", name)
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("error downloading artifact: %s", err.Error())
	}
	defer resp.Body.Close()

	err = os.MkdirAll(filepath.Join(".", folderName), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating folder %s: %s", folderName, err.Error())
	}
	fileName := filepath.Join(".", folderName, fmt.Sprintf("%s.zip", name))
	out, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("error creating output zip file %s: %s", fileName, err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("error writing content to zip file %s: %s", fileName, err)
	}
	fmt.Printf("Downloaded artifact %s\n", name)
	return nil
}
