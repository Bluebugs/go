// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends unify.go to handle SPMD types in generic type inference.

package types2

import (
	"internal/buildcfg"
)

// handleSPMDUnification handles SPMD type unification for generic type inference.
// This is called when trying to unify a type parameter with an SPMD type.
// Returns true if handled and unified successfully, false if should fall through to standard unification.
func (u *unifier) handleSPMDUnification(x, y Type, mode unifyMode) (handled bool, unified bool) {
	if !buildcfg.Experiment.SPMD {
		return false, false
	}

	// Check if we're dealing with SPMD types
	xSPMD, xIsSPMD := x.(*SPMDType)
	ySPMD, yIsSPMD := y.(*SPMDType)

	// If neither is SPMD, not our concern
	if !xIsSPMD && !yIsSPMD {
		return false, false
	}

	// Case 1: Both are SPMD types - unify their element types and qualifiers
	if xIsSPMD && yIsSPMD {
		// SPMD types unify if:
		// 1. Same qualifier (uniform/varying)
		// 2. Same constraint (if varying)
		// 3. Element types unify
		if xSPMD.qualifier != ySPMD.qualifier {
			return true, false // Different qualifiers don't unify
		}

		if xSPMD.qualifier == VaryingQualifier && xSPMD.constraint != ySPMD.constraint {
			return true, false // Different constraints don't unify
		}

		// Recursively unify element types
		return true, u.nify(xSPMD.elem, ySPMD.elem, mode, nil)
	}

	// Case 2: One is SPMD, one is not
	// For generic type inference with lanes.Count[T any](varying T), we want:
	// - varying int32 to unify with varying T, inferring T = int32
	// - uniform int32 to unify with uniform T, inferring T = int32
	
	var spmdType *SPMDType
	var regularType Type
	var spmdIsX bool
	
	if xIsSPMD {
		spmdType = xSPMD
		regularType = y
		spmdIsX = true
	} else {
		spmdType = ySPMD
		regularType = x
		spmdIsX = false
	}

	// For generic inference, we want to extract the element type from SPMD type
	// and unify it with the regular type (which might be a type parameter)
	elemType := spmdType.elem

	// Recursively unify element type with the regular type
	if spmdIsX {
		return true, u.nify(elemType, regularType, mode, nil)
	} else {
		return true, u.nify(regularType, elemType, mode, nil)
	}
}