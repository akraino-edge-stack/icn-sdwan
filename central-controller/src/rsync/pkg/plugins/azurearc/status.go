// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package azurearc

// StartClusterWatcher watches for CR
// Same as K8s
func (c *AzureArcProvider) StartClusterWatcher() error {
	return nil
}

// ApplyStatusCR applies status CR
func (p *AzureArcProvider) ApplyStatusCR(name string, content []byte) error {

	return nil

}

// DeleteStatusCR deletes status CR
func (p *AzureArcProvider) DeleteStatusCR(name string, content []byte) error {

	return nil
}
