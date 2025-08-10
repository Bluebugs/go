// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends typexpr.go to handle SPMD type validation.

package types2

import (
	"cmd/compile/internal/syntax"
	"internal/buildcfg"
	. "internal/types/errors"
)

// validVarTypeSPMD extends validVarType to handle SPMD types.
// Returns true if the type was handled as an SPMD type, false otherwise.
func (check *Checker) validVarTypeSPMD(e syntax.Expr, typ Type) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}

	// Validate SPMD type restrictions for all types (not just SPMD types)
	if err := check.validateSPMDTypeRestrictions(typ); err != "" {
		check.errorf(e, InvalidSPMDType, "%s", err)
		return true // We handled the type validation (reported error)
	}

	// Skip constraint interface validation for SPMD types
	if _, ok := typ.(*SPMDType); ok {
		// SPMD types are always valid as variable types
		return true
	}

	return false
}