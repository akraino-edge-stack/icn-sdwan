// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package context

import (
	//	"bytes"
	"encoding/json"
	"fmt"

	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	. "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"
)

// CreateCompApp creates a AppContext for a composite app, for testing
func CreateCompApp(ca CompositeApp) (string, error) {

	var compositeHandle interface{}
	var err error

	context := appcontext.AppContext{}
	ctxval, err := context.InitAppContext()
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}
	contextID := fmt.Sprintf("%v", ctxval)

	if compositeHandle, err = context.CreateCompositeApp(); err != nil {
		return "", pkgerrors.Wrap(err, "Error creating CompositeApp handle")
	}

	if err = context.AddCompositeAppMeta(ca.CompMetadata); err != nil {
		return "", pkgerrors.Wrap(err, "Error Adding CompositeAppMeta")
	}
	appOrder, err := json.Marshal(map[string][]string{"apporder": ca.AppOrder})
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error adding app order instruction")
	}
	_, err = context.AddInstruction(compositeHandle, "app", "order", string(appOrder))
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error adding app order instruction")
	}

	for _, app := range ca.Apps {
		a, err := context.AddApp(compositeHandle, app.Name)
		if err != nil {
			return "", pkgerrors.Wrap(err, "Error Adding App")
		}
		if app.Dependency != nil {
			var dep []AppCriteria
			for i, j := range app.Dependency {
				l := AppCriteria{App: i, OpStatus: j.OpStatus, Wait: j.Wait}
				dep = append(dep, l)
			}
			dependency, err := json.Marshal(dep)
			if err != nil {
				return "", pkgerrors.Wrap(err, "Error adding depencency instruction")
			}
			_, err = context.AddLevelValue(a, "instruction/dependency", string(dependency))
			if err != nil {
				return "", pkgerrors.Wrap(err, "Error Adding depencency instruction")
			}
		}
		for _, cluster := range app.Clusters {
			c, err := context.AddCluster(a, cluster.Name)
			if err != nil {
				return "", pkgerrors.Wrap(err, "Error Adding Cluster")
			}
			resOrder, err := json.Marshal(map[string][]string{"resorder": cluster.ResOrder})
			_, err = context.AddInstruction(c, "resource", "order", string(resOrder))
			if err != nil {
				return "", pkgerrors.Wrap(err, "Error Adding resorder")
			}
			resdependency, err := json.Marshal(map[string]map[string][]string{"resdependency": cluster.Dependency})
			_, err = context.AddInstruction(c, "resource", "dependency", string(resdependency))
			if err != nil {
				return "", pkgerrors.Wrap(err, "Error Adding resdependency")
			}
			for _, res := range cluster.Resources {
				_, err = context.AddResource(c, res.Name, res.Data)
				if err != nil {
					return "", pkgerrors.Wrap(err, "Error Adding Resource")
				}
			}
		}
	}
	return contextID, nil
}

