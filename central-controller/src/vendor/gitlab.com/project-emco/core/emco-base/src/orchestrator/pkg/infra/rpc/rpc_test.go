// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package rpc

import (
	"testing"
)

func TestGetGrpcDialOpts(t *testing.T) {
	t.Run("Validate return values", func(t *testing.T) {
		dialOpts := getGrpcDialOpts()
		if len(dialOpts) == 0 {
			t.Fatal("getGrpcDialOpts returned nothing")
		}
	})
}
