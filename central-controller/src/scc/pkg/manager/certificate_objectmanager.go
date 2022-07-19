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
	"encoding/json"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	pkgerrors "github.com/pkg/errors"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/db"
	"io"
	"log"
)

const (
	OverlayKey   = "Overlay"
	DeviceKey    = "Device"
	HubKey       = "Hub"
	namespaceKey = "Namespace"
	InternalKey  = "internal"
)

type CertificateObjectKey struct {
	OverlayName     string `json:"overlay-name"`
	CertificateName string `json:"certificate-name"`
}

// IPRangeObjectManager implements the ControllerObjectManager
type CertificateObjectManager struct {
	BaseObjectManager
}

func NewCertificateObjectManager() *CertificateObjectManager {
	return &CertificateObjectManager{
		BaseObjectManager{
			storeName:      StoreName,
			tagMeta:        "certificate",
			depResManagers: []ControllerObjectManager{},
			ownResManagers: []ControllerObjectManager{},
		},
	}
}

func (c *CertificateObjectManager) GetResourceName() string {
	return CertResource
}

func (c *CertificateObjectManager) GetResourceStoredName(obj module.ControllerObject) string {
	v := obj.(*module.CertificateObject)
	cert_manager := GetManagerset().Cert
	return cert_manager.GetCertName(obj.GetMetadata().Name, v.Specification.ClusterType)
}

func (c *CertificateObjectManager) IsOperationSupported(oper string) bool {
	if oper == "PUT" {
		// Not allowed for gets
		return false
	}
	return true
}

func (c *CertificateObjectManager) CreateEmptyObject() module.ControllerObject {
	return &module.CertificateObject{}
}

func (c *CertificateObjectManager) GetStoreKey(m map[string]string, t module.ControllerObject, isCollection bool) (db.Key, error) {
	overlay_name := m[OverlayResource]

	key := CertificateObjectKey{
		OverlayName:     overlay_name,
		CertificateName: "",
	}
	if isCollection == true {
		return key, nil
	}

	to := t.(*module.CertificateObject)
	meta_name := to.Metadata.Name
	if meta_name != "" {
		cert_type := to.Specification.ClusterType
		meta_name = c.GetCertName(meta_name, cert_type)
	}
	res_name := m[CertResource]

	if res_name != "" {
		if meta_name != "" && res_name != meta_name {
			return key, pkgerrors.New("Resource name unmatched metadata name")
		}

		key.CertificateName = res_name
	} else {
		if meta_name == "" {
			return key, pkgerrors.New("Unable to find resource name")
		}

		key.CertificateName = meta_name
	}

	return key, nil
}

func (c *CertificateObjectManager) ParseObject(r io.Reader) (module.ControllerObject, error) {
	var v module.CertificateObject
	err := json.NewDecoder(r).Decode(&v)

	v.Data = module.CertificateObjectData{
		CA:   "",
		Cert: "",
		Key:  "",
	}

	return &v, err
}

func (c *CertificateObjectManager) GetCertName(name string, obj_type string) string {
	switch obj_type {
	case OverlayKey:
		overlay_manager := GetManagerset().Overlay
		return overlay_manager.CertName(name)
	case HubKey:
		hub := module.HubObject{
			Metadata: module.ObjectMetaData{name, "", "", ""}}
		return hub.GetCertName()
	case DeviceKey:
		device := module.DeviceObject{
			Metadata: module.ObjectMetaData{name, "", "", ""}}
		return device.GetCertName()
	}

	log.Println("unsupported obj_type specified in GetCertName")
	return ""
}

func GetRootCA() (string, error) {
	cu, err := GetCertUtil()
	if err != nil {
		log.Println(err)
		return "", err
	}

	return cu.GetSelfSignedCA(), nil
}

func GetCertChain(m map[string]string) (string, error) {
	certChain, err := GetRootCA()
	if err != nil {
		return "", err
	}

	cu, _ := GetCertUtil()
	if _, ok := m[namespaceKey]; !ok {
		log.Println("No namespace info while searching certs. Skip returning root ca.")
		return certChain, nil
	}

	for _, key := range []string{OverlayKey, DeviceKey, HubKey} {
		if val, ok := m[key]; ok {
			cert, _, err := cu.GetKeypair(val, m[namespaceKey])
			if err != nil {
				log.Println(err)
				return "", err
			}
			certChain = certChain + "___" + cert
		}
	}

	return certChain, nil
}

func (c *CertificateObjectManager) CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	// Create Certificate
	var cert_name, issuer_name string

	overlay := GetManagerset().Overlay
	overlay_name := m[OverlayResource]

	to := t.(*module.CertificateObject)
	isCAFlag := to.Specification.IsCA
	cert_type := to.Specification.ClusterType
	cert_name = c.GetCertName(to.Metadata.Name, cert_type)
	// Construct certificate name based on cluster type
	switch cert_type {
	case HubKey, DeviceKey:
		issuer_name = overlay.IssuerName(overlay_name)
	case OverlayKey:
		issuer_name = RootCAIssuerName
	}

	cu, err := GetCertUtil()
	if err != nil {
		log.Println(err)
		return c.CreateEmptyObject(), err
	}

	_, err = cu.CreateCertificate(cert_name, NameSpaceName, issuer_name, isCAFlag)
	if err != nil {
		log.Println("Failed to create overlay[" + overlay_name + "] certificate: " + err.Error())
		return c.CreateEmptyObject(), err
	}

	cert, key, err := cu.GetKeypair(cert_name, NameSpaceName)
	if err != nil {
		log.Println(err)
		return c.CreateEmptyObject(), err
	}

	certMap := make(map[string]string)
	certMap[namespaceKey] = NameSpaceName

	switch cert_type {
	case HubKey, DeviceKey:
		certMap[OverlayKey] = c.GetCertName(overlay_name, OverlayKey)
		if to.Specification.IsCA {
			certMap[cert_type] = cert_name
		}
	}

	ca, err := GetCertChain(certMap)
	if err != nil {
		log.Println(err)
		return c.CreateEmptyObject(), err
	}

	// DB Operation
	t, err = GetDBUtils().CreateObject(c, m, t)

	// Fill Certificate data
	if err == nil {
		to := t.(*module.CertificateObject)
		to.Data.Cert = cert
		to.Data.Key = key
		to.Data.CA = ca

		return t, nil
	} else {
		log.Println(err)
		return c.CreateEmptyObject(), err
	}
}

