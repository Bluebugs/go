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

// handleSPMDIndexExpr checks if an IndexExpr is lanes.Varying[T] or lanes.Varying[T, N]
// and creates the corresponding SPMDType. Returns (type, true) if handled.
func (check *Checker) handleSPMDIndexExpr(e *syntax.IndexExpr, def *TypeName) (Type, bool) {
	if !buildcfg.Experiment.SPMD {
		return nil, false
	}

	if !check.isLanesVaryingExpr(e.X) {
		return nil, false
	}

	return check.processLanesVaryingType(e, def)
}

// isLanesVaryingExpr checks if expr resolves to lanes.Varying
func (check *Checker) isLanesVaryingExpr(x syntax.Expr) bool {
	// Case 1: Package-qualified lanes.Varying (from importing packages)
	if sel, ok := x.(*syntax.SelectorExpr); ok {
		// Check the selector is "Varying"
		if sel.Sel == nil || sel.Sel.Value != "Varying" {
			return false
		}

		// Check the package is "lanes"
		name, ok := sel.X.(*syntax.Name)
		if !ok {
			return false
		}

		// Look up the name to see if it's an import of the "lanes" package
		obj := check.lookup(name.Value)
		if obj == nil {
			return false
		}

		pkg, ok := obj.(*PkgName)
		if !ok {
			return false
		}

		if pkg.imported.path == "lanes" {
			// Mark the import as used since we're handling lanes.Varying directly
			check.usedPkgNames[pkg] = true
			return true
		}
		return false
	}

	// Case 2: Unqualified Varying (from within the lanes package itself)
	if name, ok := x.(*syntax.Name); ok && name.Value == "Varying" {
		if check.pkg.path == "lanes" {
			return true
		}
	}

	return false
}

// processLanesVaryingType creates an SPMDType from lanes.Varying[T] or lanes.Varying[T, N]
func (check *Checker) processLanesVaryingType(indexExpr *syntax.IndexExpr, def *TypeName) (Type, bool) {
	args := syntax.UnpackListExpr(indexExpr.Index)

	if len(args) < 1 || len(args) > 2 {
		check.errorf(indexExpr, InvalidSPMDType, "lanes.Varying requires 1 or 2 type arguments, got %d", len(args))
		return Typ[Invalid], true
	}

	// Type-check the element type (first argument)
	elem := check.varType(args[0])
	if !isValid(elem) {
		return Typ[Invalid], true
	}

	// Validate SPMD type restrictions (pointers, maps, channels)
	if err := check.validateSPMDTypeRestrictions(elem); err != "" {
		check.errorf(args[0], InvalidSPMDType, "%s", err)
		return Typ[Invalid], true
	}

	var constraint int64 = -1 // unconstrained by default

	// Handle optional second argument: constraint N
	if len(args) == 2 {
		var x operand
		check.expr(nil, &x, args[1])

		if x.mode() != constant_ {
			check.error(args[1], InvalidConstVal, "lanes.Varying constraint must be a compile-time constant")
			return Typ[Invalid], true
		}

		if !isInteger(x.typ()) {
			check.error(args[1], InvalidConstVal, "lanes.Varying constraint must be an integer constant")
			return Typ[Invalid], true
		}

		val, ok := constant.Int64Val(x.val)
		if !ok {
			check.error(args[1], InvalidConstVal, "lanes.Varying constraint out of range")
			return Typ[Invalid], true
		}

		if val == 0 {
			// lanes.Varying[T, 0] means universal constraint (like old varying[] T)
			constraint = 0
		} else if val < 1 {
			check.error(args[1], InvalidConstVal, "lanes.Varying constraint must be non-negative")
			return Typ[Invalid], true
		} else {
			constraint = val
		}

		// Validate constrained varying capacity (512 bits = 64 bytes limit)
		if constraint > 0 {
			elementSize := check.calculateTypeSize(elem)
			totalSize := constraint * elementSize

			const maxConstrainedCapacity = 64 // 512 bits = 64 bytes

			if totalSize > maxConstrainedCapacity {
				check.error(args[1], InvalidConstVal, "constrained varying capacity exceeded")
				return Typ[Invalid], true
			}
		}
	}

	// Create the SPMD type (always varying - uniform is implicit via regular Go types)
	typ := NewVaryingConstrained(elem, constraint)

	// Track unconstrained varying element sizes for lane count computation
	if globalSPMDInfo.inSPMDLoop && !typ.IsConstrained() {
		elemSize := check.getTypeSize(typ.elem)
		globalSPMDInfo.varyingElemSizes = append(globalSPMDInfo.varyingElemSizes, elemSize)
	}

	// Set the type on def if provided (for type declarations)
	if def != nil {
		if named := asNamed(def.typ); named != nil {
			named.fromRHS = typ
		} else if alias, ok := def.typ.(*Alias); ok {
			alias.fromRHS = typ
		}
	}

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
	switch t := t.Underlying().(type) {
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
	switch t := t.Underlying().(type) {
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

