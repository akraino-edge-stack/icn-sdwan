// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	gogithub "github.com/google/go-github/v41/github"
	k8spluginv1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GitClient struct {
	gitProviderClient gitprovider.Client
	gogithubClient    *gogithub.Client
}

type GithubAccessClient struct {
	cl           GitClient
	gitUser      string
	gitRepo      string
	cluster      string
	githubDomain string
}

var GitHubClient GithubAccessClient

/*
	Function to create gitClient
	params : userName, github token
	return : github client, error
*/
func CreateClient(userName, githubToken string) (GitClient, error) {

	var client GitClient
	var err error

	client.gitProviderClient, err = github.NewClient(gitprovider.WithOAuth2Token(githubToken), gitprovider.WithDestructiveAPICalls(true))
	if err != nil {
		return GitClient{}, err
	}

	tp := gogithub.BasicAuthTransport{
		Username: userName,
		Password: githubToken,
	}
	client.gogithubClient = gogithub.NewClient(tp.Client())

	return client, nil

}

func SetupGitHubClient() error {
	var err error
	GitHubClient, err = NewGitHubClient()
	return err
}

func NewGitHubClient() (GithubAccessClient, error) {

	githubDomain := "github.com"
	gitUser := os.Getenv("GIT_USERNAME")
	gitToken := os.Getenv("GIT_TOKEN")
	gitRepo := os.Getenv("GIT_REPO")
	clusterName := os.Getenv("GIT_CLUSTERNAME")

	// If any value is not provided then can't store in Git location
	if len(gitRepo) <= 0 || len(gitToken) <= 0 || len(gitUser) <= 0 || len(clusterName) <= 0 {
		log.Info("Github information not found:: Skipping Github storage", log.Fields{})
		return GithubAccessClient{}, nil
	}
	log.Info("GitHub Info found", log.Fields{"gitRepo::": gitRepo, "cluster::": clusterName})

	cl, err := CreateClient(gitUser, gitToken)
	if err != nil {
		return GithubAccessClient{}, err
	}
	return GithubAccessClient{
		cl:           cl,
		gitUser:      gitUser,
		gitRepo:      gitRepo,
		githubDomain: githubDomain,
		cluster:      clusterName,
	}, nil
}

func CommitCR(c client.Client, cr *k8spluginv1alpha1.ResourceBundleState, org *k8spluginv1alpha1.ResourceBundleStateStatus) error {

	// Compare status and update if status changed
	resBytesCr, err := json.Marshal(cr.Status)
	if err != nil {
		log.Info("json Marshal error for resource::", log.Fields{"cr": cr, "err": err})
		return err
	}
	resBytesOrg, err := json.Marshal(org)
	if err != nil {
		log.Info("json Marshal error for resource::", log.Fields{"cr": cr, "err": err})
		return err
	}
	// If the status is not changed no need to update CR
	if bytes.Compare(resBytesCr, resBytesOrg) == 0 {
		return nil
	}
	err = c.Status().Update(context.TODO(), cr)
	if err != nil {
		if k8serrors.IsConflict(err) {
			return err
		} else {
			log.Info("CR Update Error::", log.Fields{"err": err})
			return err
		}
	}
	resBytes, err := json.Marshal(cr)
	if err != nil {
		log.Info("json Marshal error for resource::", log.Fields{"cr": cr, "err": err})
		return err
	}
	// Check if GIT Info is provided if so store the information in the Git Repo also
	err = GitHubClient.CommitCRToGitHub(resBytes, cr.GetLabels())
	if err != nil {
		log.Info("Error commiting status to Github", log.Fields{"err": err})
	}
	return nil
}

func (c *GithubAccessClient) CommitCRToGitHub(resBytes []byte, l map[string]string) error {

	// Get cid and app id
	v, ok := l["emco/deployment-id"]
	if !ok {
		return fmt.Errorf("Unexpected error:: Inconsistent labels %v", l)
	}
	result := strings.SplitN(v, "-", 2)
	if len(result) != 2 {
		return fmt.Errorf("Unexpected error:: Inconsistent labels %v", l)
	}
	app := result[1]
	cid := result[0]
	path := "clusters/" + c.cluster + "/status/" + cid + "/app/" + app + "/" + v

	s := string(resBytes)
	var files []gitprovider.CommitFile
	files = append(files, gitprovider.CommitFile{
		Path:    &path,
		Content: &s,
	})
	commitMessage := "Adding Status for " + path

	appName := c.cluster + "-" + cid + "-" + app
	// commitfiles
	err := c.CommitFiles(context.Background(), "main", commitMessage, appName, files)

	if err != nil {
		return err
	}
	return nil
}

