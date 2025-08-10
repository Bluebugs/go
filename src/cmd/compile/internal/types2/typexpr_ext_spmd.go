// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends typexpr.go to handle SPMD types in the main typInternal function.

package types2

import (
	"cmd/compile/internal/syntax"
	"go/constant"
	"internal/buildcfg"
	. "internal/types/errors"
)

// handleSPMDTypeExpr handles SPMD type expressions in the typInternal method.
// Returns the resulting type and true if handled, or nil and false if not an SPMD type.
func (check *Checker) handleSPMDTypeExpr(e syntax.Expr, def *TypeName) (Type, bool) {
	if !buildcfg.Experiment.SPMD {
		return nil, false
	}

	// First try direct SPMD type
	if spmdExpr, ok := e.(*syntax.SPMDType); ok {
		return check.processSPMDType(spmdExpr, def)
	}
	
	// Try to detect SPMD expressions that were parsed as operations in type switch contexts
	// This handles cases like "varying *int" that got parsed as multiplication
	if operationExpr, ok := e.(*syntax.Operation); ok {
		return check.handleSPMDOperationExpr(operationExpr, def)
	}
	
	// Handle bare SPMD keywords parsed as Name nodes
	if nameExpr, ok := e.(*syntax.Name); ok {
		if nameExpr.Value == "varying" || nameExpr.Value == "uniform" {
			// Bare "varying" or "uniform" without a type - treat as error
			check.errorf(e, NotAType, "incomplete SPMD type: %s requires element type", nameExpr.Value)
			return Typ[Invalid], true
		}
	}
	
	return nil, false
}

// processSPMDType handles direct SPMDType syntax nodes
func (check *Checker) processSPMDType(spmdExpr *syntax.SPMDType, def *TypeName) (Type, bool) {
	var qualifier SPMDQualifier
	var constraint int64 = -1

	// Determine the qualifier
	if syntax.IsUniformToken(spmdExpr.Qualifier) {
		qualifier = UniformQualifier
	} else if syntax.IsVaryingToken(spmdExpr.Qualifier) {
		qualifier = VaryingQualifier
	} else {
		// Not a valid SPMD qualifier - let other handlers deal with it
		return nil, false
	}

	// Handle varying constraints
	if qualifier == VaryingQualifier && spmdExpr.Constraint != nil {
		// Check for universal constraint marker (varying[])
		if name, ok := spmdExpr.Constraint.(*syntax.Name); ok && name.Value == "__universal_constraint__" {
			// varying[] - universal constraint
			constraint = 0
		} else {
			// varying[n] - numeric constraint
			var x operand
			check.expr(nil, &x, spmdExpr.Constraint)
			if x.mode != constant_ {
				check.error(spmdExpr.Constraint, InvalidConstVal, "constraint must be compile-time constant")
				return nil, false
			}

			if !isInteger(x.typ) {
				check.error(spmdExpr.Constraint, InvalidConstVal, "constraint must be an integer constant")
				return nil, false
			}

			val, ok := constant.Int64Val(x.val)
			if !ok {
				check.error(spmdExpr.Constraint, InvalidConstVal, "constraint out of range")
				return nil, false
			}

			if val < 1 {
				check.error(spmdExpr.Constraint, InvalidConstVal, "constraint must be positive")
				return nil, false
			}

			constraint = val
		}
	}

	// Type-check the element type first to validate capacity
	elem := check.varType(spmdExpr.Elem)
	if !isValid(elem) {
		// Return nil, false instead of Typ[Invalid] to avoid compiler crash
		return nil, false
	}

	// Validate constrained varying capacity (512 bits = 64 bytes limit)
	// Apply only to specific constrained varying cases that exceed capacity
	if qualifier == VaryingQualifier && constraint > 0 {
		// Calculate total capacity: constraint * element_size
		elementSize := check.calculateTypeSize(elem)
		totalSize := constraint * elementSize
		
		// Apply capacity validation based on specific test expectations
		const maxConstrainedCapacity = 64 // 512 bits = 64 bytes
		
		if totalSize > maxConstrainedCapacity {
			check.error(spmdExpr.Constraint, InvalidConstVal, "constrained varying capacity exceeded")
			return nil, false
		}
	}
	
	// Note: For unconstrained varying with arrays (varying [n]T), we don't apply capacity limits
	// because these are standard varying types applied to array types

	// Validate SPMD type restrictions (pointers, maps, channels)
	if err := check.validateSPMDTypeRestrictions(elem); err != "" {
		check.errorf(spmdExpr.Elem, InvalidSPMDType, "%s", err)
		return nil, false
	}

	// Create the SPMD type
	var typ *SPMDType
	if qualifier == UniformQualifier {
		typ = NewUniform(elem)
	} else {
		typ = NewVaryingConstrained(elem, constraint)
	}

	setDefType(def, typ)
	return typ, true
}

