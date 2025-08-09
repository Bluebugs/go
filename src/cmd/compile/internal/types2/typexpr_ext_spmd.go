//go:build goexperiment.spmd

// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends typexpr.go to handle SPMD types in the main typInternal function.

package types2

import (
	"cmd/compile/internal/syntax"
	"go/constant"
	. "internal/types/errors"
)

// handleSPMDTypeExpr handles SPMD type expressions in the typInternal method.
// Returns the resulting type and true if handled, or nil and false if not an SPMD type.
func (check *Checker) handleSPMDTypeExpr(e syntax.Expr, def *TypeName) (Type, bool) {
	spmdExpr, ok := e.(*syntax.SPMDType)
	if !ok {
		return nil, false
	}

	var qualifier SPMDQualifier
	var constraint int64 = -1

	// Determine the qualifier
	if syntax.IsUniformToken(spmdExpr.Qualifier) {
		qualifier = UniformQualifier
	} else if syntax.IsVaryingToken(spmdExpr.Qualifier) {
		qualifier = VaryingQualifier
	} else {
		check.error(spmdExpr, InvalidSyntaxTree, "invalid SPMD qualifier")
		return Typ[Invalid], true
	}

	// Handle varying constraints
	if qualifier == VaryingQualifier && spmdExpr.Constraint != nil {
		// Parse the constraint expression
		if spmdExpr.Constraint == nil {
			// varying[] - universal constraint
			constraint = 0
		} else {
			// varying[n] - numeric constraint
			var x operand
			check.expr(nil, &x, spmdExpr.Constraint)
			if x.mode != constant_ {
				check.error(spmdExpr.Constraint, InvalidConstVal, "constraint must be a constant")
				return Typ[Invalid], true
			}

			if !isInteger(x.typ) {
				check.error(spmdExpr.Constraint, InvalidConstVal, "constraint must be an integer constant")
				return Typ[Invalid], true
			}

			val, ok := constant.Int64Val(x.val)
			if !ok {
				check.error(spmdExpr.Constraint, InvalidConstVal, "constraint out of range")
				return Typ[Invalid], true
			}

			if val < 1 {
				check.error(spmdExpr.Constraint, InvalidConstVal, "constraint must be positive")
				return Typ[Invalid], true
			}

			constraint = val
		}
	}

	// Type-check the element type
	elem := check.varType(spmdExpr.Elem)
	if !isValid(elem) {
		return Typ[Invalid], true
	}

	// Validate pointer-to-varying restrictions
	if err := check.validateSPMDPointerTypes(elem); err != "" {
		check.errorf(spmdExpr.Elem, InvalidSPMDType, "%s", err)
		return Typ[Invalid], true
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

// validateSPMDPointerTypes validates pointer type restrictions for SPMD types.
// Returns an error message if the type is invalid, empty string if valid.
func (check *Checker) validateSPMDPointerTypes(typ Type) string {
	switch t := typ.(type) {
	case *Pointer:
		// Check if pointer points to varying type: *varying T is forbidden
		if spmdType, ok := t.base.(*SPMDType); ok && spmdType.IsVarying() {
			return "pointer to varying type not supported"
		}
		// Recursively check the pointed-to type
		return check.validateSPMDPointerTypes(t.base)
		
	case *Array:
		// Check array element type recursively
		return check.validateSPMDPointerTypes(t.elem)
		
	case *Slice:
		// Check slice element type recursively  
		return check.validateSPMDPointerTypes(t.elem)
		
	case *SPMDType:
		// Check the underlying element type
		return check.validateSPMDPointerTypes(t.elem)
		
	default:
		// Other types are fine
		return ""
	}
}