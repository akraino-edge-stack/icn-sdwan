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
// Based on Code: https://github.com/johandry/klient

package client

import (
    "log"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/kubernetes/scheme"
    "k8s.io/client-go/kubernetes"
    corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
    certmanagerversioned "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
    certmanagerv1beta1 "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1beta1"
)

type KubernetesClient struct {
    Context string
    ConfigPath string
    KubeConfig  []byte
}

func NewClient(context string, configPath string, kubeConfig []byte) *KubernetesClient {
    return &KubernetesClient{
        Context: context,
        ConfigPath: configPath,
        KubeConfig: kubeConfig,
    }
}

func (c *KubernetesClient) ToRESTConfig() (*rest.Config, error) {
    var config *rest.Config
    var err error
    if len(c.KubeConfig) == 0 {
        // From: k8s.io/kubectl/pkg/cmd/util/kubectl_match_version.go > func setKubernetesDefaults()
        config, err = c.toRawKubeConfigLoader().ClientConfig()
    } else {
        config, err = clientcmd.RESTConfigFromKubeConfig(c.KubeConfig)
    }

    if err != nil {
        return nil, err
    }

    if config.GroupVersion == nil {
        config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
    }
    if config.APIPath == "" {
        config.APIPath = "/api"
    }
    if config.NegotiatedSerializer == nil {
        // This codec config ensures the resources are not converted. Therefore, resources
        // will not be round-tripped through internal versions. Defaulting does not happen
        // on the client.
        config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
    }

    rest.SetKubernetesDefaults(config)
    return config, nil
}

// toRawKubeConfigLoader creates a client using the following rules:
// 1. builds from the given kubeconfig path, if not empty
// 2. use the in cluster factory if running in-cluster
// 3. gets the factory from KUBECONFIG env var
// 4. Uses $HOME/.kube/factory
// It's required to implement the interface genericclioptions.RESTClientGetter
func (c *KubernetesClient) toRawKubeConfigLoader() clientcmd.ClientConfig {
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
    if len(c.ConfigPath) != 0 {
        loadingRules.ExplicitPath = c.ConfigPath
    }
    configOverrides := &clientcmd.ConfigOverrides{
        ClusterDefaults: clientcmd.ClusterDefaults,
    }
    if len(c.Context) != 0 {
        configOverrides.CurrentContext = c.Context
    }

    return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
}

func (c *KubernetesClient) GetCMClients() (certmanagerv1beta1.CertmanagerV1beta1Interface, corev1.CoreV1Interface, error) {
    config, err := c.ToRESTConfig()
    if err != nil {
        return nil, nil, err
    }

    cmclientset, err := certmanagerversioned.NewForConfig(config)
    if err != nil {
        return nil, nil, err
    }

    k8sclientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, nil, err
    }

    return cmclientset.CertmanagerV1beta1(), k8sclientset.CoreV1(), nil
}

func (c *KubernetesClient) KubernetesClientSet() (*kubernetes.Clientset, error) {
    config, err := c.ToRESTConfig()
    if err != nil {
        return nil, err
    }

    return kubernetes.NewForConfig(config)
}

func (c *KubernetesClient) IsReachable() bool {
    clientset, err := c.KubernetesClientSet()
    if err != nil {
        log.Println(err)
        return false
    }

    _, err = clientset.ServerVersion()
    if err != nil {
        log.Println(err)
        return false
    }

    return true
}

