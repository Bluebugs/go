//go:build goexperiment.spmd

// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends operand.go to handle SPMD type assignability rules.

package types2

import . "internal/types/errors"

// handleSPMDAssignability handles SPMD type assignability rules.
// Returns (handled, assignable, errorCode) where:
// - handled: true if SPMD types were involved in the check
// - assignable: true if assignment is valid
// - errorCode: error code if assignment is invalid
func (x *operand) handleSPMDAssignability(V, T Type, cause *string) (handled, assignable bool, code Code) {
	vSPMD, vIsSPMD := V.(*SPMDType)
	tSPMD, tIsSPMD := T.(*SPMDType)

	// If neither type is SPMD, we don't handle it
	if !vIsSPMD && !tIsSPMD {
		return false, false, 0
	}

	// At least one type is SPMD, so we handle this case
	handled = true

	// Case 1: Both are SPMD types
	if vIsSPMD && tIsSPMD {
		return true, x.checkSPMDtoSPMDAssignability(vSPMD, tSPMD, cause), IncompatibleAssign
	}

	// Case 2: V is SPMD, T is regular Go type
	if vIsSPMD && !tIsSPMD {
		return true, x.checkSPMDToRegularAssignability(vSPMD, T, cause), IncompatibleAssign
	}

	// Case 3: V is regular Go type, T is SPMD  
	if !vIsSPMD && tIsSPMD {
		return true, x.checkRegularToSPMDAssignability(V, tSPMD, cause), IncompatibleAssign
	}

	// This shouldn't be reached
	return true, false, IncompatibleAssign
}

// checkSPMDtoSPMDAssignability checks assignability between two SPMD types
func (x *operand) checkSPMDtoSPMDAssignability(vSPMD, tSPMD *SPMDType, cause *string) bool {
	// Rule 1: Identical SPMD types are always assignable
	if vSPMD.qualifier == tSPMD.qualifier && 
	   vSPMD.constraint == tSPMD.constraint && 
	   Identical(vSPMD.elem, tSPMD.elem) {
		return true
	}

	// Rule 2: Same qualifier, compatible element types
	if vSPMD.qualifier == tSPMD.qualifier && Identical(vSPMD.elem, tSPMD.elem) {
		// Rule 2a: Universal constraint compatibility (varying[n] -> varying[])
		if vSPMD.qualifier == VaryingQualifier && 
		   vSPMD.constraint > 0 && 
		   tSPMD.constraint == 0 {
			return true // varying[4] can be assigned to varying[]
		}
		
		// Rule 2b: Same constraint or no constraint
		if vSPMD.constraint == tSPMD.constraint {
			return true
		}
	}

	// Rule 3: Uniform to varying broadcast (automatic)
	if vSPMD.qualifier == UniformQualifier && 
	   tSPMD.qualifier == VaryingQualifier && 
	   Identical(vSPMD.elem, tSPMD.elem) {
		return true // uniform int can be assigned to varying int (broadcast)
	}

	// Rule 4: Varying to uniform is prohibited
	if vSPMD.qualifier == VaryingQualifier && tSPMD.qualifier == UniformQualifier {
		if cause != nil {
			*cause = "cannot assign varying expression to uniform variable"
		}
		return false
	}

	// Default: not assignable
	if cause != nil {
		*cause = "incompatible SPMD types"
	}
	return false
}

// checkSPMDToRegularAssignability checks assignability from SPMD type to regular Go type
func (x *operand) checkSPMDToRegularAssignability(vSPMD *SPMDType, T Type, cause *string) bool {
	// Only uniform types can be assigned to regular Go types
	if vSPMD.qualifier == UniformQualifier && Identical(vSPMD.elem, T) {
		return true
	}

	// Varying types cannot be assigned to regular Go types
	if vSPMD.qualifier == VaryingQualifier {
		if cause != nil {
			*cause = "cannot assign varying expression to non-SPMD variable"
		}
		return false
	}

	if cause != nil {
		*cause = "incompatible SPMD to regular type assignment"
	}
	return false
}

// checkRegularToSPMDAssignability checks assignability from regular Go type to SPMD type
func (x *operand) checkRegularToSPMDAssignability(V Type, tSPMD *SPMDType, cause *string) bool {
	// Regular Go types can be assigned to uniform SPMD types if element types match
	if tSPMD.qualifier == UniformQualifier && Identical(V, tSPMD.elem) {
		return true
	}

	// Regular Go types can be assigned to varying SPMD types (broadcast) if element types match
	if tSPMD.qualifier == VaryingQualifier && Identical(V, tSPMD.elem) {
		return true // automatic broadcast
	}

	if cause != nil {
		*cause = "incompatible regular to SPMD type assignment"
	}
	return false
}