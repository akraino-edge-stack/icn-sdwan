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
    "io"
    "encoding/json"
    "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
    "github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
    pkgerrors "github.com/pkg/errors"

    "k8s.io/apimachinery/pkg/runtime/schema"
    "log"
)

type CNFObjectKey struct {
    OverlayName string `json:"overlay-name"`
    ClusterName string `json:"cluster-name"`
    CNFName string `json:"cnf-name"`
}

// CNFObjectManager implements the ControllerObjectManager
type CNFObjectManager struct {
    BaseObjectManager
    isHub bool
}

func NewCNFObjectManager(isHub bool) *CNFObjectManager {
    object_meta := "cnf"
    if isHub {
        object_meta = "hub-" + object_meta
    } else {
        object_meta = "device-" + object_meta
    }

    return &CNFObjectManager{
        BaseObjectManager {
            storeName:  StoreName,
            tagMeta:    object_meta,
            depResManagers: []ControllerObjectManager {},
            ownResManagers: []ControllerObjectManager {},
        },
        isHub,
    }
}

func (c *CNFObjectManager) GetResourceName() string {
    return CNFResource
}

func (c *CNFObjectManager) IsOperationSupported(oper string) bool {
    if oper == "GETS" {
        return true
    }
    return false
}

func (c *CNFObjectManager) CreateEmptyObject() module.ControllerObject {
    return &module.CNFObject{}
}

func (c *CNFObjectManager) GetStoreKey(m map[string]string, t module.ControllerObject, isCollection bool) (db.Key, error) {
    overlay_name := m[OverlayResource]
    cluster_name := ""
    if c.isHub {
        cluster_name = m[HubResource]
    } else {
        cluster_name = m[DeviceResource]
    }

    key := CNFObjectKey{
        OverlayName: overlay_name,
        ClusterName: cluster_name,
        CNFName: "",
    }

    if isCollection == true {
        return key, nil
    }

    to := t.(*module.CNFObject)
    meta_name := to.Metadata.Name
    res_name := m[CNFResource]

    if res_name != "" {
        if meta_name != "" && res_name != meta_name {
            return key, pkgerrors.New("Resource name unmatched metadata name")
        } 

        key.CNFName = res_name
    } else {
        if meta_name == "" {
            return key, pkgerrors.New("Unable to find resource name")  
        }

        key.CNFName = meta_name
    }

    return key, nil;
}

func (c *CNFObjectManager) ParseObject(r io.Reader) (module.ControllerObject, error) {
    var v module.CNFObject
    err := json.NewDecoder(r).Decode(&v)

    return &v, err
}

func (c *CNFObjectManager) CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
    return c.CreateEmptyObject(), pkgerrors.New("Not implemented")
}

func (c *CNFObjectManager) GetObject(m map[string]string) (module.ControllerObject, error) {
    return c.CreateEmptyObject(), pkgerrors.New("Not implemented")
}

