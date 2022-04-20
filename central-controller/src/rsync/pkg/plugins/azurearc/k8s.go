// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package azurearc

import (
	"fmt"
	"strings"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"

	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	gitsupport "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/gitops/gitsupport"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
)

type AzureArcProvider struct {
	gitProvider       gitsupport.GitProvider
	clientID          string
	tenantID          string
	clientSecret      string
	subscriptionID    string
	arcCluster        string
	arcResourceGroup  string
	configDeleteDelay int
}

func NewAzureArcProvider(cid, app, cluster, level, namespace string) (*AzureArcProvider, error) {

	result := strings.SplitN(cluster, "+", 2)

	c, err := utils.GetGitOpsConfig(cluster, level, namespace)

	if err != nil {
		return nil, err
	}
	if c.Props.GitOpsType != "azureArc" {
		log.Error("Invalid GitOps type:", log.Fields{})
		return nil, pkgerrors.Errorf("Invalid GitOps type: " + c.Props.GitOpsType)
	}

	// Read from database
	ccc := db.NewCloudConfigClient()

	gitProvider, err := gitsupport.NewGitProvider(cid, app, cluster, level, namespace)
	if err != nil {
		log.Error("Error creating git provider", log.Fields{"err": err, "gitProvider": gitProvider})
		return nil, err
	}

	resObject, err := ccc.GetClusterSyncObjects(result[0], c.Props.GitOpsResourceObject)
	if err != nil {
		log.Error("Invalid resObject :", log.Fields{"resObj": c.Props.GitOpsResourceObject})
		return nil, pkgerrors.Errorf("Invalid resObject: " + c.Props.GitOpsResourceObject)
	}

	kvRes := resObject.Spec.Kv

	var clientID, tenantID, clientSecret, subscriptionID, arcCluster, arcResourceGroup, configDeleteDelayStr string

	for _, kvpair := range kvRes {
		log.Info("kvpair", log.Fields{"kvpair": kvpair})
		v, ok := kvpair["clientID"]
		if ok {
			clientID = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["tenantID"]
		if ok {
			tenantID = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["clientSecret"]
		if ok {
			clientSecret = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["subscriptionID"]
		if ok {
			subscriptionID = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["arcCluster"]
		if ok {
			arcCluster = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["arcResourceGroup"]
		if ok {
			arcResourceGroup = fmt.Sprintf("%v", v)
			continue
		}
		v, ok = kvpair["configDeleteDelay"]
		if ok {
			configDeleteDelayStr = fmt.Sprintf("%v", v)
			continue
		}
	}
	if len(clientID) <= 0 || len(tenantID) <= 0 || len(clientSecret) <= 0 || len(subscriptionID) <= 0 || len(arcCluster) <= 0 || len(arcResourceGroup) <= 0 || len(configDeleteDelayStr) <= 0 {
		log.Error("Missing information for Azure Arc", log.Fields{"clientID": clientID, "tenantID": tenantID, "clientSecret": clientSecret, "subscriptionID": subscriptionID,
			"arcCluster": arcCluster, "arcResourceGroup": arcResourceGroup, "configDeleteDelay": configDeleteDelayStr})
		return nil, pkgerrors.Errorf("Missing Information for Azure Arc")
	}

	var configDeleteDelay int
	_, err = fmt.Sscan(configDeleteDelayStr, &configDeleteDelay)

	if err != nil {
		log.Error("Invalid config delete delay", log.Fields{"configDeleteDelayStr": configDeleteDelayStr, "err": err})
		return nil, err
	}

	p := AzureArcProvider{

		gitProvider:       *gitProvider,
		clientID:          clientID,
		tenantID:          tenantID,
		clientSecret:      clientSecret,
		subscriptionID:    subscriptionID,
		arcCluster:        arcCluster,
		arcResourceGroup:  arcResourceGroup,
		configDeleteDelay: configDeleteDelay,
	}
	return &p, nil
}

func (p *AzureArcProvider) CleanClientProvider() error {
	return nil
}
