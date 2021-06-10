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
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/client"
	pkgerrors "github.com/pkg/errors"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	"os"
	"sigs.k8s.io/yaml"
)

func DecodeYAMLFromFile(path string, into runtime.Object) (runtime.Object, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, pkgerrors.New("File " + path + " not found")
		} else {
			return nil, pkgerrors.Wrap(err, "Stat file error")
		}
	}

	rawBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Read YAML file error")
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(rawBytes, nil, into)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
}

func DecodeYAMLFromData(data []byte, into runtime.Object) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, into)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize YAML error")
	}

	return obj, nil
}

type KubeConfigUtil struct {
}

var kubeutil = KubeConfigUtil{}

func GetKubeConfigUtil() *KubeConfigUtil {
	return &kubeutil
}

func (c *KubeConfigUtil) toYaml(data *unstructured.Unstructured) ([]byte, error) {
	byte_json, err := data.MarshalJSON()
	if err != nil {
		return []byte(""), pkgerrors.Wrap(err, "Fail to generate yaml")
	}

	byte_yaml, err := yaml.JSONToYAML(byte_json)
	if err != nil {
		return []byte(""), pkgerrors.Wrap(err, "Fail to generate yaml")
	}

	return byte_yaml, nil
}

func (c *KubeConfigUtil) UpdateK8sConfig(conf []byte, server string, insecure bool) ([]byte, error) {
	conf_us_obj := &unstructured.Unstructured{}
	_, err := DecodeYAMLFromData(conf, conf_us_obj)
	if err == nil {
		conf_obj := conf_us_obj.UnstructuredContent()
		cluster_objs, _, err := unstructured.NestedSlice(conf_obj, "clusters")
		if err == nil {
			if len(cluster_objs) > 0 {
				cluster_obj := cluster_objs[0].(map[string]interface{})
				if insecure {
					// remove certificate-authority-data
					unstructured.RemoveNestedField(cluster_obj, "cluster", "certificate-authority-data")
					// add insecure-skip-tls-verify
					err = unstructured.SetNestedField(cluster_obj, true, "cluster", "insecure-skip-tls-verify")
				}

				if err == nil {
					// set server
					err = unstructured.SetNestedField(cluster_obj, server, "cluster", "server")
					if err == nil {
						err = unstructured.SetNestedSlice(conf_obj, cluster_objs, "clusters")
						if err == nil {
							return c.toYaml(conf_us_obj)
						}
					}
				}
			} else {
				return []byte(""), pkgerrors.New("UpdateK8sConfig: No cluster")
			}
		}
	}

	return []byte(""), pkgerrors.Wrap(err, "UpdateK8sConfig")
}

func (c *KubeConfigUtil) checkKubeConfigAvail(conf []byte, ips []string, port string) ([]byte, string, error) {
	kubeclient := client.NewClient("", "", conf)
	for i := 0; i < len(ips); i++ {
		ip := ips[i]
		//UpdateConfig
		new_url := "https://" + ips[i] + ":" + port
		conf, err := kubeutil.UpdateK8sConfig(conf, new_url, true)
		if err != nil {
			log.Println(err)
			return []byte(""), "", pkgerrors.New("Error in updating kubeconfig")
		}
		kubeclient = client.NewClient("", "", []byte(conf))
		is_reachable := kubeclient.IsReachable()
		if is_reachable == true {
			return conf, ip, nil
		}
	}
	return []byte(""), "", pkgerrors.New("No public ip found workable for the cluster")
}
