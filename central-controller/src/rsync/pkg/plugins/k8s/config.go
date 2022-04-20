// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package k8s

import (
	"context"
)

// For K8s this function is not required to do any operation at this time
func (p *K8sProvider) ApplyConfig(ctx context.Context, config interface{}) error {
	return nil
}
func (p *K8sProvider) DeleteConfig(ctx context.Context, config interface{}) error {
	return nil
}
