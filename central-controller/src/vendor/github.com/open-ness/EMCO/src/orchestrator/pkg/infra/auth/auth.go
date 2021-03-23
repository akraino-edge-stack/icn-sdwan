// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package auth

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"log"

	pkgerrors "github.com/pkg/errors"
)

// GetTLSConfig initializes a tlsConfig using the CA's certificate
// This config is then used to enable the server for mutual TLS
func GetTLSConfig(caCertFile string, certFile string, keyFile string) (*tls.Config, error) {

	// Initialize tlsConfig once
	caCert, err := ioutil.ReadFile(caCertFile)

	if err != nil {
		return nil, pkgerrors.Wrap(err, "Read CA Cert file")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		// Change to RequireAndVerify once we have mandatory certs
		ClientAuth: tls.VerifyClientCertIfGiven,
		ClientCAs:  caCertPool,
		MinVersion: tls.VersionTLS12,
	}

	certPEMBlk, err := readPEMBlock(certFile)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Read Cert File")
	}

	keyPEMBlk, err := readPEMBlock(keyFile)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Read Key File")
	}

	tlsConfig.Certificates = make([]tls.Certificate, 1)
	tlsConfig.Certificates[0], err = tls.X509KeyPair(certPEMBlk, keyPEMBlk)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Load x509 cert and key")
	}

	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}

func readPEMBlock(filename string) ([]byte, error) {

	pemData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Read PEM File")
	}

	pemBlock, rest := pem.Decode(pemData)
	if len(rest) > 0 {
		log.Println("Pemfile has extra data")
	}

	if x509.IsEncryptedPEMBlock(pemBlock) {
		password, err := ioutil.ReadFile(filename + ".pass")
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Read Password File")
		}

		pByte, err := base64.StdEncoding.DecodeString(string(password))
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Decode PEM Password")
		}

		pemData, err = x509.DecryptPEMBlock(pemBlock, pByte)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Decrypt PEM Data")
		}
		var newPEMBlock pem.Block
		newPEMBlock.Type = pemBlock.Type
		newPEMBlock.Bytes = pemData
		// Converting back to PEM from DER data you get from
		// DecryptPEMBlock
		pemData = pem.EncodeToMemory(&newPEMBlock)
	}

	return pemData, nil
}