// ReadAppContext reads a composite app for AppContext
func ReadAppContext(contextID interface{}) (CompositeApp, error) {
	var ca CompositeApp

	acID := fmt.Sprintf("%v", contextID)
	ac := appcontext.AppContext{}
	_, err := ac.LoadAppContext(acID)
	if err != nil {
		logutils.Error("", logutils.Fields{"err": err})
		return CompositeApp{}, err
	}

	caMeta, err := ac.GetCompositeAppMeta()
	// ignore error (in case appcontext has no metadata) VERIFY
	if err == nil {
		ca.CompMetadata = caMeta
	}

	appsOrder, err := ac.GetAppInstruction("order")
	if err != nil {
		return CompositeApp{}, err
	}
	var appList map[string][]string
	json.Unmarshal([]byte(appsOrder.(string)), &appList)
	ca.AppOrder = appList["apporder"]
	appsList := make(map[string]*App)
	for _, app := range appList["apporder"] {
		clusterNames, err := ac.GetClusterNames(app)
		if err != nil {
			return CompositeApp{}, err
		}
		//var clusterList []Cluster
		clusterList := make(map[string]*Cluster)
		for k := 0; k < len(clusterNames); k++ {
			cluster := clusterNames[k]
			resorder, err := ac.GetResourceInstruction(app, cluster, "order")
			if err != nil {
				logutils.Error("Resorder not found for cluster ", logutils.Fields{"cluster": cluster})
				// In Status AppContext some clusters may not have resorder
				// Only used to collect status
				continue
			}
			var aov map[string][]string
			json.Unmarshal([]byte(resorder.(string)), &aov)
			resList := make(map[string]*AppResource)
			for _, res := range aov["resorder"] {
				r := &AppResource{Name: res}
				resList[res] = r
			}
			clusterList[cluster] = &Cluster{Name: cluster, Resources: resList, ResOrder: aov["resorder"]}
			resdependency, err := ac.GetResourceInstruction(app, cluster, "dependency")
			if err != nil {
				// Not all applications have dependency
				continue
			}
			var dov map[string]map[string][]string
			err = json.Unmarshal([]byte(resdependency.(string)), &dov)
			if err != nil {
				logutils.Error("Res dependency Marshalling error, ignoring ", logutils.Fields{"cluster": cluster, "dependency": resdependency.(string)})
				continue
			}
			for _, lt := range dov["resdependency"] {
				for _, res := range lt {
					r := &AppResource{Name: res}
					//resList = append(resList, r)
					resList[res] = r
				}
			}
			clusterList[cluster] = &Cluster{Name: cluster, Resources: resList, ResOrder: aov["resorder"], Dependency: dov["resdependency"]}
		}
		dep, err := ac.GetAppLevelInstruction(app, "dependency")
		depList := make(map[string]*Criteria)
		if err == nil {
			var a []AppCriteria
			// If instruction available read it
			json.Unmarshal([]byte(dep.(string)), &a)
			for _, crt := range a {
				depList[crt.App] = &Criteria{Wait: crt.Wait, OpStatus: crt.OpStatus}
			}
		}
		appsList[app] = &App{Name: app, Clusters: clusterList, Dependency: depList}
	}
	ca.Apps = appsList
	return ca, nil
}

// PrintCompositeApp prints the composite app
func PrintCompositeApp(ca CompositeApp) {

	fmt.Printf("Metadata: %v\n", ca.CompMetadata)
	fmt.Printf("AppOrder: %v\n", ca.AppOrder)
	for _, app := range ca.Apps {
		fmt.Println("")
		fmt.Println("  App: ", app.Name)
		fmt.Println("  Dependency: ", app.Dependency)
		for _, cluster := range app.Clusters {

			fmt.Println("    Cluster: ", cluster.Name)
			fmt.Printf("      ResourceOrder: %v\n", cluster.ResOrder)
			fmt.Println("      Resources: ")
			for _, res := range cluster.Resources {
				fmt.Printf("        %v\n", res.Name)
			}
		}
	}
}

// FindApp finds app in the appcontext and returns true or false
func FindApp(ca CompositeApp, app string) bool {
	for _, a := range ca.Apps {
		if app == a.Name {
			return true
		}
	}
	return false
}

// FindCluster finds cluster in an app in the appcontext and returns true or false
func FindCluster(ca CompositeApp, app string, cluster string) bool {
	for _, a := range ca.Apps {
		if app == a.Name {
			for _, c := range ca.Apps[app].Clusters {
				if cluster == c.Name {
					return true
				}
			}
		}
	}
	return false
}

// FindCluster finds resource in a cluster in an app in the appcontext and returns true or false
func FindResource(ca CompositeApp, app, cluster, res string) bool {
	for _, a := range ca.Apps {
		if app == a.Name {
			for _, c := range ca.Apps[app].Clusters {
				if cluster == c.Name {
					for _, r := range ca.Apps[app].Clusters[cluster].Resources {
						if res == r.Name {
							return true
						}
					}
				}
			}
		}
	}
	return false
}
