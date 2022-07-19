// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package emcogit

import (
	"context"

	"github.com/fluxcd/go-git-providers/gitprovider"
	pkgerrors "github.com/pkg/errors"
	emcogithub "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogithub"
)

/*
	Helper function to convert interface to []gitprovider.CommitFile
	params: files interface{}
	return: []gitprovider.CommitFile
*/
func convertToCommitFile(ref interface{}) []gitprovider.CommitFile {
	var exists bool
	switch ref.(type) {
	case []gitprovider.CommitFile:
		exists = true
	default:
		exists = false
	}
	var rf []gitprovider.CommitFile
	// Create rf is doesn't exist
	if !exists {
		rf = []gitprovider.CommitFile{}
	} else {
		rf = ref.([]gitprovider.CommitFile)
	}
	return rf
}

/*
	Helper function to convert interface to emcogithub.GithubClient
	params: files interface{}
	return: emcogithub.GithubClient
*/
func convertToClient(c interface{}) emcogithub.GithubClient {
	return c.(emcogithub.GithubClient)
}

/*
	Function to create git client
	params : git token
	return : git client, error
*/
func CreateClient(userName, gitToken, gitType string) (interface{}, error) {
	switch gitType {
	case "github":
		c, err := emcogithub.CreateClient(userName, gitToken)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	//Add other types like gitlab, bitbucket etc
	return nil, pkgerrors.New("Git Provider type not supported")
}

/*
	Function to create a new Repo in github
	params : context, git client, Repository Name, User Name, description
	return : nil/error
*/
func CreateRepo(ctx context.Context, c interface{}, repoName string, userName string, desc string, gitType string) error {

	switch gitType {
	case "github":
		err := emcogithub.CreateRepo(ctx, c, repoName, userName, desc)
		if err != nil {
			return err
		}
		return nil
	}
	//Add other types like gitlab, bitbucket etc
	return pkgerrors.New("Git Provider type not supported")

}

/*
	Function to delete repo
	params : context, git client , user name, repo name
	return : nil/error
*/
func DeleteRepo(ctx context.Context, c interface{}, userName string, repoName string, gitType string) error {

	switch gitType {
	case "github":
		err := emcogithub.DeleteRepo(ctx, c, userName, repoName)
		if err != nil {
			return err
		}
		return nil
	}
	//Add other types like gitlab, bitbucket etc
	return pkgerrors.New("Git Provider type not supported")
}

/*
	Function to commit multiple files to the github repo
	params : context, git client, User Name, Repo Name, Branch Name, Commit Message, appName, files
	return : nil/error
*/
func CommitFiles(ctx context.Context, c interface{}, userName, repoName, branch, commitMessage, appName string, files interface{}, gitType string) error {

	switch gitType {
	case "github":
		err := emcogithub.CommitFiles(ctx, c, userName, repoName, branch, commitMessage, appName, convertToCommitFile(files))
		if err != nil {
			return err
		}
		return nil
	}
	//Add other types like gitlab, bitbucket etc
	return pkgerrors.New("Git Provider type not supported")
}

/*
	Function to Add file to the commit
	params : path , content, files
	return : files
*/
func Add(path string, content string, files interface{}, gitType string) interface{} {
	switch gitType {
	case "github":
		ref := emcogithub.Add(path, content, convertToCommitFile(files))
		return ref
	}
	//Add other types like gitlab, bitbucket etc
	return nil
}

/*
	Function to Delete file from the commit
	params : path, files
	return : files
*/
func Delete(path string, files interface{}, gitType string) interface{} {
	switch gitType {
	case "github":
		ref := emcogithub.Delete(path, convertToCommitFile(files))
		return ref
	}
	//Add other types like gitlab, bitbucket etc
	return nil
}

func GetFiles(ctx context.Context, c interface{}, userName, repoName, branch, path, gitType string) (interface{}, error) {
	switch gitType {
	case "github":
		ref, err := emcogithub.GetFiles(ctx, c, userName, repoName, branch, path)
		return ref, err
	}
	//Add other types like gitlab, bitbucket etc
	return nil, nil
}

/*
	Function to obtaion the SHA of latest commit
	params : context, go git client, User Name, Repo Name, Branch, Path, gitType
	return : LatestCommit string, error
*/
func GetLatestCommitSHA(ctx context.Context, c interface{}, userName, repoName, branch, path, gitType string) (string, error) {
	switch gitType {
	case "github":
		latestCommitSHA, err := emcogithub.GetLatestCommitSHA(ctx, c, userName, repoName, branch, path)
		return latestCommitSHA, err
	}
	//Add other types like gitlab, bitbucket etc
	return "", nil
}

/*
	Function to check if file exists
	params : context, go git client, User Name, Repo Name, Branch, Path, gitType
	return : LatestCommit string, error
*/
func CheckIfFileExists(ctx context.Context, c interface{}, userName, repoName, branch, path, gitType string) (bool, error) {
	switch gitType {
	case "github":
		check, err := emcogithub.CheckIfFileExists(ctx, c, userName, repoName, branch, path)
		return check, err
	}
	//Add other types like gitlab, bitbucket etc
	return false, nil
}

/*
	Function to delete the branch
	params : context, go git client, User Name, Repo Name, mergeBranch, gitType
	return : LatestCommit string, error
*/
func DeleteBranch(ctx context.Context, c interface{}, userName, repoName, mergeBranch, gitType string) error {
	switch gitType {
	case "github":
		err := emcogithub.DeleteBranch(ctx, c, userName, repoName, mergeBranch)
		if err != nil {
			return err
		}
		return nil
	}
	//Add other types like gitlab, bitbucket etc
	return pkgerrors.New("Git Provider type not supported")
}

func CreateBranch(ctx context.Context, c interface{}, latestCommitSHA, userName, repoName, branch, gitType string) error {
	switch gitType {
	case "github":
		err := emcogithub.CreateBranch(ctx, c, latestCommitSHA, userName, repoName, branch)
		if err != nil {
			return err
		}
		return nil
	}
	//Add other types like gitlab, bitbucket etc
	return pkgerrors.New("Git Provider type not supported")
}
