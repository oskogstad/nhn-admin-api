package main

import (
	"context"
	"os"
	"time"

	console "github.com/fatih/color"
	memfs "github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	git_http "github.com/go-git/go-git/v5/plumbing/transport/http"
	memory "github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

var gitRepoURL, gitUrlFound = os.LookupEnv("NHN_K8S_GIT_REPO_URL")
var gitRepoName, gitRepoNameFound = os.LookupEnv("NHN_K8S_GIT_REPO_NAME")
var gitUserName, gitUserNameFound = os.LookupEnv("NHN_GIT_USERNAME")
var gitAccessToken, gitTokenFound = os.LookupEnv("NHN_GIT_ACCESS_TOKEN")

var fileSystem = memfs.New()
var gitRepo *git.Repository
var auth *git_http.BasicAuth

func GitSetup() bool {
	if !EnsureGitEnvFound() {
		return false
	}

	CreateAuth()
	CloneRepo()

	return true
}

// EnsureGitEnvFound checks that git environment variables are found, Access token and repo URL
func EnsureGitEnvFound() bool {
	if !gitUrlFound || !gitTokenFound || !gitRepoNameFound || !gitUserNameFound {
		console.Red("Failed to init ENV vars for git.")
		console.Red("Git repo name was: " + gitRepoName)
		console.Red("Git username was: " + gitUserName)
		console.Red("Git repo URL was: " + gitRepoURL)
		console.Red("Access token was: " + gitAccessToken)
		return false
	}
	return true
}

func CreateAuth() {
	auth = &git_http.BasicAuth{
		Username: "any string except empty",
		Password: gitAccessToken,
	}
}

func CloneRepo() bool {
	console.Green("git clone " + gitRepoURL)
	var err error
	gitRepo, err = git.Clone(memory.NewStorage(), fileSystem, &git.CloneOptions{
		URL:      gitRepoURL,
		Progress: os.Stdout,
		Auth:     auth,
	})

	CheckIfError(err)
	return true
}

func CreateFile(path string, fileContent []byte) string {
	filePath := path
	file, err := fileSystem.Create(filePath)
	CheckIfError(err)

	file.Write(fileContent)
	file.Close()

	return filePath
}

// CreateNewServiceConfig creates a new pull request with config files for a new microservice
func CreateNewServiceConfig(service Service) {
	workTree, err := gitRepo.Worktree()
	CheckIfError(err)

	mainBranch := plumbing.ReferenceName("refs/heads/main")

	console.Green("git checkout main")
	err = workTree.Checkout(&git.CheckoutOptions{
		Create: false,
		Branch: mainBranch,
	})
	CheckIfError(err)

	console.Green("git pull")
	workTree.Pull(&git.PullOptions{
		Auth: auth,
	})

	branchRef := plumbing.ReferenceName("refs/heads/" + service.Name)

	console.Green("git checkout -b " + service.Name)
	err = workTree.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: branchRef,
	})

	CheckIfError(err)

	helmFileContent := CreateHelmValuesFile(service)
	helmFilePath := CreateFile("app/"+service.GatewayEndpoint+"/"+service.Name+"-helm-values.yaml", helmFileContent)
	workTree.Add(helmFilePath)

	argoFileContent := CreateArgoAppFile(service)
	argoFilePath := CreateFile("argocd/apps/"+service.GatewayEndpoint+"/values.yaml", argoFileContent)
	workTree.Add(argoFilePath)

	workTree.Commit("Add service config for "+service.Name, &git.CommitOptions{Author: &object.Signature{
		Name:  "Adminservice",
		Email: "adminservice@api.nhn.no",
		When:  time.Now(),
	}})

	err = gitRepo.Push(&git.PushOptions{Auth: auth, Force: true})
	CheckIfError(err)

	// Create PR
	context := context.Background()
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gitAccessToken},
	)
	oauthClient := oauth2.NewClient(context, tokenSource)

	githubClient := github.NewClient(oauthClient)
	newPR := &github.NewPullRequest{
		Title:               github.String("New kubernetes deployment: " + service.Name),
		Head:                github.String(service.Name),
		Base:                github.String("main"),
		Body:                github.String("Automated PR for application " + service.Name),
		MaintainerCanModify: github.Bool(true),
	}

	_, _, err = githubClient.PullRequests.Create(context, gitUserName, gitRepoName, newPR)
	CheckIfError(err)
	console.Yellow("PR created")
}
