package main

import (
	"context"
	"fmt"
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

	branchRef := plumbing.ReferenceName("refs/heads/" + service.Name)

	console.Green("git checkout -b " + service.Name)
	err = workTree.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: branchRef,
	})

	CheckIfError(err)

	filePath := "api/" + service.Name + "/" + "config-" + service.Name + ".yaml"
	newFile, err := fileSystem.Create(filePath)
	CheckIfError(err)

	newFile.Write([]byte("Config for service " + service.Name))
	newFile.Write([]byte("\nService name: " + service.Name))
	newFile.Write([]byte("\nAPI endpoint: " + service.APIBaseEndpoint))
	newFile.Write([]byte("\nOCI Image: " + service.OciImage))
	newFile.Write([]byte("\nPort: " + fmt.Sprint(service.Port)))
	newFile.Close()

	workTree.Add(filePath)

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
		Title:               github.String("New k8s service: " + service.Name),
		Head:                github.String(service.Name),
		Base:                github.String("main"),
		Body:                github.String("Automated PR for new microservice " + service.Name),
		MaintainerCanModify: github.Bool(true),
	}

	_, _, err = githubClient.PullRequests.Create(context, gitUserName, gitRepoName, newPR)
	CheckIfError(err)
}
