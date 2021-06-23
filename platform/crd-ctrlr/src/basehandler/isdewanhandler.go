// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package basehandler

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sdewan.akraino.org/sdewan/openwrt"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ISdewanHandler interface {
	GetType() string
	GetName(instance runtime.Object) string
	GetFinalizer() string
	GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error)
	Convert(o runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error)
	IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool
	GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error)
	CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error)
	UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error)
	DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error
	Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error)
}
