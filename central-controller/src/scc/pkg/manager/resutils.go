/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package manager

import (
	"crypto/sha256"

	rsyncclient "github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/client"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/resource"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/resourcestatus"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/rpc"
	controller "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/module/controller"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"

	"encoding/json"
	"fmt"
	pkgerrors "github.com/pkg/errors"
	"log"
	"sync"
	"time"
)

var rsync_initialized = false
var provider_name = "akraino_scc"
var project_name = "akraino_scc"
var Resource_mux  = sync.Mutex{}

// sdewan definition
type DeployResource struct {
	Action   string
	Resource resource.ISdewanResource
	Status   int // 0: to be (un)deployed; 1: success; 2: failed
}

type DeployResources struct {
	Resources []DeployResource
}

type ReadResource struct {
	Gvk       schema.GroupVersionKind `json:"GVK,omitempty"`
	Name      string                  `json:"name,omitempty"`
	Namespace string                  `json:"namespace,omitempty"`
}

type QueryResource struct {
	Handle   interface{}
	Resource ReadResource
}

type QueryResources struct {
	Resources []*QueryResource
}

type ResUtil struct {
	resmap    map[module.ControllerObject]*DeployResources
	qryResmap map[module.ControllerObject]*QueryResources
	qryCtxId  string
}

func NewResUtil() *ResUtil {
	if rsync_initialized == false {
		rsync_initialized = InitRsyncClient()
	}

	return &ResUtil{
		resmap:    make(map[module.ControllerObject]*DeployResources),
		qryResmap: make(map[module.ControllerObject]*QueryResources),
		qryCtxId:  "",
	}
}

type contextForCompositeApp struct {
	context            appcontext.AppContext
	ctxval             interface{}
	compositeAppHandle interface{}
}

func makeAppContextForCompositeApp(p, ca, v, rName, dig string, namespace string, level string) (contextForCompositeApp, error) {
	// ctxval: context.rtcObj.id
	context := appcontext.AppContext{}
	ctxval, err := context.InitAppContext()
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}
	compositeHandle, err := context.CreateCompositeApp()
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating CompositeApp handle")
	}
	err = context.AddCompositeAppMeta(appcontext.CompositeAppMeta{Project: p, CompositeApp: ca, Version: v, Release: rName, DeploymentIntentGroup: dig, Namespace: namespace, Level: level})
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error Adding CompositeAppMeta")
	}

	//_, err = context.GetCompositeAppMeta()

	log.Println(":: The meta data stored in the runtime context :: ")

	cca := contextForCompositeApp{context: context, ctxval: ctxval, compositeAppHandle: compositeHandle}

	return cca, nil
}

func getResourceName(resource DeployResource) string {
	return resource.Resource.GetName() + "+" + resource.Resource.GetType()
}

func addResourcesToCluster(ct appcontext.AppContext, ch interface{}, target string, resources []DeployResource, isDeploy bool) error {

	var resOrderInstr struct {
		Resorder []string `json:"resorder"`
	}

	var resDepInstr struct {
		Resdep map[string][]string `json:"resdependency"`
	}
	resdep := make(map[string][]string)

	for _, resource := range resources {
		resource_name := getResourceName(resource)
		resource_data := resource.Resource.ToYaml(target)
		resOrderInstr.Resorder = append(resOrderInstr.Resorder, resource_name)
		resdep[resource_name] = []string{}

		rh, err := ct.AddResource(ch, resource_name, resource_data)
		if isDeploy == false {
			//Delete resource
			ct.AddLevelValue(rh, "status", resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Applied})
		}
		if err != nil {
			cleanuperr := ct.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Printf(":: Error Cleaning up AppContext after add resource failure ::")
			}
			return pkgerrors.Wrapf(err, "Error adding resource ::%s to AppContext", resource_name)
		}
		jresOrderInstr, _ := json.Marshal(resOrderInstr)
		resDepInstr.Resdep = resdep
		jresDepInstr, _ := json.Marshal(resDepInstr)

		_, err = ct.AddInstruction(ch, "resource", "order", string(jresOrderInstr))
		_, err = ct.AddInstruction(ch, "resource", "dependency", string(jresDepInstr))
		if err != nil {
			cleanuperr := ct.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Printf(":: Error Cleaning up AppContext after add instruction failure ::")
			}
			return pkgerrors.Wrapf(err, "Error adding instruction for resource ::%s to AppContext", resource_name)
		}
	}
	return nil
}