// validateSPMDTypeRestrictions validates type restrictions for SPMD types.
// Returns an error message if the type is invalid, empty string if valid.
func (check *Checker) validateSPMDTypeRestrictions(typ Type) string {
	switch t := typ.(type) {
	case *Pointer:
		// Pointers to varying types are allowed (*varying T is valid)
		// Only restriction is taking address of varying variables (handled in expr.go)
		// Recursively check the pointed-to type
		return check.validateSPMDTypeRestrictions(t.base)

	case *Map:
		// Check if map key is varying type: map[varying T] is forbidden
		if spmdType, ok := t.key.(*SPMDType); ok && spmdType.IsVarying() {
			return "varying map keys not supported"
		}
		// Recursively check both key and value types
		if err := check.validateSPMDTypeRestrictions(t.key); err != "" {
			return err
		}
		return check.validateSPMDTypeRestrictions(t.elem)

	case *Chan:
		// Channels can carry varying types - no restrictions
		// Recursively check the element type for other restrictions
		return check.validateSPMDTypeRestrictions(t.elem)

	case *Array:
		// Arrays of pointers to varying types are allowed ([n]*varying T is valid)
		// Check array element type recursively
		return check.validateSPMDTypeRestrictions(t.elem)

	case *Slice:
		// Check slice element type recursively
		return check.validateSPMDTypeRestrictions(t.elem)

	case *SPMDType:
		// Check the underlying element type
		return check.validateSPMDTypeRestrictions(t.elem)

	default:
		// Other types are fine
		return ""
	}
}

// calculateTypeSize returns the size in bytes of a single element type for capacity calculations
// For constrained varying, this calculates the size of T in varying[n] T
func (check *Checker) calculateTypeSize(t Type) int64 {
	switch t := under(t).(type) {
	case *Basic:
		switch t.kind {
		case Bool, Uint8, Int8:
			return 1
		case Uint16, Int16:
			return 2
		case Uint32, Int32, Float32:
			return 4
		case Uint64, Int64, Float64:
			return 8
		case Uintptr, UnsafePointer:
			return 8 // Assume 64-bit pointers
		default:
			return 8 // Default safe size
		}
	case *Array:
		// For arrays in constrained varying context, we want the total array size
		// because varying[n] [4]int means n instances of [4]int
		elemSize := check.calculateTypeSize(t.elem)
		return t.len * elemSize
	case *Slice:
		return 24 // Slice header: pointer + len + cap (3 * 8 bytes)
	case *Pointer:
		return 8 // 64-bit pointer
	default:
		return 8 // Default safe size
	}
}

// calculateBaseTypeSize returns the size of the base element type, ignoring array dimensions
// For constrained varying capacity validation: varying[n] [4]int -> size of int (not [4]int)
func (check *Checker) calculateBaseTypeSize(t Type) int64 {
	switch t := under(t).(type) {
	case *Basic:
		switch t.kind {
		case Bool, Uint8, Int8:
			return 1
		case Uint16, Int16:
			return 2
		case Uint32, Int32, Float32:
			return 4
		case Uint64, Int64, Float64:
			return 8
		case Uintptr, UnsafePointer:
			return 8 // Assume 64-bit pointers
		default:
			return 8 // Default safe size
		}
	case *Array:
		// For constrained varying, drill down to the base element type
		// varying[16] [4]int -> size of int (4 bytes), not [4]int (16 bytes)
		return check.calculateBaseTypeSize(t.elem)
	case *Slice:
		return 24 // Slice header: pointer + len + cap (3 * 8 bytes)
	case *Pointer:
		return 8 // 64-bit pointer
	default:
		return 8 // Default safe size
	}
}

// handleSPMDOperationExpr detects SPMD type expressions that were parsed as operations
// This handles cases like "varying *int" that get parsed as multiplication operations
func (check *Checker) handleSPMDOperationExpr(opExpr *syntax.Operation, def *TypeName) (Type, bool) {
	// Check for patterns like "varying *int" (parsed as multiplication)
	if opExpr.Op == syntax.Mul {
		// Check if left operand is a SPMD qualifier
		if leftName, ok := opExpr.X.(*syntax.Name); ok {
			if leftName.Value == "varying" || leftName.Value == "uniform" {
				// This looks like "varying *SomeType" - get the element type
				elemType := check.varType(opExpr.Y)
				if !isValid(elemType) {
					return nil, false
				}
				
				// Create pointer to element type
				ptrType := &Pointer{base: elemType}
				
				// Wrap in appropriate SPMD type
				var spmdType Type
				if leftName.Value == "varying" {
					spmdType = NewVarying(ptrType)
				} else {
					spmdType = NewUniform(ptrType)
				}
				
				setDefType(def, spmdType)
				return spmdType, true
			}
		}
	}
	
	// Check for patterns like "*varying int" (trickier - need to examine parsed structure)
	// This might be parsed differently depending on precedence
	
	return nil, false
}
