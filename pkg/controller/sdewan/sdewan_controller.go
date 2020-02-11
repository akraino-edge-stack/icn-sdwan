package sdewan

import (
	"context"
	"fmt"
	"encoding/json"
	"reflect"

	sdewanv1alpha1 "sdewan-operator/pkg/apis/sdewan/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_sdewan")
var OpenwrtTag = "hle2/openwrt-1806-mwan3:latest"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Sdewan Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSdewan{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("sdewan-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Sdewan
	err = c.Watch(&source.Kind{Type: &sdewanv1alpha1.Sdewan{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Sdewan
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &sdewanv1alpha1.Sdewan{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSdewan implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSdewan{}

// ReconcileSdewan reconciles a Sdewan object
type ReconcileSdewan struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Sdewan object and makes changes based on the state read
// and what is in the Sdewan.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSdewan) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Sdewan")

	// Fetch the Sdewan instance
	instance := &sdewanv1alpha1.Sdewan{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	for i, network := range instance.Spec.Networks {
		if network.Interface == "" {
			instance.Spec.Networks[i].Interface = fmt.Sprintf("net%d", i)
		}
	}

	cm := newConfigmapForCR(instance)
	if err := controllerutil.SetControllerReference(instance, cm, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	foundcm := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, foundcm)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Configmap", "Configmap.Namespace", cm.Namespace, "Configmap.Name", cm.Name)
		err = r.client.Create(context.TODO(), cm)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	} else if reflect.DeepEqual(foundcm.Data, cm.Data) {
		reqLogger.Info("Updating Configmap", "Configmap.Namespace", cm.Namespace, "Configmap.Name", cm.Name)
		err = r.client.Update(context.TODO(), cm)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else {
		reqLogger.Info("Configmap not changed", "Configmap.Namespace", foundcm.Namespace, "Configmap.Name", foundcm.Name)
	}
	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set Sdewan instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	foundpod := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, foundpod)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
                reqLogger.Info("A new Pod created", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		// Pod already exists - don't requeue
		reqLogger.Info("Pod already exists", "Pod.Namespace", foundpod.Namespace, "Pod.Name", foundpod.Name)
	}

        svc := newSvcForCR(instance)
	// Set Sdewan instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
        // Check if this svc already exists
        foundsvc := &corev1.Service{}
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, foundsvc)
        if err != nil && errors.IsNotFound(err) {
                reqLogger.Info("Creating a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
                err = r.client.Create(context.TODO(), svc)
                if err != nil {
                        return reconcile.Result{}, err
                }
                reqLogger.Info("A new Service created", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
        } else if err != nil {
                return reconcile.Result{}, err
        } else {
                reqLogger.Info("Service already exists", "Service.Namespace", foundsvc.Namespace, "Service.Name", foundsvc.Name)
        }

	return reconcile.Result{}, nil
}

// newSvcForCR returns a busybox pod with the same name/namespace as the cr
func newSvcForCR(cr *sdewanv1alpha1.Sdewan) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort {
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

// newConfigmapForCR returns a busybox pod with the same name/namespace as the cr
func newConfigmapForCR(cr *sdewanv1alpha1.Sdewan) *corev1.ConfigMap {
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
func newPodForCR(cr *sdewanv1alpha1.Sdewan) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	priv := true
	var netmaps []map[string]interface{}
	for _, net := range cr.Spec.Networks {
		netmaps = append(netmaps, map[string]interface{}{
			"name": net.Name,
			"interface": net.Interface,
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
			Annotations:    map[string]string{
				"k8s.v1.cni.cncf.io/networks": `[{ "name": "ovn-networkobj"}]`,
				"k8s.plugin.opnfv.org/nfn-network": `{ "type": "ovn4nfv", "interface": ` + string(netjson) + "}",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "sdewan",
					Image:   OpenwrtTag,
					Command: []string{"/bin/sh", "/tmp/sdewan/entrypoint.sh"},
					ImagePullPolicy: corev1.PullIfNotPresent,
					SecurityContext: &corev1.SecurityContext{
						Privileged: &priv,
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: cr.Name,
							ReadOnly: true,
							MountPath: "/tmp/sdewan",
						},
					},
				},
			},
			NodeSelector: map[string]string{"kubernetes.io/hostname": cr.Spec.Node},
			Volumes: volumes,
		},
	}
}


