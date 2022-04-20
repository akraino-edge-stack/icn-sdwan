// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package k8sexp

import (
	"context"
)

// For K8s this function is not required to do any operation at this time
func (p *K8sProviderExp) ApplyConfig(ctx context.Context, config interface{}) error {
	return nil
}
func (p *K8sProviderExp) DeleteConfig(ctx context.Context, config interface{}) error {
	return nil
}
