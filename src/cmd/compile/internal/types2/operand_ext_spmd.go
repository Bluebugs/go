// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends operand.go to handle SPMD type assignability rules.

package types2

import (
	"internal/buildcfg"
	. "internal/types/errors"
)

// isEmptyInterface reports whether T is interface{}
func isEmptyInterface(T Type) bool {
	if iface, ok := T.(*Interface); ok {
		return iface.Empty()
	}
	return false
}

// handleSPMDAssignability handles SPMD type assignability rules.
// Returns (handled, assignable, errorCode) where:
// - handled: true if SPMD types were involved in the check
// - assignable: true if assignment is valid
// - errorCode: error code if assignment is invalid
func (x *operand) handleSPMDAssignability(V, T Type, cause *string) (handled, assignable bool, code Code) {
	if !buildcfg.Experiment.SPMD {
		return false, false, 0
	}

	vSPMD, vIsSPMD := V.(*SPMDType)
	tSPMD, tIsSPMD := T.(*SPMDType)

	// Check for pointer-to-SPMD cases
	if !vIsSPMD && !tIsSPMD {
		// Check if this is a pointer-to-SPMD assignment case
		if x.checkPointerToSPMDAssignability(V, T, cause) {
			return true, true, 0
		}
		// If neither type is SPMD and not a pointer-to-SPMD case, we don't handle it
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
	if vSPMD.qualifier == tSPMD.qualifier {
		// Check element type compatibility (identical or convertible)
		elementsCompatible := Identical(vSPMD.elem, tSPMD.elem) || x.convertibleToSPMDElement(vSPMD.elem, tSPMD.elem)

		if elementsCompatible {
			// Reject upcasting for varying types in assignments (same as in conversions)
			if vSPMD.qualifier == VaryingQualifier && !Identical(vSPMD.elem, tSPMD.elem) {
				srcSize := spmdBasicSize(vSPMD.elem)
				dstSize := spmdBasicSize(tSPMD.elem)
				if srcSize > 0 && dstSize > 0 && dstSize > srcSize {
					if cause != nil {
						*cause = "cannot assign varying type to larger element type (would exceed SIMD register capacity)"
					}
					return false
				}
			}

			// Rule 2a: Constraint compatibility rules
			if vSPMD.qualifier == VaryingQualifier {
				// ALLOWED: Constrained varying can be assigned to universal varying[] (varying[4] -> varying[])
				if vSPMD.constraint > 0 && tSPMD.constraint == 0 {
					return true // varying[4] can be assigned to varying[]
				}

				// ALLOWED: Unconstrained varying can be assigned to any constrained varying (broadcast)
				if vSPMD.constraint == -1 && tSPMD.constraint >= 0 {
					return true // varying can be assigned to varying[4] or varying[]
				}

				// ALLOWED: Constrained varying to unconstrained varying (varying[4] -> varying)
				// The TinyGo backend handles width mismatch via spmdResizeVector().
				if vSPMD.constraint > 0 && tSPMD.constraint == -1 {
					return true
				}
			}

			// Rule 2b: Same constraint or no constraint
			if vSPMD.constraint == tSPMD.constraint {
				return true
			}
		}
	}

	// Rule 3: Uniform to varying broadcast (automatic)
	if vSPMD.qualifier == UniformQualifier && tSPMD.qualifier == VaryingQualifier {
		// Allow uniform to varying broadcast with element type conversion
		elementsCompatible := Identical(vSPMD.elem, tSPMD.elem) || x.convertibleToSPMDElement(vSPMD.elem, tSPMD.elem)
		if elementsCompatible {
			return true // uniform int can be assigned to varying int (broadcast)
		}
	}

	// Rule 4: Interface assignment support
	// Any SPMD type can be assigned to varying interface{}
	if tSPMD.qualifier == VaryingQualifier && isEmptyInterface(tSPMD.elem) {
		return true // varying int can be assigned to varying interface{}
	}

	// uniform T can be assigned to uniform interface{}
	if vSPMD.qualifier == UniformQualifier &&
		tSPMD.qualifier == UniformQualifier &&
		isEmptyInterface(tSPMD.elem) {
		return true // uniform int can be assigned to uniform interface{}
	}

	// Rule 5: Varying to uniform is prohibited
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
	// Handle interface{} assignment - both uniform and varying can be assigned to interface{}
	if isEmptyInterface(T) {
		return true // Both uniform and varying types can be assigned to interface{}
	}

	// Only uniform types can be assigned to regular Go types
	if vSPMD.qualifier == UniformQualifier && Identical(vSPMD.elem, T) {
		return true
	}

	// Varying types cannot be assigned to regular Go types (except interface{})
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
	// Regular Go types can be assigned to uniform SPMD types if element types match or are convertible
	if tSPMD.qualifier == UniformQualifier {
		elementsCompatible := Identical(V, tSPMD.elem) || ConvertibleTo(V, tSPMD.elem)
		if elementsCompatible {
			return true
		}
	}

	// Regular Go types can be assigned to varying SPMD types (broadcast) if element types match or are convertible
	if tSPMD.qualifier == VaryingQualifier {
		elementsCompatible := Identical(V, tSPMD.elem) || ConvertibleTo(V, tSPMD.elem)
		if elementsCompatible {
			return true // automatic broadcast
		}
	}

	if cause != nil {
		*cause = "incompatible regular to SPMD type assignment"
	}
	return false
}

// convertibleToSPMD checks if x can be converted to T via SPMD-specific conversion rules.
// This is more permissive than assignability and allows conversions between different
// constrained varying types with compatible element types.
func (x *operand) convertibleToSPMD(check *Checker, T Type, cause *string) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}

	V := x.typ()
	vSPMD, vIsSPMD := V.(*SPMDType)
	tSPMD, tIsSPMD := T.(*SPMDType)

	// Array-to-constrained-varying: [N]T → Varying[T, N]
	if !vIsSPMD && tIsSPMD {
		if tSPMD.IsVarying() && tSPMD.constraint > 0 {
			if arr, ok := V.Underlying().(*Array); ok {
				if arr.Len() == tSPMD.constraint && Identical(arr.Elem(), tSPMD.elem) {
					return true
				}
			}
		}
		return false
	}

	// Only handle SPMD to SPMD conversions beyond this point
	if !vIsSPMD || !tIsSPMD {
		return false
	}

	// Rule 1: Same qualifier and convertible element types
	if vSPMD.qualifier == tSPMD.qualifier {
		if x.isElementConvertible(vSPMD.elem, tSPMD.elem) {
			// Reject upcasting for varying types: converting to a larger element type
			// would require more SIMD register bits (e.g., Varying[uint16] has 8 lanes
			// on 128-bit SIMD, but Varying[uint32] only has 4 lanes).
			if vSPMD.qualifier == VaryingQualifier {
				srcSize := spmdBasicSize(vSPMD.elem)
				dstSize := spmdBasicSize(tSPMD.elem)
				if srcSize > 0 && dstSize > 0 && dstSize > srcSize {
					if cause != nil {
						*cause = "cannot upcast varying type to larger element type (would exceed SIMD register capacity)"
					}
					return false
				}
			}
			return true
		}
	}

	// Rule 2: Uniform to varying conversion (broadcast with conversion)
	if vSPMD.qualifier == UniformQualifier && tSPMD.qualifier == VaryingQualifier {
		if x.isElementConvertible(vSPMD.elem, tSPMD.elem) {
			return true
		}
	}

	// Rule 3: Varying to uniform is forbidden (would lose data)
	if vSPMD.qualifier == VaryingQualifier && tSPMD.qualifier == UniformQualifier {
		if cause != nil {
			*cause = "cannot convert varying expression to uniform type"
		}
		return false
	}

	return false
}

