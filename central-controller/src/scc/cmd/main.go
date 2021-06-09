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

package main

import (
    "context"
    "log"
    "math/rand"
    "net/http"
    "os"
    "os/signal"
    "time"
    "strconv"

    logs "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/logutils"
    "github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/api"
    "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/auth"
    "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/config"
    "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
    "github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/manager"
    contextDb "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/contextdb"
    controller "github.com/open-ness/EMCO/src/orchestrator/pkg/module/controller"
    "github.com/gorilla/handlers"
    mtypes "github.com/open-ness/EMCO/src/orchestrator/pkg/module/types"

    rconfig "github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/infra/config"
)

const default_rsync_name = "rsync"
const ENV_RSYNC_NAME = "RSYNC_NAME"

func main() {

    rand.Seed(time.Now().UnixNano())

    // create database and context database
    err := db.InitializeDatabaseConnection("scc")
    if err != nil {
            log.Println("Unable to initialize database connection...")
            log.Println(err)
            log.Fatalln("Exiting...")
    }

    err = contextDb.InitializeContextDatabase()
    if err != nil {
            log.Println("Unable to initialize database connection...")
            log.Println(err)
            log.Fatalln("Exiting...")
    }

    // create sdewan namespace and root certificate
    cu, err := manager.GetCertUtil()
    if err == nil {
        _, err = cu.CreateNamespace(manager.NameSpaceName)
        if err == nil {
            log.Println("Namespace is available : " + manager.NameSpaceName)
            _, err = cu.CreateSelfSignedIssuer(manager.RootIssuerName, manager.NameSpaceName)
            if err == nil {
                log.Println("SDEWAN root issuer is available : " + manager.RootIssuerName)
                _, err = cu.CreateCertificate(manager.RootCertName, manager.NameSpaceName, manager.RootIssuerName, true)
                if err == nil {
                    log.Println("SDEWAN root certificate is available : " + manager.RootCertName)
                    _, err = cu.CreateCAIssuer(manager.RootCAIssuerName, manager.NameSpaceName, manager.RootCertName)
                    if err == nil {
                        log.Println("SDEWAN root ca issuer is available : " + manager.RootCAIssuerName)
                    }
                    _, err = cu.CreateCertificate(manager.SCCCertName, manager.NameSpaceName, manager.RootCAIssuerName, false)
                    if err == nil {
                        log.Println("SDEWAN central controller base certificates is available : " + manager.SCCCertName)
                    }
                }
            }
        }
    }

    if err != nil {
        log.Println(err)
    }

    //Register rsync client
    serviceName := os.Getenv(ENV_RSYNC_NAME)
    if serviceName == "" {
            serviceName = default_rsync_name
            logs.Info("Using default name for rsync service name", logs.Fields{
                        "Name": serviceName,
            })
    }

    client := controller.NewControllerClient()

    // Create or update the controller entry
    rsync_port, _ := strconv.Atoi(rconfig.GetConfiguration().RsyncPort)
    controller := controller.Controller{
            Metadata: mtypes.Metadata{
                    Name: serviceName,
            },
            Spec: controller.ControllerSpec{
                    Host:     rconfig.GetConfiguration().RsyncIP,
                    Port:     rsync_port,
                    Type:     controller.CONTROLLER_TYPE_ACTION,
                    Priority: controller.MinControllerPriority,
            },
    }
    _, err = client.CreateController(controller, true)
    if err != nil {
            logs.Error("Failed to create/update a gRPC controller", logs.Fields{
                    "Error":      err,
                    "Controller": serviceName,
            })
    }

    // create http server
    httpRouter := api.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
    loggedRouter := handlers.LoggingHandler(os.Stdout, httpRouter)
    log.Println("Starting SDEWAN Central Controller API")

    httpServer := &http.Server{
            Handler:    loggedRouter,
            Addr:       ":" + config.GetConfiguration().ServicePort,
    }

    connectionsClose := make(chan struct{})
    go func() {
            c := make(chan os.Signal, 1)
            signal.Notify(c, os.Interrupt)
            <-c
            httpServer.Shutdown(context.Background())
            close(connectionsClose)
    }()

    tlsConfig, err := auth.GetTLSConfig("ca.cert", "server.cert", "server.key")
    if err != nil {
        log.Println("Error Getting TLS Configuration. Starting without TLS...")
        log.Fatal(httpServer.ListenAndServe())
    } else {
            httpServer.TLSConfig = tlsConfig

            err = httpServer.ListenAndServeTLS("", "")
    }

}
