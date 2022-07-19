// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package azurearc

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	emcogit "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/emcogit"
)

const subscriptionURL = "https://management.azure.com/subscriptions/"
const azureCoreURL = "https://management.core.windows.net/"
const azureLoginURL = "https://login.microsoftonline.com/"

type Token struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	ExtExpiresIn string `json:"ext_expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	AccessToken  string `json:"access_token"`
}

type Properties struct {
	RepositoryUrl         string `json:"repositoryUrl"`
	OperatorNamespace     string `json:"operatorNamespace"`
	OperatorInstanceName  string `json:"operatorInstanceName"`
	OperatorType          string `json:"operatorType"`
	OperatorParams        string `json:"operatorParams"`
	OperatorScope         string `json:"operatorScope"`
	SshKnownHostsContents string `json:"sshKnownHostsContents"`
}

type Requestbody struct {
	Properties Properties `json:"properties"`
}

/*
	Function to get the access token for azure arc
	params: clientId, ClientSecret, tenantIdValue
	return: Token, error
*/
func (p *AzureArcProvider) getAccessToken(clientId string, clientSecret string, tenantIdValue string) (string, error) {

	client := http.Client{}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Add("client_id", clientId)
	data.Add("resource", azureCoreURL)
	data.Add("client_secret", clientSecret)

	urlPost := azureLoginURL + tenantIdValue + "/oauth2/token"

	//Rest api to get the access token
	req, err := http.NewRequest("POST", urlPost, bytes.NewBufferString(data.Encode()))
	if err != nil {
		//Handle Error
		log.Error("Couldn't create Azure Access Token request", log.Fields{"err": err, "req": req})
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	res, err := client.Do(req)
	if err != nil {
		log.Error(" Azure Access Token response error", log.Fields{"err": err, "res": res})
		return "", err
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error(" Azure Access Token response marshall error", log.Fields{"err": err, "responseData": responseData})
		return "", err
	}

	// Unmarshall the response body into json and get token value
	newToken := Token{}
	json.Unmarshal(responseData, &newToken)

	return newToken.AccessToken, nil
}

/*
	Function to create a git configuration for the mentioned user repo
	params: accessToken, repositoryUrl, gitConfiguration, operatorScopeType, subscriptionId, Arc Cluster ResourceGroup, Arc ClusterName
			git branch, git path
	return: response, error
*/
func (p *AzureArcProvider) createGitConfiguration(accessToken string, repositoryUrl string, gitConfiguration string, operatorScopeType string, subscriptionIdValue string, arcClusterResourceGroupName string, arcClusterName string, gitbranch string, gitpath string) (string, error) {

	// PUT request for creating git configuration
	client := http.Client{}

	flags := "--git-branch=" + gitbranch + " --git-poll-interval=1s --sync-garbage-collection --git-path=" + gitpath

	// PUT request body
	properties := Requestbody{
		Properties{
			RepositoryUrl:         repositoryUrl,
			OperatorNamespace:     gitConfiguration,
			OperatorInstanceName:  gitConfiguration,
			OperatorType:          "flux",
			OperatorParams:        flags,
			OperatorScope:         operatorScopeType,
			SshKnownHostsContents: ""}}

	dataProperties, err := json.Marshal(properties)
	if err != nil {
		log.Error(" Data properties marshall error", log.Fields{"err": err, "dataProperties": dataProperties})
		return "", err
	}

	urlPut := subscriptionURL + subscriptionIdValue + "/resourceGroups/" + arcClusterResourceGroupName + "/providers/Microsoft.Kubernetes/connectedClusters/" + arcClusterName + "/providers/Microsoft.KubernetesConfiguration/sourceControlConfigurations/" + gitConfiguration + "?api-version=2021-03-01"

	reqPut, err := http.NewRequest(http.MethodPut, urlPut, bytes.NewBuffer(dataProperties))
	if err != nil {
		//Handle Error
		log.Error("Error in creating request for git configuration", log.Fields{"err": err})
		return "", err
	}
	// Add request header
	authorizationString := "Bearer " + accessToken
	reqPut.Header.Set("Content-Type", "application/json; charset=UTF-8")
	reqPut.Header.Add("Authorization", authorizationString)

	resPut, err := client.Do(reqPut)
	if err != nil {
		//Handle Error
		log.Error("Error in response from create git configuration", log.Fields{"err": err})
		return "", err
	}

	responseDataPut, err := ioutil.ReadAll(resPut.Body)
	if err != nil {
		//Handle Error
		log.Error("Error in parsing response from create git configuration", log.Fields{"err": err})
		return "", err
	}

	log.Info("Response for GitConfiguration creation:", log.Fields{"ResponseData": string(responseDataPut)})
	return string(responseDataPut), nil

}

/*
	Function to add dummy file to prevent the path getting deleted
	params : context
	return : error
*/
func (p *AzureArcProvider) addDummyFile(ctx context.Context, gitBranch string) error {

	c, err := emcogit.CreateClient(p.gitProvider.UserName, p.gitProvider.GitToken, p.gitProvider.GitType)

	if err != nil {
		log.Error("Error Creating emcogit client", log.Fields{"err": err})
		return err
	}
	files := []gitprovider.CommitFile{}

	path := "clusters/" + p.gitProvider.Cluster + "/context/" + p.gitProvider.Cid
	files = emcogit.Add(path+"/DoNotDelete", "Dummy file", files, p.gitProvider.GitType).([]gitprovider.CommitFile)

	appName := p.gitProvider.Cid + p.gitProvider.App
	err = emcogit.CommitFiles(ctx, c, p.gitProvider.UserName, p.gitProvider.RepoName, gitBranch, "New Commit", appName, files, p.gitProvider.GitType)
	if err != nil {
		log.Error("Couldn't commit the file", log.Fields{"err": err})
		return err
	}

	return nil
}

/*
	Function to delete dummy file
	params : context
	return : error
*/
func (p *AzureArcProvider) deleteDummyFile(ctx context.Context, gitBranch string) error {

	c, err := emcogit.CreateClient(p.gitProvider.UserName, p.gitProvider.GitToken, p.gitProvider.GitType)

	if err != nil {
		log.Error("Error Creating emcogit client", log.Fields{"err": err})
		return err
	}
	files := []gitprovider.CommitFile{}

	path := "clusters/" + p.gitProvider.Cluster + "/context/" + p.gitProvider.Cid
	files = emcogit.Delete(path+"/DoNotDelete", files, p.gitProvider.GitType).([]gitprovider.CommitFile)

	appName := p.gitProvider.Cid + p.gitProvider.App
	err = emcogit.CommitFiles(ctx, c, p.gitProvider.UserName, p.gitProvider.RepoName, gitBranch, "New Commit", appName, files, p.gitProvider.GitType)
	if err != nil {
		log.Error("Couldn't commit the file", log.Fields{"err": err})
		return err
	}

	return nil
}

/*
	Function to Delete Git configuration
	params : Access Token, Subscription Id, Arc Cluster ResourceName, Arc Cluster Name, Flux Configuration name
	return : Response, error

*/
func (p *AzureArcProvider) deleteGitConfiguration(accessToken string, subscriptionIdValue string, arcClusterResourceGroupName string, arcClusterName string, gitConfiguration string) (string, error) {

	// Create client
	client := &http.Client{}
	// Create request
	urlDelete := subscriptionURL + subscriptionIdValue + "/resourceGroups/" + arcClusterResourceGroupName + "/providers/Microsoft.Kubernetes/connectedClusters/" + arcClusterName + "/providers/Microsoft.KubernetesConfiguration/sourceControlConfigurations/" + gitConfiguration + "?api-version=2021-03-01"

	reqDelete, err := http.NewRequest("DELETE", urlDelete, nil)
	if err != nil {
		//Handle Error
		log.Error("Error in request of delete configuration", log.Fields{"Response": reqDelete, "err": err})
		return "", err
	}
	// Add request header
	authorizationString := "Bearer " + accessToken
	reqDelete.Header.Set("Content-Type", "application/json; charset=UTF-8")
	reqDelete.Header.Add("Authorization", authorizationString)

	resPut, err := client.Do(reqDelete)
	if err != nil {
		//Handle Error
		log.Error("Error in response of delete configuration", log.Fields{"Response": resPut, "err": err})
		return "", err
	}
	responseDataPut, err := ioutil.ReadAll(resPut.Body)
	if err != nil {
		log.Error("Error in parsing response of delete configuration", log.Fields{"Response": responseDataPut, "err": err})
		return "", err
	}

	return string(responseDataPut), nil
}

/*
	Function to create gitconfiguration of fluxv1 type in azure
	params : ctx context.Context, config interface{}
	return : error
*/
func (p *AzureArcProvider) ApplyConfig(ctx context.Context, config interface{}) error {

	//Add dummy file to git
	resp := p.addDummyFile(ctx, p.gitProvider.Branch)
	if resp != nil {
		log.Error("Couldn't add dummy file", log.Fields{"err": resp})
		return resp
	}

	//get accesstoken for azure
	accessToken, err := p.getAccessToken(p.clientID, p.clientSecret, p.tenantID)

	log.Info("Obtained AccessToken: ", log.Fields{"accessToken": accessToken})

	if err != nil {
		log.Error("Couldn't obtain access token", log.Fields{"err": err, "accessToken": accessToken})
		return err
	}

	gitConfiguration := "config-" + p.gitProvider.Cid
	operatorScope := "cluster"
	gitPath := "clusters/" + p.gitProvider.Cluster + "/context/" + p.gitProvider.Cid
	gitBranch := p.gitProvider.Branch

	_, err = p.createGitConfiguration(accessToken, p.gitProvider.Url, gitConfiguration, operatorScope, p.subscriptionID,
		p.arcResourceGroup, p.arcCluster, gitBranch, gitPath)

	if err != nil {
		log.Error("Error in creating git configuration", log.Fields{"err": err})
		return err
	}

	return nil

}

/*
	Function to delete the git configuration
	params : ctx context.Context, config interface{}
	return : error
*/
func (p *AzureArcProvider) DeleteConfig(ctx context.Context, config interface{}) error {

	//get accesstoken for azure
	accessToken, err := p.getAccessToken(p.clientID, p.clientSecret, p.tenantID)

	if err != nil {
		log.Error("Couldn't obtain access token", log.Fields{"err": err, "accessToken": accessToken})
		return err
	}

	gitConfiguration := "config-" + p.gitProvider.Cid

	// to allow enough time for the flux to delete deployments from azure arc
	time.Sleep(time.Duration(p.configDeleteDelay) * time.Second)

	_, err = p.deleteGitConfiguration(accessToken, p.subscriptionID, p.arcResourceGroup, p.arcCluster, gitConfiguration)

	if err != nil {
		log.Error("Error in deleting git configuration", log.Fields{"err": err})
		return err
	}

	//Delete dummy file to git
	resp := p.deleteDummyFile(ctx, p.gitProvider.Branch)
	if resp != nil {
		log.Error("Couldn't delete dummy file", log.Fields{"err": resp})
		return resp
	}

	return nil
}
