// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation
package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
//var label_check_log = logf.Log.WithName("label-validator")

func SetupLabelValidateWebhookWithManager(mgr ctrl.Manager) error {
	mgr.GetWebhookServer().Register(
		"/validate-label",
		&webhook.Admission{Handler: &labelValidator{Client: mgr.GetClient()}})
	return nil
}

// +kubebuilder:webhook:path=/validate-label,mutating=false,failurePolicy=fail,groups=apps;batch.sdewan.akraino.org,resources=deployments;mwan3policies;mwan3rules;firewallzones;firewallforwardings;firewallrules;firewallsnats;firewalldnats;cnfnats;cnfservices;cnfroutes;cnfrouterules;cnflocalservices;cnfstatuses;sdewanapplication;ipsecproposals;ipsechosts;ipsecsites,verbs=update,versions=v1,name=validate-label.akraino.org,admissionReviewVersions=v1,sideEffects=none

type labelValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (v *labelValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	var obj runtime.Object
	switch req.Kind.Kind {
	case "Deployment":
		obj = &appsv1.Deployment{}
	case "Mwan3Policy":
		obj = &Mwan3Policy{}
	case "Mwan3Rule":
		obj = &Mwan3Rule{}
	case "FirewallForwarding":
		obj = &FirewallForwarding{}
	case "FirewallZone":
		obj = &FirewallZone{}
	case "FirewallRule":
		obj = &FirewallRule{}
	case "CNFNAT":
		obj = &CNFNAT{}
	case "CNFRoute":
		obj = &CNFRoute{}
	case "CNFRouteRule":
		obj = &CNFRouteRule{}
	case "FirewallDNAT":
		obj = &FirewallDNAT{}
	case "FirewallSNAT":
		obj = &FirewallSNAT{}
	case "IpsecProposal":
		obj = &IpsecProposal{}
	case "IpsecHost":
		obj = &IpsecHost{}
	case "IpsecSite":
		obj = &IpsecSite{}
	case "CNFService":
		obj = &CNFService{}
	case "CNFLocalService":
		obj = &CNFLocalService{}
	case "CNFStatus":
		obj = &CNFStatus{}
	case "SdewanApplication":
		obj = &SdewanApplication{}
	default:
		return admission.Errored(
			http.StatusBadRequest,
			fmt.Errorf("Kind is not supported: %v", req.Kind))
	}

	if req.Operation != "UPDATE" {
		return admission.Allowed("")
	} else {
		oldobj := obj.DeepCopyObject()
		err1 := v.decoder.DecodeRaw(req.OldObject, oldobj)
		old_value := get_label(oldobj, "sdewanPurpose")
		err2 := v.decoder.Decode(req, obj)
		new_value := get_label(obj, "sdewanPurpose")
		if err1 != nil || err2 != nil {
			return admission.Errored(http.StatusBadRequest, errors.New("object Decode error"))
		}
		if old_value != new_value {
			return admission.Denied("Label 'sdewanPurpose' is immutable")
		}
		return admission.Allowed("")
	}
}

func get_label(oldobj runtime.Object, name string) string {
	metadata := reflect.ValueOf(oldobj).Elem().Field(1).Interface().(metav1.ObjectMeta)
	if value, ok := metadata.Labels[name]; ok {
		return value
	} else {
		return ""
	}
}

// labelValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *labelValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