// isElementConvertible checks if source element type can be converted to target element type
// without creating circular calls to the conversion system.
func (x *operand) isElementConvertible(sourceElem, targetElem Type) bool {
	// Handle identical types first
	if Identical(sourceElem, targetElem) {
		return true
	}

	// Check if we need basic types for byte/rune conversions
	targetBasic, targetOK := targetElem.Underlying().(*Basic)

	// Allow numeric conversions (same as Go conversion rules)
	if isNumeric(sourceElem) && isNumeric(targetElem) {
		return true
	}

	// Allow integer to string conversion
	if isInteger(sourceElem) && isString(targetElem) {
		return true
	}

	// Allow string to slice of bytes/runes conversion  
	if isString(sourceElem) && targetOK && (isByte(targetBasic) || isRune(targetBasic)) {
		return true
	}

	return false
}

// Helper functions for basic type checking
func isByte(t *Basic) bool {
	return t.kind == Byte
}

func isRune(t *Basic) bool {
	return t.kind == Rune
}

// convertibleToSPMDElement checks if source element type can be converted to target element type
// This handles cases like int32 -> int, float32 -> float64, etc.
func (x *operand) convertibleToSPMDElement(sourceElem, targetElem Type) bool {
	// Use Go's standard type convertibility rules
	// This includes numeric conversions, string conversions, etc.
	return ConvertibleTo(sourceElem, targetElem)
}

// spmdBasicSize returns the byte size of a numeric basic type for SPMD upcasting checks.
// Returns 0 for non-numeric or platform-dependent types (int, uint, uintptr).
func spmdBasicSize(t Type) int64 {
	b, ok := t.Underlying().(*Basic)
	if !ok {
		return 0
	}
	switch b.kind {
	case Int8, Uint8:
		return 1
	case Int16, Uint16:
		return 2
	case Int32, Uint32, Float32:
		return 4
	case Int64, Uint64, Float64:
		return 8
	default:
		// Platform-dependent types (int, uint, uintptr) and non-numeric:
		// skip the upcast check to avoid false positives.
		return 0
	}
}

// checkPointerToSPMDAssignability checks if V can be assigned to T where T is a pointer to an SPMD type
// For example: *int assignable to *uniform int
func (x *operand) checkPointerToSPMDAssignability(V, T Type, cause *string) bool {
	// Check if T is a pointer to an SPMD type
	if ptrT, ok := T.(*Pointer); ok {
		if spmdBase, ok := ptrT.base.(*SPMDType); ok {
			// Check if V is a pointer to a compatible type
			if ptrV, ok := V.(*Pointer); ok {
				// For uniform types, allow assignment from compatible pointers
				if spmdBase.qualifier == UniformQualifier {
					elementsCompatible := Identical(ptrV.base, spmdBase.elem) || ConvertibleTo(ptrV.base, spmdBase.elem)
					if elementsCompatible {
						return true
					}
				}
			}
		}
	}
	return false
}
