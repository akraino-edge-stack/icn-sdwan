package main

import (
    "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
    rsync "github.com/open-ness/EMCO/src/rsync/pkg/db"
    "log"
    "math/rand"
    "time"
    "io/ioutil"
    "encoding/base64"
    pkgerrors "github.com/pkg/errors"
)

func registerCluster(provider_name string, cluster_name string, kubeconfig_file string) error {
    content, err := ioutil.ReadFile(kubeconfig_file)

    ccc := rsync.NewCloudConfigClient()

    _, err = ccc.CreateCloudConfig(provider_name, cluster_name, "0", "default", base64.StdEncoding.EncodeToString(content))
    if err != nil {
        return pkgerrors.Wrap(err, "Error creating cloud config")
    }

    return nil
}

func main() {
    rand.Seed(time.Now().UnixNano())

    // Initialize the mongodb
    err := db.InitializeDatabaseConnection("scc")
    if err != nil {
        log.Println("Unable to initialize database connection...")
        log.Println(err)
        log.Fatalln("Exiting...")
    }

    provider_name := "akraino_scc"
    cluster_name := "local"
    // Register cluster kubeconfig
    registerCluster(provider_name, cluster_name, "admin.conf")    
}
