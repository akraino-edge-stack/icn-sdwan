// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package k8sexp

import (
	"context"
	"io/ioutil"
	"os"

	pkgerrors "github.com/pkg/errors"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	. "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"strings"
)

type approval struct {
	name string
	res  []byte
}
type reference struct {
	op          RsyncOperation
	filepath    string
	approveList []approval
}

func appendRef(ref interface{}, b []byte, op RsyncOperation, apv ...approval) interface{} {

	var exists bool
	switch ref.(type) {
	case *reference:
		exists = true
	default:
		exists = false

	}
	var rf *reference
	// Create rf is doesn't exist
	if !exists {
		f, err := ioutil.TempFile("/tmp", "k8sexp"+"-")
		if err != nil {
			log.Error("Unable to create temp file in tmp directory", log.Fields{"err": err})
			return nil
		}
		rf = &reference{
			op:       op,
			filepath: f.Name(),
		}
	} else {
		rf = ref.(*reference)
	}
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(rf.filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Error opening file", log.Fields{"err": err, "filepath": rf.filepath})
		return rf
	}
	defer f.Close()
	// Write seperator
	if _, err := f.Write([]byte("\n---\n")); err != nil {
		log.Error("Error appending file", log.Fields{"err": err, "filepath": rf.filepath})
		return rf
	}
	if _, err := f.Write(b); err != nil {
		log.Error("Error appending file", log.Fields{"err": err, "filepath": rf.filepath})
	}
	rf.approveList = append(rf.approveList, apv...)
	return rf
}

// Creates a new resource if the not already existing
func (p *K8sProviderExp) Create(name string, ref interface{}, content []byte) (interface{}, error) {
	// Add the label based on the Status Appcontext ID
	label := p.cid + "-" + p.app
	b, err := status.TagResource(content, label)
	if err != nil {
		log.Error("Error Tag Resoruce with label:", log.Fields{"err": err, "label": label, "resource": name})
		return nil, err
	}
	ref = appendRef(ref, b, OpCreate)
	return ref, nil
}

// Apply resource to the cluster
func (p *K8sProviderExp) Apply(name string, ref interface{}, content []byte) (interface{}, error) {
	var apv approval
	// Add the label based on the Status Appcontext ID
	label := p.cid + "-" + p.app
	b, err := status.TagResource(content, label)
	if err != nil {
		log.Error("Error Tag Resoruce with label:", log.Fields{"err": err, "label": label, "resource": name})
		return nil, err
	}
	// Check if subresource
	acUtils, err := utils.NewAppContextReference(p.cid)
	if err != nil {
		return ref, err
	}
	// Currently only subresource supported is approval
	subres, _, err := acUtils.GetSubResApprove(name, p.app, p.cluster)
	if err == nil {
		result := strings.Split(name, "+")
		if result[0] == "" {
			return nil, pkgerrors.Errorf("Resource name is nil %s:", name)
		}
		log.Info("Approval Subresource::", log.Fields{"cluster": p.cluster, "resource": result[0], "approval": string(subres)})
		apv = approval{name: result[0], res: subres}
		ref = appendRef(ref, b, OpApply, apv)
	} else {
		ref = appendRef(ref, b, OpApply)
	}
	return ref, nil
}

// Delete resource from the cluster
func (p *K8sProviderExp) Delete(name string, ref interface{}, content []byte) (interface{}, error) {
	ref = appendRef(ref, content, OpDelete)
	return ref, nil

}

// Get resource from the cluster
func (p *K8sProviderExp) Get(name string, gvkRes []byte) ([]byte, error) {
	b, err := p.client.Get(gvkRes, p.namespace)
	if err != nil {
		log.Error("Failed to get res", log.Fields{"error": err, "resource": name})
		return nil, err
	}
	return b, nil
}

// Commit resources to the cluster
func (p *K8sProviderExp) Commit(ctx context.Context, ref interface{}) error {
	var exists bool
	switch ref.(type) {
	case *reference:
		exists = true
	default:
		exists = false

	}
	// Check for rf
	if !exists {
		log.Error("Commit: No ref found", log.Fields{})
		return nil
	}
	rf := ref.(*reference)
	log.Info("Commit:: Ref found", log.Fields{"filename": rf.filepath, "op": rf.op})
	defer os.Remove(rf.filepath)
	content, err := ioutil.ReadFile(rf.filepath)
	if err != nil {
		log.Error("File read failed", log.Fields{"Error": err})
		return err
	}
	switch rf.op {
	case OpDelete:
		if err := p.client.Delete(content); err != nil {
			log.Error("Failed to delete resources", log.Fields{"error": err})
			return err
		}
	case OpApply:
		if err := p.client.Apply(content); err != nil {
			log.Error("Failed to apply resources", log.Fields{"error": err})
			return err
		}
		//Check if approval list is not 0
		if len(rf.approveList) > 0 {
			for _, apv := range rf.approveList {
				if err := p.client.Approve(apv.name, apv.res); err != nil {
					log.Error("Failed to approve resources", log.Fields{"error": err})
					return err
				}
			}
		}
	case OpCreate:
		if err := p.client.Create(content); err != nil {
			if apierrors.IsAlreadyExists(err) {
				log.Warn("Resources is already present, Skipping", log.Fields{"error": err})
				return nil
			} else {
				log.Error("Failed to create resources", log.Fields{"error": err})
				return err
			}
		}
	}
	return nil
}

// IsReachable cluster reachablity test
func (p *K8sProviderExp) IsReachable() error {
	return p.client.IsReachable()
}

func (m *K8sProviderExp) TagResource(res []byte, label string) ([]byte, error) {
	b, err := status.TagResource(res, label)
	if err != nil {
		log.Error("Error Tag Resoruce with label:", log.Fields{"err": err, "label": label, "resource": res})
		return nil, err
	}
	return b, nil
}
