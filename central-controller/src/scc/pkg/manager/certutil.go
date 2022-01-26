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
	"context"
	kclient "github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/client"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	certmanagerv1 "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/typed/certmanager/v1"
	pkgerrors "github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log"
	"time"
)

const SELFSIGNEDCA = "sdewan-controller"

type CertUtil struct {
	client    certmanagerv1.CertmanagerV1Interface
	k8sclient corev1.CoreV1Interface
}

var certutil = CertUtil{}

func GetCertUtil() (*CertUtil, error) {
	var err error
	if certutil.client == nil || certutil.k8sclient == nil {
		certutil.client, certutil.k8sclient, err = kclient.NewClient("", "", []byte{}).GetCMClients()
	}

	return &certutil, err
}

func (c *CertUtil) CreateNamespace(name string) (*v1.Namespace, error) {
	ns, err := c.k8sclient.Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		return ns, nil
	}

	log.Println("Create Namespace: " + name)
	return c.k8sclient.Namespaces().Create(context.TODO(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}, metav1.CreateOptions{})
}

func (c *CertUtil) DeleteNamespace(name string) error {
	return c.k8sclient.Namespaces().Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func (c *CertUtil) GetIssuer(name string, namespace string) (*certv1.Issuer, error) {
	return c.client.Issuers(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (c *CertUtil) DeleteIssuer(name string, namespace string) error {
	return c.client.Issuers(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func (c *CertUtil) CreateSelfSignedIssuer(name string, namespace string) (*certv1.Issuer, error) {
	issuer, err := c.GetIssuer(name, namespace)
	if err == nil {
		return issuer, nil
	}

	// Not existing issuer, create a new one
	return c.client.Issuers(namespace).Create(context.TODO(), &certv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				SelfSigned: &certv1.SelfSignedIssuer{},
			},
		},
	}, metav1.CreateOptions{})
}

func (c *CertUtil) CreateCAIssuer(name string, namespace string, caname string) (*certv1.Issuer, error) {
	issuer, err := c.GetIssuer(name, namespace)
	if err == nil {
		return issuer, nil
	}

	// Not existing issuer, create a new one
	return c.client.Issuers(namespace).Create(context.TODO(), &certv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: certv1.IssuerSpec{
			IssuerConfig: certv1.IssuerConfig{
				CA: &certv1.CAIssuer{
					SecretName: c.GetCertSecretName(caname),
				},
			},
		},
	}, metav1.CreateOptions{})
}

func (c *CertUtil) GetCertSecretName(name string) string {
	return name + "-cert-secret"
}

func (c *CertUtil) GetCertificate(name string, namespace string) (*certv1.Certificate, error) {
	return c.client.Certificates(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (c *CertUtil) DeleteCertificate(name string, namespace string) error {
	return c.client.Certificates(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func (c *CertUtil) CreateCertificate(name string, namespace string, issuer string, isCA bool) (*certv1.Certificate, error) {
	cert, err := c.GetCertificate(name, namespace)
	if err == nil {
		return cert, nil
	}

	// Not existing cert, create a new one
	// Todo: add Duration, RenewBefore, DNSNames
	cert, err = c.client.Certificates(namespace).Create(context.TODO(), &certv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: certv1.CertificateSpec{
			CommonName: name,
			// Duration: duration,
			// RenewBefore: renewBefore,
			// DNSNames: dnsNames,
			SecretName: c.GetCertSecretName(name),
			IssuerRef: cmmeta.ObjectReference{
				Name: issuer,
				Kind: "Issuer",
			},
			IsCA: isCA,
		},
	}, metav1.CreateOptions{})

	if err == nil {
		if c.IsCertReady(name, namespace) {
			return cert, nil
		} else {
			return cert, pkgerrors.New("Failed to get certificate " + name)
		}
	}

	return cert, err
}

func (c *CertUtil) IsCertReady(name string, namespace string) bool {
	err := wait.PollImmediate(time.Second, time.Second*20,
		func() (bool, error) {
			var err error
			var crt *certv1.Certificate
			crt, err = c.GetCertificate(name, namespace)
			if err != nil {
				log.Println("Failed to find certificate " + name + ": " + err.Error())
				return false, err
			}
			curConditions := crt.Status.Conditions
			for _, cond := range curConditions {
				if certv1.CertificateConditionReady == cond.Type && cmmeta.ConditionTrue == cond.Status {
					return true, nil
				}
			}
			log.Println("Waiting for Certificate " + name + " to be ready.")
			return false, nil
		},
	)

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func (c *CertUtil) GetKeypair(certname string, namespace string) (string, string, error) {
	secret, err := c.k8sclient.Secrets(namespace).Get(
		context.TODO(),
		c.GetCertSecretName(certname),
		metav1.GetOptions{})
	if err != nil {
		log.Println("Failed to get certificate's key pair: " + err.Error())
		return "", "", err
	}

	return string(secret.Data["tls.crt"]), string(secret.Data["tls.key"]), nil
}

func (c *CertUtil) GetSelfSignedCA() string {
	ca, _, _ := c.GetKeypair(RootCertName, NameSpaceName)
	return ca
}
