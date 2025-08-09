//go:build !goexperiment.spmd

// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import . "internal/types/errors"

// handleSPMDAssignability stub for when SPMD experiment is disabled.
func (x *operand) handleSPMDAssignability(V, T Type, cause *string) (bool, bool, Code) {
	// SPMD types don't exist when experiment is disabled
	return false, false, 0
}