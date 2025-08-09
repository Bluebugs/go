//go:build !goexperiment.spmd

// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import "cmd/compile/internal/syntax"

// handleSPMDTypeExpr stub for when SPMD experiment is disabled.
func (check *Checker) handleSPMDTypeExpr(e syntax.Expr, def *TypeName) (Type, bool) {
	// SPMD types don't exist when experiment is disabled
	return nil, false
}