func (c *CNFObjectManager) GetObjects(m map[string]string) ([]module.ControllerObject, error) {
    overlay_name := m[OverlayResource]
    var cobj module.ControllerObject
    var cluster_name string
    var err error

    if c.isHub {
        cluster_name = m[HubResource]
	hub_name := m[HubResource]
        hub_manager := GetManagerset().Hub
        cobj, err = hub_manager.GetObject(m)
        if err != nil {
            return []module.ControllerObject{}, pkgerrors.Wrap(err, "Hub " + hub_name + " is not defined")
        }
    } else {
        cluster_name = m[DeviceResource]
	device_name := m[DeviceResource]
        dev_manager := GetManagerset().Device
        cobj, err = dev_manager.GetObject(m)
        if err != nil {
            //return c.CreateEmptyObject(), pkgerrors.Wrap(err, "Device " + device_name + " is not defined")
	    return []module.ControllerObject{}, pkgerrors.Wrap(err, "Device " + device_name + " is not defined")
        }
    }

    // Query CNFStatus
    resutil := NewResUtil()

//    cobj = &module.DeviceObject{
  //              Metadata: module.ObjectMetaData{"local", "", "", ""}, 
    //            Specification: module.DeviceObjectSpec{}}

    res := QueryResource{
                Resource: ReadResource {
                    Gvk: schema.GroupVersionKind{Group: "batch.sdewan.akraino.org", Version: "v1alpha1", Kind:"CNFStatus"},
                    Name: "cnf-status",
                    Namespace: "sdewan-system",
                }}
//    res := QueryResource{
//                Resource: ReadResource {
//                    Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind:"ConfigMap"},
//                    Name: "mycm",
//                    Namespace: "default",
//                }}
    resutil.AddQueryResource(cobj, res)
    ctx_id, err := resutil.Query("ewo-query-app")

    if err != nil {
        log.Println(err)
        return []module.ControllerObject{}, pkgerrors.Wrap(err, "Failed to Query CNFs")
    }

    // Todo: save ctx_id in DB
    log.Println(ctx_id)
    // val, err := resutil.GetResourceData(&deviceObject, "default", "mycm")
    val, err := resutil.GetResourceData(cobj, "sdewan-system", "cnf-status")
    if err != nil {
        log.Println(err)
        return []module.ControllerObject{}, pkgerrors.Wrap(err, "CNF information is not available")
    }

    log.Println(val)
    status, err := c.ParseStatus(val)
    if err != nil {
        log.Println(err)
        return []module.ControllerObject{}, pkgerrors.Wrap(err, "CNF information is not available")
    }

    return []module.ControllerObject{&module.CNFObject{
        Metadata: module.ObjectMetaData{overlay_name+"."+cluster_name, "cnf informaiton", "", ""},
        Status: status,
    }}, nil
}

func (c *CNFObjectManager) ParseStatus(val string) (string, error) {
/*    val = `{"apiVersion":"batch.sdewan.akraino.org/v1alpha1",
"kind":"CNFStatus",
"metadata":{"creationTimestamp":"2021-04-06T04:51:18Z","generation":1,"managedFields":[{"apiVersion":"batch.sdewan.akraino.org/v1alpha1","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{},"f:status":{".":{},"f:appliedGeneration":{},"f:appliedTime":{},"f:information":{}}},"manager":"manager","operation":"Update","time":"2021-05-07T09:42:18Z"}],"name":"cnf-status","namespace":"sdewan-system","resourceVersion":"29161589","uid":"940de684-ca24-4686-9f3f-2bad0c099bc2"},
"spec":{},
"status":{"appliedGeneration":1,"appliedTime":"2021-05-28T06:43:48Z","information":[{"ip":"10.233.97.182","name":"sdewan-test-7bb68cd886-b7ntl","namespace":"default","node":"master","purpose":"sdewan-test","status":"Not Available"},{"ip":"10.233.97.155","name":"sdewan-raw-cnf-9fdcd875c-qkvhk","namespace":"default","node":"master","purpose":"sdewan-raw-cnf","status":"Not Available"},{"ip":"10.233.97.158","name":"sdewan-safe-6555988556-fhmvw","namespace":"default","node":"master","purpose":"sdewan-safe","status":"Not Available"},{"ip":"10.233.97.144","name":"sdewan-edge-a-7fb6b8bbf6-lvs5d","namespace":"default","node":"master","purpose":"sdewan-edge-a","status":"Not Available"}]}}`
*/
    var vi interface{}
    err := json.Unmarshal([]byte(val), &vi)
    if err != nil {
        return "", err
    }

    status := vi.(map[string]interface{})["status"]
    status_val, err := json.Marshal(status)

    //status_val, err := status
    if err != nil {
        return "", err
    }

    log.Println("_____________-----")
    log.Println(string(status_val))
    return string(status_val), nil
/*
    status := vi.(map[string]interface{})["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})["kubectl.kubernetes.io/last-applied-configuration"]
    return status.(string), nil
    */
}

func (c *CNFObjectManager) UpdateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
    return c.CreateEmptyObject(), pkgerrors.New("Not implemented")
}

func (c *CNFObjectManager) DeleteObject(m map[string]string) error {
    return pkgerrors.New("Not implemented")
}
