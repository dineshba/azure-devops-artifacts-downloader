# Azure Devops Artifacts Downloader (ado-ad)

A tool to download all the aritifacts specified the `azure-pipelines.yaml` and keep it ready for testing things in local

- [Why do we need this tool](#why-do-we-need-this-tool)
- [Usage example](#usage-example)
  * [Use your existing azure-pipelines yaml](#use-your-existing-azure-pipelines-yaml)
  * [Setup authentication and azure-devops project configuration](#setup-authentication-and-azure-devops-project-configuration)
  * [Run ado-ad to download the artifacts and put in current directory](#run-ado-ad-to-download-the-artifacts-and-put-in-current-directory)
  * [Re-run of ado-ad and forcefully download new artifact](#re-run-of-ado-ad-and-forcefully-download-new-artifact)
- [TODO](#todo)

## Why do we need this tool
- When we want to setup a environment for debugging things in our local machine instead of testing in ado agents
- When we want to test our code with artifacts from multiple pipeline
- Because it is not part of [azure-pipeline-agent](https://github.com/microsoft/azure-pipelines-agent/issues/2479)

## Usage example

### Use your existing azure-pipelines yaml
```yaml
resources:
  pipelines:
  - pipeline: _pipeline1
    source: teamA\repo1\pipeline1
    branch: main                  # ado-ad will download latest successful build artifact of branch main and put in _pipeline1 directory

  - pipeline: _directory1/pipeline2
    source: teamA\repo1\pipeline2
    version: "20230215.8"         # ado-ad will download build's artifact from run id "20230215.8" and put in _directory1/pipeline2 directory

  - pipeline: _directory2
    source: teamA\repo2\pipeline1
    branch: my-new-feature-branch # ado-ad will download latest successful build artifact of branch my-new-feature-branch and put in _directory2 directory

  - pipeline: _directory3
    source: teamA\repo2\pipeline2 # ado-ad will download latest successful build artifact across all branch and put in _directory3 directory
```

> when both version and branch are specified, then version will have higher precedence than branch.

> when branch is not specified, it will download latest successful across all branch (same as ado pipeline agents)

> For testing with artifacts from your branch/PR/buildNumber, edit yaml file and run below steps

### Setup authentication and azure-devops project configuration
Setup 3 environment variables

> You should have PAT token which has access to list builds,builddefinitions and download artifacts

#### For Linux
```sh
export AZURE_DEVOPS_ORG_URL="https://dev.azure.com/your-org-url"
export AZURE_DEVOPS_PROJECT_NAME="your-project-name"
export AZURE_DEVOPS_EXT_PAT="your-personal-access-token-xxxxxxxxxxxx"
```

### For windows
```pwsh
$env:AZURE_DEVOPS_ORG_URL = "https://dev.azure.com/your-org-url"
$env:AZURE_DEVOPS_PROJECT_NAME = "your-project-name"
$env:AZURE_DEVOPS_EXT_PAT = "your-personal-access-token-xxxxxxxxxxxx"
```

### Run ado-ad to download the artifacts and put in current directory

#### For Linux
```sh
./ado-ad azure-pipeline.yaml
```
### For windows
```pwsh
./ado-ad.exe azure-pipeline.yaml
```

### Re-run of ado-ad and forcefully download new artifact
During the first run of `./ado-ad` run, it will create file called `.ado-ad.state` file, whose content will be like
```yaml
state:
  _pipeline1: 546638
  _directory1/pipeline2: 910629
  _directory2: 721132
  _directory3: 721134
```
During re-run, it will not download already downloaded artifact. It will only download the updated artifacts and missing artifact.

> It will move the current `.ado-ad.state` to `.ado-ad.state.bak`

#### Scenerio 1
Let say for first pipeline `_pipeline1`, there is a new build for main branch available. (in above example, `branch: main` is specified)
During next, it will update the artifact of `_pipeline1`
#### Scenerio 2
You want to re-download few artifacts. Then remove the corresponding line in the `.ado-ad.state` file. In next run, it will download the missing artifacts.

## TODO
- [x] Way to skip the download of artifact if it is already present
- [x] Way to have a downloaded versions state and on running next time, if there is a change, download only the changed artifacts
- [ ] Download artifacts across multiple projects (now, it is possible only within `AZURE_DEVOPS_PROJECT_NAME` project)
- [ ] Run artifacts in parallel using go routine
- [ ] New Flag to specify the destination directory
- [ ] New Flag to print the version of the binary