func InitRsyncClient() bool {
	client := controller.NewControllerClient("resources", "data", "orchestrator")

	vals, _ := client.GetControllers()
	found := false
	for _, v := range vals {
		if v.Metadata.Name == "rsync" {
			log.Println("Initializing RPC connection to resource synchronizer")
			rpc.UpdateRpcConn(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			found = true
			break
		}
	}
	return found
}

func initializeAppContextStatus(ac appcontext.AppContext, acStatus appcontext.AppContextStatus) error {
	h, err := ac.GetCompositeAppHandle()
	if err != nil {
		return err
	}
	sh, err := ac.GetLevelHandle(h, "status")
	if sh == nil {
		_, err = ac.AddLevelValue(h, "status", acStatus)
	} else {
		err = ac.UpdateValue(sh, acStatus)
	}
	if err != nil {
		return err
	}
	return nil
}

func (d *ResUtil) contains(reses []DeployResource, res DeployResource) bool {
	for _, r := range reses {
		if r.Action == res.Action &&
			r.Resource.GetName() == res.Resource.GetName() &&
			r.Resource.GetType() == res.Resource.GetType() {
			return true
		}
	}

	return false
}

func (d *ResUtil) GetResources() map[module.ControllerObject]*DeployResources {
	return d.resmap
}

func (d *ResUtil) AddResource(device module.ControllerObject, action string, resource resource.ISdewanResource) error {
	if d.resmap[device] == nil {
		d.resmap[device] = &DeployResources{Resources: []DeployResource{}}
	}

	ds := DeployResource{Action: action, Resource: resource, Status: 0}
	if !d.contains(d.resmap[device].Resources, ds) {
		d.resmap[device].Resources = append(d.resmap[device].Resources, ds)
	}
	return nil
}

func (d *ResUtil) TargetName(o module.ControllerObject) string {
	return o.GetType() + "." + o.GetMetadata().Name
}

func (d *ResUtil) getDeviceAppName(device module.ControllerObject) string {
	return device.GetMetadata().Name + "-app"
}

func (d *ResUtil) getDeviceClusterName(device module.ControllerObject) string {
	return provider_name + "+" + device.GetMetadata().Name
}

func (d *ResUtil) DeployOneResource(app_name string, format string, device module.ControllerObject, resource DeployResource) (string, error) {
	resource_app_name := app_name + resource.Resource.GetName()
	cca, err := makeAppContextForCompositeApp(project_name, resource_app_name, "1.0", "1.0", "di", "default", "0")
	context := cca.context                    // appcontext.AppContext
	ctxval := cca.ctxval                      // id
	compositeHandle := cca.compositeAppHandle // cid

	var appOrderInstr struct {
		Apporder []string `json:"apporder"`
	}
	var appDepInstr struct {
		Appdep map[string]string `json:"appdependency"`
	}
	appdep := make(map[string]string)
	device_app_name := d.getDeviceAppName(device)
	appOrderInstr.Apporder = append(appOrderInstr.Apporder, device_app_name)
	appdep[device_app_name] = ""

	apphandle, _ := context.AddApp(compositeHandle, device_app_name)

	clusterhandle, _ := context.AddCluster(apphandle, d.getDeviceClusterName(device))
	err = addResourcesToCluster(context, clusterhandle, d.TargetName(device), []DeployResource{resource}, true)

	jappOrderInstr, _ := json.Marshal(appOrderInstr)
	appDepInstr.Appdep = appdep
	jappDepInstr, _ := json.Marshal(appDepInstr)
	context.AddInstruction(compositeHandle, "app", "order", string(jappOrderInstr))
	context.AddInstruction(compositeHandle, "app", "dependency", string(jappDepInstr))

	// invoke deployment process
	appContextID := fmt.Sprintf("%v", ctxval)
	err = rsyncclient.InvokeInstallApp(appContextID)
	if err != nil {
		log.Println(err)
		cleanuperr := context.DeleteCompositeApp()
		if cleanuperr != nil {
			log.Printf(":: Error Cleaning up AppContext after add instruction failure ::")
		}

		return "", err
	}

	return appContextID, nil
}

func (d *ResUtil) UpdateOneResource(cid string, device module.ControllerObject, resourceName string, resourceValue string) error {
	context := appcontext.AppContext{}
	_, err := context.LoadAppContext(cid)
	if err != nil {
		return err
	}

	rh, err := context.GetResourceHandle(d.getDeviceAppName(device), d.getDeviceClusterName(device), resourceName)
	if err != nil {
		return err
	}

	err = context.UpdateResourceValue(rh, resourceValue)
	if err != nil {
		return err
	}

	err = rsyncclient.InvokeInstallApp(cid)
	return err
}

func (d *ResUtil) Deploy(overlay string, app_name string, format string) error {
	return d.DeployUpdate(overlay, app_name, format, false)
}

func (d *ResUtil) DeployUpdate(overlay string, app_name string, format string, update bool) error {
	isErr := false
	errMessage := "Failed:"
	res_manager := GetManagerset().Resource
	m := make(map[string]string)
	m[OverlayResource] = overlay

	Resource_mux.Lock()
	defer Resource_mux.Unlock()

	for device, res := range d.resmap {
		m[DeviceResource] = device.GetType() + "." + device.GetMetadata().Name

		for _, resource := range res.Resources {
			operation := 1
			m["Name"] = resource.Resource.GetName()
			m["Type"] = resource.Resource.GetType()
			robj, err := res_manager.GetObject(m)
			resobj := robj.(*module.ResourceObject)
			if err != nil {
				// create a new resource object
				resobj.Metadata.Name = m["Name"]
				resobj.Specification.Hash = ""
				resobj.Specification.ContextId = ""
				resobj.Specification.Ref = 0
				resobj.Specification.Status = Resource_Status_NotDeployed
			}

			resource_data := resource.Resource.ToYaml(d.TargetName(device))
			resource_data_hash_byte := sha256.Sum256([]byte(resource_data))
			resource_data_hash := string(resource_data_hash_byte[:])
			if resobj.Specification.Ref > 0 && resource_data_hash != resobj.Specification.Hash {
				operation = 2
			}

			switch operation {
			case 1:
				// Add resource
				if resource.Status != 1 {
					// resource is not deployed or failed to deploy
					if resobj.Specification.Ref == 0 {
						// resource needs to be deployed	
						cid, err := d.DeployOneResource(app_name, format, device, resource)

						if err != nil {
							isErr = true
							resource.Status = 2
							errMessage = errMessage + " " + resource.Resource.GetName()
						} else {
							resource.Status = 1
							resobj.Specification.Hash = resource_data_hash
							resobj.Specification.ContextId = cid
							resobj.Specification.Ref = 1
							resobj.Specification.Status = Resource_Status_Deployed

							res_manager.CreateObject(m, resobj)
						}
					} else {
						// add ref
						if !update {
							resobj.Specification.Ref += 1
						}
						resource.Status = 1
						res_manager.UpdateObject(m, resobj)
					}
				}
			case 2:
				// Update resource
				if resource.Status != 1 {
					err := d.UpdateOneResource(resobj.Specification.ContextId, device, getResourceName(resource), resource_data)
					if err != nil {
						isErr = true
						resource.Status = 2
						errMessage = errMessage + " " + resource.Resource.GetName()
						log.Println(err)
					} else {
						resource.Status = 1
						// add ref
						resobj.Specification.Hash = resource_data_hash
						if !update {
							resobj.Specification.Ref += 1
						}

						res_manager.UpdateObject(m, resobj)
					}
				}
			default:
				log.Println("Unknown operation type")
			}
		}
	}

	if isErr {
		return pkgerrors.New(errMessage)
	}
	return nil
}

func (d *ResUtil) Undeploy(overlay string) error {
	isErr := false
	errMessage := "Failed:"
	res_manager := GetManagerset().Resource
	m := make(map[string]string)
	m[OverlayResource] = overlay

	Resource_mux.Lock()
	defer Resource_mux.Unlock()

	for device, res := range d.resmap {
		m[DeviceResource] = device.GetType() + "." + device.GetMetadata().Name

		// Use reversed order to do undeploy
		for i:=len(res.Resources)-1; i>=0; i-- {
		// for _, resource := range res.Resources {
			resource := res.Resources[i]
			m["Name"] = resource.Resource.GetName()
			m["Type"] = resource.Resource.GetType()
			robj, err := res_manager.GetObject(m)
			resobj := robj.(*module.ResourceObject)
			if err != nil || resobj.Specification.Ref <= 0 {
				// resource had not been deployed before, nothing to do
				log.Println("Resource " + resource.Resource.GetName() + " hasn't been deployed, ignore the operation")
				continue
			}

			if resource.Status != 1 {
				// resource is not undeployed or failed to undeploy
				if resobj.Specification.Ref <= 1 {
					err = rsyncclient.InvokeUninstallApp(resobj.Specification.ContextId)
					if err != nil {
						log.Println(err)
						isErr = true
						resource.Status = 2
						errMessage = errMessage + " " + resource.Resource.GetName()
					} else {
						// reset resource status
						resource.Status = 1
						resobj.Specification.Ref = 0

						/*
						// delete app from context db
						context := appcontext.AppContext{}
						_, err := context.LoadAppContext(resobj.Specification.ContextId)
						if err != nil {
							err = context.DeleteCompositeApp()
						}

						if err != nil {
							log.Println(err)
						}
						*/
						res_manager.DeleteObject(m)
					}
				} else {
					resobj.Specification.Ref -= 1
					resource.Status = 1
					res_manager.UpdateObject(m, resobj)
				}
			}
		}
	}

	if isErr {
		return pkgerrors.New(errMessage)
	}
	return nil
}

func (d *ResUtil) AddQueryResource(device module.ControllerObject, resource QueryResource) error {
	if d.qryResmap[device] == nil {
		d.qryResmap[device] = &QueryResources{Resources: []*QueryResource{}}
	}

	d.qryResmap[device].Resources = append(d.qryResmap[device].Resources, &resource)
	d.qryCtxId = ""

	return nil
}

func addQueryResourcesToCluster(ct appcontext.AppContext, ch interface{}, resources *[]*QueryResource) error {

	var resOrderInstr struct {
		Resorder []string `json:"resorder"`
	}

	var resDepInstr struct {
		Resdep map[string]string `json:"resdependency"`
	}
	resdep := make(map[string]string)

	for _, resource := range *resources {
		resource_name := resource.Resource.Namespace + "+" + resource.Resource.Name
		v, _ := json.Marshal(resource.Resource)
		resOrderInstr.Resorder = append(resOrderInstr.Resorder, resource_name)
		resdep[resource_name] = "go"

		rh, err := ct.AddResource(ch, resource_name, string(v))

		if err != nil {
			return pkgerrors.Wrapf(err, "Error adding resource ::%s to AppContext", resource_name)
		}

		// save the resource handler for query result
		resource.Handle = rh

		jresOrderInstr, _ := json.Marshal(resOrderInstr)
		resDepInstr.Resdep = resdep
		jresDepInstr, _ := json.Marshal(resDepInstr)

		_, err = ct.AddInstruction(ch, "resource", "order", string(jresOrderInstr))
		_, err = ct.AddInstruction(ch, "resource", "dependency", string(jresDepInstr))
		if err != nil {
			return pkgerrors.Wrapf(err, "Error adding instruction for resource ::%s to AppContext", resource_name)
		}
	}
	return nil
}

func (d *ResUtil) Query(app_name string) (string, error) {
	if d.qryCtxId == "" {
		// Generate Application context
		cca, err := makeAppContextForCompositeApp(project_name, app_name+"-d", "1.0", "1.0", "di", "default", "0")
		if err != nil {
			log.Println(err)
			return "", err
		}

		context := cca.context                    // appcontext.AppContext
		ctxval := cca.ctxval                      // id
		compositeHandle := cca.compositeAppHandle // cid

		var appOrderInstr struct {
			Apporder []string `json:"apporder"`
		}
		var appDepInstr struct {
			Appdep map[string]string `json:"appdependency"`
		}
		appdep := make(map[string]string)
		// create a com_app for each device
		for device, res := range d.qryResmap {
			// Add application
			app_name := device.GetMetadata().Name + "-query-app"
			appOrderInstr.Apporder = append(appOrderInstr.Apporder, app_name)
			appdep[app_name] = "go"

			apphandle, _ := context.AddApp(compositeHandle, app_name)

			// Add cluster
			clusterhandle, _ := context.AddCluster(apphandle, provider_name+"+"+device.GetMetadata().Name)
			err = addQueryResourcesToCluster(context, clusterhandle, &res.Resources)
		}

		jappOrderInstr, _ := json.Marshal(appOrderInstr)
		appDepInstr.Appdep = appdep
		jappDepInstr, _ := json.Marshal(appDepInstr)
		context.AddInstruction(compositeHandle, "app", "order", string(jappOrderInstr))
		context.AddInstruction(compositeHandle, "app", "dependency", string(jappDepInstr))

		// invoke query process
		appContextID := fmt.Sprintf("%v", ctxval)
		d.qryCtxId = appContextID
	}

	err := rsyncclient.InvokeGetResource(d.qryCtxId)
	if err != nil {
		log.Println(err)
		d.qryCtxId = ""
		return "", err
	}

	return d.qryCtxId, nil
}

func (d *ResUtil) GetResourceData(device module.ControllerObject, ns string, name string) (string, error) {
	if d.qryCtxId == "" {
		return "", pkgerrors.New("Query failed to be executed.")
	}

	var ac appcontext.AppContext
	ah, err := ac.LoadAppContext(d.qryCtxId)
	if err != nil {
		return "", pkgerrors.Wrap(err, "AppContext is not found.")
	}

	// wait for resource ready
	err = wait.PollImmediate(time.Second, time.Second*20,
		func() (bool, error) {
			sh, err := ac.GetLevelHandle(ah, "status")
			if err != nil {
				log.Println("Waiting for Resource status to be ready.")
				return false, nil
			}

			s, err := ac.GetValue(sh)
			if err != nil {
				log.Println("Waiting for Resource status to be ready..")
				return false, nil
			}

			acStatus := appcontext.AppContextStatus{}
			js, _ := json.Marshal(s)
			json.Unmarshal(js, &acStatus)
			log.Println(acStatus.Status)
			if acStatus.Status == appcontext.AppContextStatusEnum.Instantiated {
				return true, nil
			}

			log.Println("Waiting for Resource status to be ready...")
			return false, nil
		},
	)

	if err != nil {
		log.Println(err)
		return "", pkgerrors.Wrap(err, "Resource is not available")
	}

	// found resource handle
	if d.qryResmap[device] == nil {
		return "", pkgerrors.Wrap(err, "No query resource found.")
	}

	for _, resource := range d.qryResmap[device].Resources {
		if ns == resource.Resource.Namespace && name == resource.Resource.Name {
			rdh, _ := ac.GetLevelHandle(resource.Handle, "definition")
			if rdh != nil {
				ret, err := ac.GetValue(rdh)
				if err == nil {
					return ret.(string), nil
				}
			}
			return "", pkgerrors.Wrap(err, "Failed to query the resource value.")
		}
	}

	return "", pkgerrors.New("No query resource found.")
}
