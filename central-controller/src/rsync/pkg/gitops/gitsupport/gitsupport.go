// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package gitsupport

import (
	"context"
	"fmt"
	"strings"
	"time"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	"github.com/fluxcd/go-git-providers/gitprovider"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	emcogit "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogit"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	v1alpha1 "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"

)

type GitProvider struct {
	Cid       string
	Cluster   string
	App       string
	Namespace string
	Level     string
	GitType   string
	GitToken  string
	UserName  string
	Branch    string
	RepoName  string
	Url       string
	Client    gitprovider.Client
}

/*
	Function to create a New gitProivider
	params : cid, app, cluster, level, namespace string
	return : GitProvider, error
*/
func NewGitProvider(cid, app, cluster, level, namespace string) (*GitProvider, error) {

	result := strings.SplitN(cluster, "+", 2)

	c, err := utils.GetGitOpsConfig(cluster, level, namespace)
	if err != nil {
		return nil, err
	}
	// Read from database
	ccc := db.NewCloudConfigClient()
	refObject, err := ccc.GetClusterSyncObjects(result[0], c.Props.GitOpsReferenceObject)

	if err != nil {
		log.Error("Invalid refObject :", log.Fields{"refObj": c.Props.GitOpsReferenceObject, "error": err})
		return nil, err
	}

	kvRef := refObject.Spec.Kv

	var gitType, gitToken, branch, userName, repoName string

	for _, kvpair := range kvRef {
		log.Info("kvpair", log.Fields{"kvpair": kvpair})
		v, ok := kvpair["gitType"]
		if ok {
			gitType = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["gitToken"]
		if ok {
			gitToken = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["repoName"]
		if ok {
			repoName = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["userName"]
		if ok {
			userName = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["branch"]
		if ok {
			branch = fmt.Sprintf("%v", v)
			continue
		}
	}
	if len(gitType) <= 0 || len(gitToken) <= 0 || len(branch) <= 0 || len(userName) <= 0 || len(repoName) <= 0 {
		log.Error("Missing information for Git", log.Fields{"gitType": gitType, "token": gitToken, "branch": branch, "userName": userName, "repoName": repoName})
		return nil, pkgerrors.Errorf("Missing Information for Git")
	}

	p := GitProvider{
		Cid:       cid,
		App:       app,
		Cluster:   cluster,
		Level:     level,
		Namespace: namespace,
		GitType:   gitType,
		GitToken:  gitToken,
		Branch:    branch,
		UserName:  userName,
		RepoName:  repoName,
		Url:       "https://" + gitType + ".com/" + userName + "/" + repoName,
	}
	client, err := emcogit.CreateClient(gitToken, gitType)
	if err != nil {
		log.Error("Error getting git client", log.Fields{"err": err})
		return nil, err
	}
	p.Client = client.(gitprovider.Client)
	return &p, nil
}

/*
	Function to get path of files stored in git
	params : string
	return : string
*/

func (p *GitProvider) GetPath(t string) string {
	return "clusters/" + p.Cluster + "/" + t + "/" + p.Cid + "/app/" + p.App + "/"
}

/*
	Function to create a new resource if the not already existing
	params : name string, ref interface{}, content []byte
	return : interface{}, error
*/
func (p *GitProvider) Create(name string, ref interface{}, content []byte) (interface{}, error) {

	path := p.GetPath("context") + name + ".yaml"
	ref = emcogit.Add(path, string(content), ref, p.GitType)
	return ref, nil
}

/*
	Function to apply resource to the cluster
	params : name string, ref interface{}, content []byte
	return : interface{}, error
*/
func (p *GitProvider) Apply(name string, ref interface{}, content []byte) (interface{}, error) {

	path := p.GetPath("context") + name + ".yaml"
	ref = emcogit.Add(path, string(content), ref, p.GitType)
	return ref, nil

}

/*
	Function to delete resource from the cluster
	params : name string, ref interface{}, content []byte
	return : interface{}, error
*/
func (p *GitProvider) Delete(name string, ref interface{}, content []byte) (interface{}, error) {

	path := p.GetPath("context") + name + ".yaml"
	ref = emcogit.Delete(path, ref, p.GitType)
	return ref, nil

}

/*
	Function to get resource from the cluster
	params : name string, gvkRes []byte
	return : []byte, error
*/
func (p *GitProvider) Get(name string, gvkRes []byte) ([]byte, error) {

	return []byte{}, nil
}

/*
	Function to commit resources to the cluster
	params : ctx context.Context, ref interface{}
	return : error
*/
func (p *GitProvider) Commit(ctx context.Context, ref interface{}) error {

	var exists bool
	switch ref.(type) {
	case []gitprovider.CommitFile:
		exists = true
	default:
		exists = false

	}
	// Check for rf
	if !exists {
		log.Error("Commit: No ref found", log.Fields{})
		return nil
	}
	err := emcogit.CommitFiles(ctx, p.Client, p.UserName, p.RepoName, p.Branch, "Commit for "+p.GetPath("context"), ref.([]gitprovider.CommitFile), p.GitType)

	return err
}

/*
	Function for cluster reachablity test
	params : null
	return : error
*/
func (p *GitProvider) IsReachable() error {
	return nil
}

// Wait time between reading git status (seconds)
var waitTime int = 60
// StartClusterWatcher watches for CR changes in git location
// go routine starts and reads after waitTime
// Thread exists when the AppContext is deleted
func (p *GitProvider) StartClusterWatcher() error {
	// Start thread to sync monitor CR
	go func() error {
		ctx := context.Background()
		for {
			select {
			case <-time.After(time.Duration(waitTime) * time.Second):
				if ctx.Err() != nil {
					return ctx.Err()
				}
				// Check if AppContext doesn't exist then exit the thread
				if _, err := utils.NewAppContextReference(p.Cid); err != nil {
					// Delete the Status CR updated by Monitor running on the cluster
					p.DeleteClusterStatusCR()
					// AppContext deleted - Exit thread
					return nil
				}
				path :=  p.GetPath("status")
				// Read file
				c, err := emcogit.GetFiles(ctx, p.Client, p.UserName, p.RepoName, p.Branch, path, p.GitType)
				if err != nil {
					log.Error("Status file not available", log.Fields{"error": err, "cluster": p.Cluster, "resource": path})
					continue
				}
				cp := c.([]*gitprovider.CommitFile)
				if len(cp) > 0 {
					// Only one file expected in the location
					content := &v1alpha1.ResourceBundleState{}
					_, err := utils.DecodeYAMLData(*cp[0].Content, content)
					if err != nil {
						log.Error("", log.Fields{"error": err, "cluster": p.Cluster, "resource": path})
						return err
					}
					status.HandleResourcesStatus(p.Cid, p.App, p.Cluster, content)
				}
			// Check if the context is canceled
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}()
	return nil
}

// DeleteClusterStatusCR deletes the status CR provided by the monitor on the cluster
func (p *GitProvider) DeleteClusterStatusCR() error {
	// Delete the status CR
	path := p.GetPath("status") + p.Cid + "-" + p.App
	rf := []gitprovider.CommitFile{}
	ref := emcogit.Delete(path, rf, p.GitType)
	err := emcogit.CommitFiles(context.Background(), p.Client, p.UserName, p.RepoName, p.Branch,
	"Commit for Delete Status CR "+p.GetPath("status"), ref, p.GitType)
	return err
}