/*
	Internal function to create a repo refercnce
	params : user name, repo name
	return : repo reference
*/
func (c *GithubAccessClient) getRepoRef(userName string, repoName string) gitprovider.UserRepositoryRef {
	// Create the user reference
	userRef := gitprovider.UserRef{
		Domain:    c.githubDomain,
		UserLogin: userName,
	}

	// Create the repo reference
	userRepoRef := gitprovider.UserRepositoryRef{
		UserRef:        userRef,
		RepositoryName: repoName,
	}

	return userRepoRef
}

var mutex = sync.Mutex{}

/*
	Function to commit multiple files to the github repo
	params : context, Branch Name, Commit Message, appName, files ([]gitprovider.CommitFile)
	return : nil/error
*/
func (c *GithubAccessClient) CommitFiles(ctx context.Context, branch, commitMessage, appName string, files []gitprovider.CommitFile) error {

	mergeBranch := appName
	// Only one process to commit to Github location to avoid conflicts
	mutex.Lock()
	defer mutex.Unlock()

	// commit the files to this new branch
	// create repo reference
	log.Info("Creating Repo Reference. ", log.Fields{})
	userRepoRef := c.getRepoRef(c.gitUser, c.gitRepo)
	log.Info("UserRepoRef:", log.Fields{"UserRepoRef": userRepoRef})

	log.Info("Obtaining user repo. ", log.Fields{})
	userRepo, err := c.cl.gitProviderClient.UserRepositories().Get(ctx, userRepoRef)
	if err != nil {
		return err
	}
	log.Info("UserRepo:", log.Fields{"UserRepo": userRepo})

	log.Info("Commiting Files:", log.Fields{"files": files})
	//Commit file to this repo
	resp, err := userRepo.Commits().Create(ctx, mergeBranch, commitMessage, files)
	if err != nil {
		if !strings.Contains(err.Error(), "404 Not Found") {
			log.Error("Error in commiting the files", log.Fields{"err": err, "mergeBranch": mergeBranch, "commitMessage": commitMessage, "files": files})
		}
		return err
	}
	log.Info("CommitResponse for userRepo:", log.Fields{"resp": resp})
	return nil
}

/*
	Function to obtaion the SHA of latest commit
	params : context, git client, User Name, Repo Name, Branch, Path
	return : LatestCommit string, error
*/
func GetLatestCommitSHA(ctx context.Context, c GitClient, userName, repoName, branch, path string) (string, error) {

	perPage := 1
	page := 1

	lcOpts := &gogithub.CommitsListOptions{
		ListOptions: gogithub.ListOptions{
			PerPage: perPage,
			Page:    page,
		},
		SHA:  branch,
		Path: path,
	}
	//Get the latest SHA
	resp, _, err := c.gogithubClient.Repositories.ListCommits(ctx, userName, repoName, lcOpts)
	if err != nil {
		log.Error("Error in obtaining the list of commits", log.Fields{"err": err})
		return "", err
	}
	if len(resp) == 0 {
		log.Debug("File not created yet.", log.Fields{"Latest Commit Array": resp})
		return "", nil
	}
	latestCommitSHA := *resp[0].SHA

	return latestCommitSHA, nil
}

/*
	Function to create a new branch
	params : context, git client,latestCommitSHA, User Name, Repo Name, Branch
	return : error
*/
func createBranch(ctx context.Context, c GitClient, latestCommitSHA, userName, repoName, branch string) error {
	// create a new branch
	ref, _, err := c.gogithubClient.Git.CreateRef(ctx, userName, repoName, &gogithub.Reference{
		Ref: gogithub.String("refs/heads/" + branch),
		Object: &gogithub.GitObject{
			SHA: gogithub.String(latestCommitSHA),
		},
	})
	if err != nil {
		log.Error("Git.CreateRef returned error:", log.Fields{"err": err})
		return err

	}
	log.Info("Branch Created: ", log.Fields{"ref": ref})
	return nil
}
