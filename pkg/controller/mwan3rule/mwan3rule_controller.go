package mwan3rule

import (
	"fmt"
	"context"

	sdewanv1alpha1 "sdewan-operator/pkg/apis/sdewan/v1alpha1"
	"sdewan-operator/pkg/openwrt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_mwan3rule")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Mwan3Rule Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMwan3Rule{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("mwan3rule-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Mwan3Rule
	err = c.Watch(&source.Kind{Type: &sdewanv1alpha1.Mwan3Rule{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Mwan3Rule
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &sdewanv1alpha1.Mwan3Rule{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMwan3Rule implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMwan3Rule{}

// ReconcileMwan3Rule reconciles a Mwan3Rule object
type ReconcileMwan3Rule struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Mwan3Rule object and makes changes based on the state read
// and what is in the Mwan3Rule.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMwan3Rule) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Mwan3Rule")

	// Fetch the Mwan3Rule instance
	instance := &sdewanv1alpha1.Mwan3Rule{}
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

	foundSdewans := &sdewanv1alpha1.SdewanList{}
	// err = r.client.List(context.TODO(), foundSdewans, client.MatchingFields{"spec.mwan3Rule": instance.Name})
	err = r.client.List(context.TODO(), foundSdewans, client.InNamespace(instance.Namespace))
        if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("No sdewan instances using", "Mwan3Rule.Namespace", instance.Namespace)
        } else if err != nil {
                return reconcile.Result{}, err
        } else {
		for _, sdewan := range foundSdewans.Items {
			if sdewan.Spec.Mwan3Rule == instance.Name {
				reqLogger.Info("applying rules for sdewan: " + sdewan.Name)
				openwrtClient := openwrt.NewOpenwrtClient(sdewan.Name, "root", "")
				mwan3 := openwrt.Mwan3Client{openwrtClient}
				// apply policy
				netMap := sdewanv1alpha1.NetworkInterfaceMap(sdewan)
				for policyName, members := range instance.Spec.Policy {
					openwrtMembers := make([]openwrt.SdewanMember, len(members))
					for i, member := range members {
						openwrtMembers[i] = openwrt.SdewanMember{
							Interface: netMap[member.Network],
							Metric: fmt.Sprintf("%d", member.Metric),
							Weight: fmt.Sprintf("%d", member.Weight),
						}
					}

					_, err := mwan3.CreatePolicy(openwrt.SdewanPolicy{
						Name: policyName,
						Members:openwrtMembers})
					if (err != nil) {
						reqLogger.Info("Failed to add policy " + policyName)
					}
				}
				// apply rules TODO
			}
		}
        }

	return reconcile.Result{}, nil
}


// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *sdewanv1alpha1.Mwan3Rule) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
