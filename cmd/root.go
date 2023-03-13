package cmd

import (
	"ado-ad/pkg/ado"
	downloaderPkg "ado-ad/pkg/downloader"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var rootCmd = &cobra.Command{
	Use:          "ado-ad",
	SilenceUsage: true,
	Short:        "azure-devops artifacts-downloader",
	Long:         "A tool to download all the aritifacts specified the `azure-pipelines.yaml` and keep it ready for testing things in local",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Printf("Provide the azure-pipeline yaml file\n")
			os.Exit(1)
		}
		err := downloadArtifacts(args[0])
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		}
	},
}

var adoAdStateFile = ".ado-ad.state"

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func downloadArtifacts(fileName string) error {
	pipelineConfig, err := ado.GetAzurePipelineConfig(fileName)
	if err != nil {
		return err
	}
	adoContext, err := getADOContext()
	if err != nil {
		return fmt.Errorf("error getting ado context: %s", err.Error())
	}
	downloader, err := downloaderPkg.NewDownloader(adoContext)
	if err != nil {
		return fmt.Errorf("error creating downloader: %s", err.Error())
	}
	currentState, err := getAdoAdState()
	if err != nil {
		return fmt.Errorf("error getting ado-ad state: %s", err.Error())
	}
	nextState, err := downloader.DownloadResources(pipelineConfig, currentState)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(currentState, nextState) {
		fmt.Printf("\n ------- No new builds found -------\n")
		return nil
	} else {
		fmt.Printf("Created artifacts successfully\n")
	}
	return writeAdoAdState(nextState)
}

func getAdoAdState() (*downloaderPkg.DownloadState, error) {
	if _, err := os.Stat(adoAdStateFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error getting the file info for file %s: %w", adoAdStateFile, err)
		} else {
			return &downloaderPkg.DownloadState{}, nil
		}
	}
	data, err := os.ReadFile(adoAdStateFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", adoAdStateFile, err)
	}
	downloadState := &downloaderPkg.DownloadState{}
	err = yaml.Unmarshal(data, &downloadState)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling yaml: %w", err)
	}
	return downloadState, nil
}

func writeAdoAdState(newState *downloaderPkg.DownloadState) error {
	if _, err := os.Stat("/path/to/whatever"); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error getting the file info for file %s: %w", adoAdStateFile, err)
		}
		os.Rename(adoAdStateFile, adoAdStateFile+".bak")
	}

	v, err := yaml.Marshal(newState)
	if err != nil {
		return fmt.Errorf("error marshalling yaml: %w", err)
	}

	f, err := os.Create(adoAdStateFile)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", adoAdStateFile, err)
	}

	_, err = f.Write(v)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", adoAdStateFile, err)
	}
	return nil
}

func getADOContext() (*ado.ADOContext, error) {
	personalAccessToken := os.ExpandEnv("$AZURE_DEVOPS_EXT_PAT")
	projectName := os.ExpandEnv("$AZURE_DEVOPS_PROJECT_NAME")
	organizationUrl := os.ExpandEnv("$AZURE_DEVOPS_ORG_URL")

	if len(organizationUrl) == 0 {
		return nil, fmt.Errorf("missing AZURE_DEVOPS_ORG_URL env variable")
	}
	if len(projectName) == 0 {
		return nil, fmt.Errorf("missing AZURE_DEVOPS_PROJECT_NAME env variable")
	}
	if len(personalAccessToken) == 0 {
		return nil, fmt.Errorf("missing AZURE_DEVOPS_EXT_PAT env variable")
	}

	context := ado.NewADOContext(projectName, organizationUrl, personalAccessToken)
	return &context, nil
}
