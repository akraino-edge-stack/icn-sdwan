package main

import (
	"encoding/base64"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	rsync "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/db"
	"k8s.io/client-go/kubernetes"
        "k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgerrors "github.com/pkg/errors"
	"context"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
	"flag"
	"os"
)

const (
	defaultUsername="scc"
	defaultNamespace="sdewan-system"
	defaultKey="userPassword"
)

func registerCluster(provider_name string, cluster_name string, kubeconfig_file string) error {
	content, err := ioutil.ReadFile(kubeconfig_file)

	ccc := rsync.NewCloudConfigClient()

	config, _ := ccc.GetCloudConfig(provider_name, cluster_name, "0", "default")
	if config.Config != "" {
		ccc.DeleteCloudConfig(provider_name, cluster_name, "0", "default")
	}

	_, err = ccc.CreateCloudConfig(provider_name, cluster_name, "0", "default", base64.StdEncoding.EncodeToString(content))
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating cloud config")
	}

	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var mongo_secret, mongo_data_secret, kubeconfig_path *string
	mongo_secret = flag.String("mongoSecret", "mongo-secret", "secret name for mongo access")
	mongo_data_secret = flag.String("mongoDataSecret", "mongo-data-secret", "secret name for mongo encryption")
	kubeconfig_path = flag.String("kubeconfigPath", "admin.conf", "path for kubeconfig")
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig_path)
	if err != nil {
		log.Println("Error in getting kubeconfig")
		log.Println(err)
		log.Fatalln("Exiting...")
	}
	clientset, _ := kubernetes.NewForConfig(cfg)
	mongoSecret, errs := clientset.CoreV1().Secrets(defaultNamespace).Get(context.Background(), *mongo_secret, metav1.GetOptions{})
	mongoDataSecret, errs := clientset.CoreV1().Secrets(defaultNamespace).Get(context.Background(), *mongo_data_secret, metav1.GetOptions{})
	if errs != nil {
		log.Println(errs)
		log.Fatalln("Exiting...")
	}

	os.Setenv("DB_EMCO_USERNAME", defaultUsername)
	os.Setenv("DB_EMCO_PASSWORD", string(mongoSecret.Data[defaultKey]))
	os.Setenv("EMCO_DATA_KEY", string(mongoDataSecret.Data["key"]))

	// Initialize the mongodb
	err = db.InitializeDatabaseConnection("scc")
	if err != nil {
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	provider_name := "akraino_scc"
	cluster_name := "local"
	// Register cluster kubeconfig
	registerCluster(provider_name, cluster_name, *kubeconfig_path)
}
