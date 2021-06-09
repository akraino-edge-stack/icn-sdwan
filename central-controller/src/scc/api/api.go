/*
Copyright 2020 Intel Corporation.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
    "github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/manager"
    "github.com/gorilla/mux"
)

// NewRouter creates a router that registers the various urls that are
// supported

func createHandlerMapping(
    objectClient manager.ControllerObjectManager,
    router *mux.Router,
    collections string,
    resource string ) {
    objectHandler := ControllerHandler{client: objectClient}
    if objectClient.IsOperationSupported("POST") == true {
        router.HandleFunc(
            "/" + collections,
            objectHandler.createHandler).Methods("POST")
    }

    if objectClient.IsOperationSupported("GETS") == true {
        router.HandleFunc(
            "/" + collections,
            objectHandler.getsHandler).Methods("GET")
    }

    if objectClient.IsOperationSupported("GET") == true {
        router.HandleFunc(
            "/" + collections + "/{" + resource + "}",
            objectHandler.getHandler).Methods("GET")
    }

    if objectClient.IsOperationSupported("DELETE") == true {
        router.HandleFunc(
            "/" + collections + "/{" + resource + "}",
            objectHandler.deleteHandler).Methods("DELETE")
    }

    if objectClient.IsOperationSupported("PUT") == true {
        router.HandleFunc(
            "/" + collections + "/{" + resource + "}",
            objectHandler.updateHandler).Methods("PUT")
    }
}

func NewRouter(
    overlayObjectClient manager.ControllerObjectManager,
    proposalObjectClient manager.ControllerObjectManager,
    hubObjectClient manager.ControllerObjectManager,
    hubConnObjectClient manager.ControllerObjectManager,
    hubDeviceObjectClient manager.ControllerObjectManager,
    hubCNFObjectClient manager.ControllerObjectManager,
    deviceObjectClient manager.ControllerObjectManager,
    deviceConnObjectClient manager.ControllerObjectManager,
    deviceCNFObjectClient manager.ControllerObjectManager,
    ipRangeObjectClient manager.ControllerObjectManager,
    providerIpRangeObjectClient manager.ControllerObjectManager,
    certificateObjectClient manager.ControllerObjectManager) *mux.Router {

    router := mux.NewRouter()
    ver := "v1"
    mgrset := manager.GetManagerset()

    // router
    verRouter := router.PathPrefix("/scc/" + ver).Subrouter()
    providerRouter := router.PathPrefix("/scc/" + ver + "/provider").Subrouter()
    olRouter := verRouter.PathPrefix("/" + manager.OverlayCollection + "/{" + manager.OverlayResource + "}").Subrouter()
    hubRouter := olRouter.PathPrefix("/" + manager.HubCollection + "/{" + manager.HubResource + "}").Subrouter()
    devRouter := olRouter.PathPrefix("/" + manager.DeviceCollection + "/{" + manager.DeviceResource + "}").Subrouter()

    // overlay API
    if overlayObjectClient == nil {
         overlayObjectClient = manager.NewOverlayObjectManager()
    }
    mgrset.Overlay = overlayObjectClient.(*manager.OverlayObjectManager)
    createHandlerMapping(overlayObjectClient, verRouter, manager.OverlayCollection, manager.OverlayResource)

    // proposal API
    if proposalObjectClient == nil {
         proposalObjectClient = manager.NewProposalObjectManager()
    }
    mgrset.Proposal = proposalObjectClient.(*manager.ProposalObjectManager)
    createHandlerMapping(proposalObjectClient, olRouter, manager.ProposalCollection, manager.ProposalResource)

    // hub API
    if hubObjectClient == nil {
         hubObjectClient = manager.NewHubObjectManager()
    }
    mgrset.Hub = hubObjectClient.(*manager.HubObjectManager)
    createHandlerMapping(hubObjectClient, olRouter, manager.HubCollection, manager.HubResource)

    // hub-connection API
    if hubConnObjectClient == nil {
         hubConnObjectClient = manager.NewHubConnObjectManager()
    }
    mgrset.HubConn = hubConnObjectClient.(*manager.HubConnObjectManager)
    createHandlerMapping(hubConnObjectClient, hubRouter, manager.ConnectionCollection, manager.ConnectionResource)

    // hub-cnf API
    if hubCNFObjectClient == nil {
         hubCNFObjectClient = manager.NewCNFObjectManager(true)
    }
    mgrset.HubCNF = hubCNFObjectClient.(*manager.CNFObjectManager)
    createHandlerMapping(hubCNFObjectClient, hubRouter, manager.CNFCollection, manager.CNFResource)

    // hub-device API
    if hubDeviceObjectClient == nil {
         hubDeviceObjectClient = manager.NewHubDeviceObjectManager()
    }
    mgrset.HubDevice = hubDeviceObjectClient.(*manager.HubDeviceObjectManager)
    createHandlerMapping(hubDeviceObjectClient, hubRouter, manager.DeviceCollection, manager.DeviceResource)

    // device API
    if deviceObjectClient == nil {
         deviceObjectClient = manager.NewDeviceObjectManager()
    }
    mgrset.Device = deviceObjectClient.(*manager.DeviceObjectManager)
    createHandlerMapping(deviceObjectClient, olRouter, manager.DeviceCollection, manager.DeviceResource)

    // device-connection API
    if deviceConnObjectClient == nil {
         deviceConnObjectClient = manager.NewDeviceConnObjectManager()
    }
    mgrset.DeviceConn = deviceConnObjectClient.(*manager.DeviceConnObjectManager)
    createHandlerMapping(deviceConnObjectClient, devRouter, manager.ConnectionCollection, manager.ConnectionResource)

    // device-cnf API
    if deviceCNFObjectClient == nil {
         deviceCNFObjectClient = manager.NewCNFObjectManager(false)
    }
    mgrset.DeviceCNF = deviceCNFObjectClient.(*manager.CNFObjectManager)
    createHandlerMapping(deviceCNFObjectClient, devRouter, manager.CNFCollection, manager.CNFResource)

    // provider iprange API
    if providerIpRangeObjectClient == nil {
         providerIpRangeObjectClient = manager.NewIPRangeObjectManager(true)
    }
    mgrset.ProviderIPRange = providerIpRangeObjectClient.(*manager.IPRangeObjectManager)
    createHandlerMapping(providerIpRangeObjectClient, providerRouter, manager.IPRangeCollection, manager.IPRangeResource)

    // iprange API
    if ipRangeObjectClient == nil {
         ipRangeObjectClient = manager.NewIPRangeObjectManager(false)
    }
    mgrset.IPRange = ipRangeObjectClient.(*manager.IPRangeObjectManager)
    createHandlerMapping(ipRangeObjectClient, olRouter, manager.IPRangeCollection, manager.IPRangeResource)

    // certificate API
    if certificateObjectClient == nil {
         certificateObjectClient = manager.NewCertificateObjectManager()
    }
    mgrset.Cert = certificateObjectClient.(*manager.CertificateObjectManager)
    createHandlerMapping(certificateObjectClient, olRouter, manager.CertCollection, manager.CertResource)

    // Add depedency
    overlayObjectClient.AddOwnResManager(proposalObjectClient)
    overlayObjectClient.AddOwnResManager(hubObjectClient)
    overlayObjectClient.AddOwnResManager(deviceObjectClient)
    overlayObjectClient.AddOwnResManager(ipRangeObjectClient)
    overlayObjectClient.AddOwnResManager(certificateObjectClient)
    hubObjectClient.AddOwnResManager(hubDeviceObjectClient)
    deviceObjectClient.AddOwnResManager(hubDeviceObjectClient)

    proposalObjectClient.AddDepResManager(overlayObjectClient)
    hubObjectClient.AddDepResManager(overlayObjectClient)
    deviceObjectClient.AddDepResManager(overlayObjectClient)
    ipRangeObjectClient.AddDepResManager(overlayObjectClient)
    certificateObjectClient.AddDepResManager(overlayObjectClient)
    hubDeviceObjectClient.AddDepResManager(hubObjectClient)
    hubDeviceObjectClient.AddDepResManager(deviceObjectClient)
    hubConnObjectClient.AddDepResManager(hubObjectClient)
    deviceConnObjectClient.AddDepResManager(deviceObjectClient)
    hubCNFObjectClient.AddDepResManager(hubObjectClient)
//    deviceCNFObjectClient.AddDepResManager(deviceObjectClient)

    return router
}
