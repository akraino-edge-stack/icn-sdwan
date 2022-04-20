// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package emcogit

import (
	"context"

	emcogithub "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogithub"

	"github.com/fluxcd/go-git-providers/gitprovider"
	pkgerrors "github.com/pkg/errors"
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
	Helper function to convert interface to gitprovider.Client
	params: files interface{}
	return: gitprovider.Client
*/
func convertToClient(c interface{}) gitprovider.Client {
	return c.(gitprovider.Client)
}

/*
	Function to create git client
	params : git token
	return : git client, error
*/
func CreateClient(gitToken string, gitType string) (interface{}, error) {
	switch gitType {
	case "github":
		c, err := emcogithub.CreateClient(gitToken)
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
		err := emcogithub.CreateRepo(ctx, convertToClient(c), repoName, userName, desc)
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
		err := emcogithub.DeleteRepo(ctx, convertToClient(c), userName, repoName)
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
	params : context, git client, User Name, Repo Name, Branch Name, Commit Message, files
	return : nil/error
*/
func CommitFiles(ctx context.Context, c interface{}, userName string, repoName string, branch string, commitMessage string, files interface{}, gitType string) error {

	switch gitType {
	case "github":
		err := emcogithub.CommitFiles(ctx, convertToClient(c), userName, repoName, branch, commitMessage, convertToCommitFile(files))
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
		ref, err := emcogithub.GetFiles(ctx, convertToClient(c), userName, repoName, branch, path)
		return ref, err
	}
	//Add other types like gitlab, bitbucket etc
	return nil, nil
}
