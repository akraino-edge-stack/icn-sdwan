/*

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

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/wrtprovider"
)

var OpenwrtTag = "hle2/openwrt-1806-mwan3:latest"

func init() {
	if img := os.Getenv("OPENWRT_IMAGE"); img != "" {
		OpenwrtTag = img
	}
}

// SdewanReconciler reconciles a Sdewan object
type SdewanReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=sdewans,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=sdewans/status,verbs=get;update;patch

func (r *SdewanReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("sdewan", req.NamespacedName)

	// your logic here
	// Fetch the Sdewan instance
	instance := &batchv1alpha1.Sdewan{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}
	for i, network := range instance.Spec.Networks {
		if network.Interface == "" {
			instance.Spec.Networks[i].Interface = fmt.Sprintf("net%d", i)
		}
	}

	cm := newConfigmapForCR(instance)
	if err := ctrl.SetControllerReference(instance, cm, r.Scheme); err != nil {
		log.Error(err, "Failed to set configmap controller reference", "Configmap.Name", cm.Name)
		return ctrl.Result{}, nil
	}
	foundcm := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, foundcm)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Configmap", "Configmap.Namespace", cm.Namespace, "Configmap.Name", cm.Name)
		err = r.Create(ctx, cm)
		if err != nil {
			log.Error(err, "Failed to create configmap", "Configmap.Name", cm.Name)
			return ctrl.Result{}, nil
		}
	} else if err != nil {
		log.Error(err, "Error happends when fetch the configmap", "Configmap.Name", cm.Name)
		return ctrl.Result{}, nil
	} else if reflect.DeepEqual(foundcm.Data, cm.Data) {
		log.Info("Updating Configmap", "Configmap.Namespace", cm.Namespace, "Configmap.Name", cm.Name)
		err = r.Update(ctx, cm)
		if err != nil {
			log.Error(err, "Failed to update configmap", "Configmap.Name", cm.Name)
			return ctrl.Result{}, nil
		}
	} else {
		log.Info("Configmap not changed", "Configmap.Namespace", foundcm.Namespace, "Configmap.Name", foundcm.Name)
	}
	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set Sdewan instance as the owner and controller
	if err := ctrl.SetControllerReference(instance, pod, r.Scheme); err != nil {
		return ctrl.Result{}, nil
	}

	// Check if this Pod already exists
	foundpod := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, foundpod)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.Create(ctx, pod)
		if err != nil {
			log.Error(err, "Failed to create the new pod", "Name", pod.Name)
			return ctrl.Result{}, nil
		}

		// Pod created successfully - don't requeue
		log.Info("A new Pod created", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	} else if err != nil {
		return ctrl.Result{}, nil
	} else {
		// Pod already exists - don't requeue
		log.Info("Pod already exists", "Pod.Namespace", foundpod.Namespace, "Pod.Name", foundpod.Name)
	}

	svc := newSvcForCR(instance)
	// Set Sdewan instance as the owner and controller
	if err := ctrl.SetControllerReference(instance, svc, r.Scheme); err != nil {
		return ctrl.Result{}, nil
	}
	// Check if this svc already exists
	foundsvc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundsvc)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			return ctrl.Result{}, nil
		}
		log.Info("A new Service created", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
	} else if err != nil {
		return ctrl.Result{}, nil
	} else {
		log.Info("Service already exists", "Service.Namespace", foundsvc.Namespace, "Service.Name", foundsvc.Name)
	}

	// Apply rules if the pod is ready
	if len(foundpod.Status.ContainerStatuses) > 0 && foundpod.Status.ContainerStatuses[0].Ready {
		mwan3conf := &batchv1alpha1.Mwan3Conf{}
		err = r.Get(ctx, types.NamespacedName{Name: instance.Spec.Mwan3Conf, Namespace: instance.Namespace}, mwan3conf)
		if err != nil {
			log.Error(err, "unable to find the mwan3conf", "namespace", instance.Namespace, "mwan3 name", instance.Spec.Mwan3Conf)
			instance.Status.Mwan3Status = batchv1alpha1.Mwan3Status{Name: instance.Spec.Mwan3Conf, IsApplied: false}
			err = r.Status().Update(ctx, instance)
			if err != nil {
				log.Error(err, "Failed to update Sdewan status")
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, nil
		}
		if (instance.Status.Mwan3Status.Name != instance.Spec.Mwan3Conf) || !instance.Status.Mwan3Status.IsApplied {
			err = wrtprovider.Mwan3Apply(mwan3conf, instance)
			if err != nil {
				log.Error(err, "Failed to apply the mwan3conf", "namespace", instance.Namespace, "mwan3 name", instance.Spec.Mwan3Conf)
				instance.Status.Mwan3Status = batchv1alpha1.Mwan3Status{Name: instance.Spec.Mwan3Conf, IsApplied: false}
				r.Status().Update(ctx, instance)
				return ctrl.Result{}, nil
			}
			instance.Status.Mwan3Status = batchv1alpha1.Mwan3Status{Name: instance.Spec.Mwan3Conf, IsApplied: true, AppliedTime: &metav1.Time{Time: time.Now()}}
			err = r.Status().Update(ctx, instance)
			if err != nil {
				log.Error(err, "Failed to update Sdewan status")
				return ctrl.Result{}, nil
			}
			log.Info("sdewan config applied")
		} else {
			log.Info("mwan3 conf not chnaged, so not re-apply", "sdewan", instance.Name)
		}
	} else {
		log.Info("Don't apply conf as the pod is not ready", "sdewan", instance.Name)
	}

	return ctrl.Result{}, nil
}

func (r *SdewanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.Sdewan{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}

// newSvcForCR returns a svc with the same name/namespace as the cr
func newSvcForCR(cr *batchv1alpha1.Sdewan) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "api",
					Port: 80,
				},
			},
			Selector: map[string]string{
				"app": cr.Name,
			},
		},
	}
}

// newConfigmapForCR returns a configmap with the same name/namespace as the cr
func newConfigmapForCR(cr *batchv1alpha1.Sdewan) *corev1.ConfigMap {
	netjson, _ := json.MarshalIndent(cr.Spec.Networks, "", "  ")
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Data: map[string]string{
			"networks.json": string(netjson),
			"entrypoint.sh": `#!/bin/bash
# Always exit on errors.
set -e
echo "" > /etc/config/network
cat > /etc/config/mwan3 <<EOF
config globals 'globals'
    option mmx_mask '0x3F00'
    option local_source 'lan'
EOF
for net in $(jq -c ".[]" /tmp/sdewan/networks.json)
do
  interface=$(echo $net | jq -r .interface)
  ipaddr=$(ifconfig $interface | awk '/inet/{print $2}' | cut -f2 -d ":" | awk 'NR==1 {print $1}')
  if [ "$isProvider" == "true" ] || [ "$isProvider" == "1" ]]
  then
    vif="wan_$interface"
  else
    vif="lan_$interface"
  fi
  netmask=$(ifconfig $interface | awk '/inet/{print $4}' | cut -f2 -d ":" | awk 'NR==1 {print $1}')
  cat >> /etc/config/network <<EOF
config interface '$vif'
    option ifname '$interface'
    option proto 'static'
    option ipaddr '$ipaddr'
    option netmask '$netmask'
EOF
  cat >> /etc/config/mwan3 <<EOF
config interface '$vif'
        option enabled '1'
        option family 'ipv4'
        option reliability '2'
        option count '1'
        option timeout '2'
        option failure_latency '1000'
        option recovery_latency '500'
        option failure_loss '20'
        option recovery_loss '5'
        option interval '5'
        option down '3'
        option up '8'
EOF
done
/sbin/procd &
/sbin/ubusd &
iptables -S
sleep 1
/etc/init.d/rpcd start
/etc/init.d/dnsmasq start
/etc/init.d/network start
/etc/init.d/odhcpd start
/etc/init.d/uhttpd start
/etc/init.d/log start
/etc/init.d/dropbear start
/etc/init.d/mwan3 restart
echo "Entering sleep... (success)"
# Sleep forever.
while true; do sleep 100; done`,
		},
	}
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *batchv1alpha1.Sdewan) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	priv := true
	var netmaps []map[string]interface{}
	for _, net := range cr.Spec.Networks {
		netmaps = append(netmaps, map[string]interface{}{
			"name":           net.Name,
			"interface":      net.Interface,
			"defaultGateway": fmt.Sprintf("%t", net.DefaultGateway),
		})
	}
	netjson, _ := json.MarshalIndent(netmaps, "", "  ")
	volumes := []corev1.Volume{
		{
			Name: cr.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: cr.Name},
				},
			},
		},
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"k8s.v1.cni.cncf.io/networks":      `[{ "name": "ovn-networkobj"}]`,
				"k8s.plugin.opnfv.org/nfn-network": `{ "type": "ovn4nfv", "interface": ` + string(netjson) + "}",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "sdewan",
					Image:           OpenwrtTag,
					Command:         []string{"/bin/sh", "/tmp/sdewan/entrypoint.sh"},
					ImagePullPolicy: corev1.PullIfNotPresent,
					ReadinessProbe: &corev1.Probe{
						Handler: corev1.Handler{HTTPGet: &corev1.HTTPGetAction{
							Path:   "/",
							Port:   intstr.IntOrString{IntVal: 80},
							Scheme: corev1.URISchemeHTTP},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       5,
						FailureThreshold:    5,
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: &priv,
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      cr.Name,
							ReadOnly:  true,
							MountPath: "/tmp/sdewan",
						},
					},
				},
			},
			NodeSelector: map[string]string{"kubernetes.io/hostname": cr.Spec.Node},
			Volumes:      volumes,
		},
	}
}