func (c *CertificateObjectManager) GetObject(m map[string]string) (module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObject(c, m)

	if err == nil {
		overlay_name := m[OverlayResource]

		to := t.(*module.CertificateObject)
		cert_type := to.Specification.ClusterType
		cert_name := c.GetCertName(to.Metadata.Name, cert_type)

		cu, err := GetCertUtil()
		if err != nil {
			log.Println(err)
			return c.CreateEmptyObject(), err
		}

		cert, key, err := cu.GetKeypair(cert_name, NameSpaceName)
		if err != nil {
			log.Println(err)
			return c.CreateEmptyObject(), err
		}

		certMap := make(map[string]string)
		certMap[namespaceKey] = NameSpaceName
		switch cert_type {
		case HubKey, DeviceKey:
			certMap[OverlayKey] = c.GetCertName(overlay_name, OverlayKey)
			if to.Specification.IsCA {
				certMap[cert_type] = cert_name
			}
		}

		ca, err := GetCertChain(certMap)
		if err != nil {
			log.Println(err)
			return c.CreateEmptyObject(), err
		}

		to.Data.Cert = cert
		to.Data.Key = key
		to.Data.CA = ca

		return t, nil
	} else {
		log.Println(err)
		return c.CreateEmptyObject(), err
	}
}

func (c *CertificateObjectManager) GetObjects(m map[string]string) ([]module.ControllerObject, error) {
	// DB Operation
	t, err := GetDBUtils().GetObjects(c, m)

	return t, err
}

func (c *CertificateObjectManager) UpdateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	return c.CreateEmptyObject(), pkgerrors.New("Not implemented")
}

func (c *CertificateObjectManager) DeleteObject(m map[string]string) error {
	t, err := c.GetObject(m)
	if err != nil {
		return pkgerrors.Wrap(err, "Certificate is not available")
	}

	to := t.(*module.CertificateObject)
	cert_type := to.Specification.ClusterType
	cert_name := c.GetCertName(to.Metadata.Name, cert_type)

	// Delete certificate
	cu, err := GetCertUtil()
	if err != nil {
		log.Println(err)
	}

	log.Println("Delete Certificate: " + m[CertResource])
	err = cu.DeleteCertificate(m[CertResource], NameSpaceName)
	if err != nil {
		log.Println("Failed to delete " + cert_name + " certificate: " + err.Error())
	}

	// DB Operation
	err = GetDBUtils().DeleteObject(c, m)

	return err
}

func (c *CertificateObjectManager) GetOrCreateDC(overlay_name string, dev_name string, isCA bool) (string, string, string, error) {
	return c.GetOrCreateCertificateByType(overlay_name, dev_name, DeviceKey, isCA)
}

func (c *CertificateObjectManager) GetOrCreateCertificateByType(overlay_name string, dev_name string, dev_type string, isCA bool) (string, string, string, error) {
	var certObj module.CertificateObject

	m := make(map[string]string)
	m[OverlayResource] = overlay_name
	switch dev_type {
	case OverlayKey:
		m[CertResource] = c.GetCertName(overlay_name, dev_type)
		certObj = module.CertificateObject{Metadata: module.ObjectMetaData{overlay_name, "", InternalKey, ""}, Specification: module.CertificateObjectSpec{isCA, OverlayKey}}
	case DeviceKey, HubKey:
		m[CertResource] = c.GetCertName(dev_name, dev_type)
		certObj = module.CertificateObject{Metadata: module.ObjectMetaData{dev_name, "", InternalKey, ""}, Specification: module.CertificateObjectSpec{isCA, dev_type}}
	}

	t, err := c.GetObject(m)
	if err != nil {
		_, err := c.CreateObject(m, &certObj)
		if err != nil {
			return "", "", "", err
		}
		return certObj.Data.CA, certObj.Data.Cert, certObj.Data.Key, nil
	}

	to := t.(*module.CertificateObject)

	return to.Data.CA, to.Data.Cert, to.Data.Key, nil
}

func (c *CertificateObjectManager) DeleteCertificateByType(overlay_name string, dev_name string, dev_type string) error {
	var cert_name string
	m := make(map[string]string)

	m[OverlayResource] = overlay_name
	switch dev_type {
	case HubKey, DeviceKey:
		cert_name = c.GetCertName(dev_name, dev_type)
	case OverlayKey:
		cert_name = c.GetCertName(overlay_name, dev_type)
	}
	m[CertResource] = cert_name

	return c.DeleteObject(m)
}
