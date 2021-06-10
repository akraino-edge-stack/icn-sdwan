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
	"encoding/base64"
	"encoding/json"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	pkgerrors "github.com/pkg/errors"
	"io"
	"log"
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
		RootCA: "",
		Ca:     "",
		Key:    "",
	}

	return &v, err
}

func (c *CertificateObjectManager) GetDeviceCertName(name string) string {
	device := module.DeviceObject{
		Metadata: module.ObjectMetaData{name, "", "", ""}}
	return device.GetCertName()
}

func GetRootCA(overlay_name string) string {
	overlay := GetManagerset().Overlay
	cu, _ := GetCertUtil()

	root_ca := cu.GetSelfSignedCA()
	interim_ca, _, _ := overlay.GetCertificate(overlay_name)

	root_ca += interim_ca

	return root_ca
}

func GetRootBaseCA() string {
	cu, _ := GetCertUtil()

	root_ca := cu.GetSelfSignedCA()

	return root_ca
}

func (c *CertificateObjectManager) CreateObject(m map[string]string, t module.ControllerObject) (module.ControllerObject, error) {
	// Create Certificate
	overlay := GetManagerset().Overlay
	overlay_name := m[OverlayResource]
	cert_name := c.GetDeviceCertName(t.GetMetadata().Name)

	ca, key, err := overlay.CreateCertificate(overlay_name, cert_name)
	if err != nil {
		log.Println(err)
		return c.CreateEmptyObject(), err
	}

	// DB Operation
	t, err = GetDBUtils().CreateObject(c, m, t)

	// Fill Certificate data
	if err == nil {
		to := t.(*module.CertificateObject)
		to.Data.RootCA = base64.StdEncoding.EncodeToString([]byte(GetRootCA(overlay_name)))
		to.Data.Ca = base64.StdEncoding.EncodeToString([]byte(ca))
		to.Data.Key = base64.StdEncoding.EncodeToString([]byte(key))

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
		overlay := GetManagerset().Overlay
		overlay_name := m[OverlayResource]
		cert_name := c.GetDeviceCertName(t.GetMetadata().Name)

		ca, key, err := overlay.CreateCertificate(overlay_name, cert_name)
		if err != nil {
			log.Println(err)
			return c.CreateEmptyObject(), err
		}

		to := t.(*module.CertificateObject)
		to.Data.RootCA = base64.StdEncoding.EncodeToString([]byte(GetRootCA(overlay_name)))
		to.Data.Ca = base64.StdEncoding.EncodeToString([]byte(ca))
		to.Data.Key = base64.StdEncoding.EncodeToString([]byte(key))

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

	// Delete certificate
	overlay := GetManagerset().Overlay
	cert_name := c.GetDeviceCertName(t.GetMetadata().Name)

	log.Println("Delete Certificate: " + cert_name)
	overlay.DeleteCertificate(cert_name)

	// DB Operation
	err = GetDBUtils().DeleteObject(c, m)

	return err
}

// Create or Get certificate for a device
func (c *CertificateObjectManager) GetOrCreateDC(overlay_name string, dev_name string) (string, string, string, error) {
	m := make(map[string]string)
	m[OverlayResource] = overlay_name
	t := &module.CertificateObject{Metadata: module.ObjectMetaData{dev_name, "", "", ""}}

	_, err := c.CreateObject(m, t)
	if err != nil {
		return "", "", "", err
	}

	return t.Data.RootCA, t.Data.Ca, t.Data.Key, nil
}

// Delete certificate for a device
func (c *CertificateObjectManager) DeleteDC(overlay_name string, dev_name string) error {
	m := make(map[string]string)
	m[OverlayResource] = overlay_name
	m[CertResource] = dev_name

	return c.DeleteObject(m)
}
