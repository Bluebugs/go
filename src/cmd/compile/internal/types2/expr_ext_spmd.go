// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends expr.go to handle SPMD expression type propagation.

package types2

import (
	"cmd/compile/internal/syntax"
	"internal/buildcfg"
)

// handleSPMDComparison handles comparison operations for SPMD types.
// Returns true if SPMD handling was applied, false otherwise.
func (check *Checker) handleSPMDComparison(x *operand, y *operand, op syntax.Operator) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}

	// Check if either operand is varying
	xVarying := check.isVaryingOperand(x)
	yVarying := check.isVaryingOperand(y)

	// If neither operand is varying, use regular comparison logic
	if !xVarying && !yVarying {
		return false
	}

	// At least one operand is varying, so result should be varying bool
	// Set the result type to varying bool
	x.typ_ = NewVarying(Typ[Bool])
	return true
}

// isVaryingOperand checks if an operand has a varying type
func (check *Checker) isVaryingOperand(x *operand) bool {
	if x.typ() == nil {
		return false
	}

	// Check if the type is explicitly an SPMD varying type
	if spmdType, ok := x.typ().(*SPMDType); ok {
		return spmdType.IsVarying()
	}

	return false
}

// handleSPMDBinaryExpr handles binary expressions for SPMD types.
// Returns true if SPMD handling was applied, false otherwise.
func (check *Checker) handleSPMDBinaryExpr(x *operand, y *operand, op syntax.Operator) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}

	// Check if either operand is varying
	xVarying := check.isVaryingOperand(x)
	yVarying := check.isVaryingOperand(y)

	// If neither operand is varying, use regular binary logic
	if !xVarying && !yVarying {
		return false
	}

	// For arithmetic operations with at least one varying operand, result is varying
	if isArithmetic(op) {
		// Determine result element type - should be compatible with both operands
		var elemType Type

		if xVarying {
			if spmdType, ok := x.typ().(*SPMDType); ok {
				elemType = spmdType.elem
			}
		}

		if yVarying && elemType == nil {
			if spmdType, ok := y.typ().(*SPMDType); ok {
				elemType = spmdType.elem
			}
		}

		// If we still don't have an element type, use the non-varying operand's type
		if elemType == nil {
			if xVarying {
				// y is uniform, use y's type as element type
				elemType = y.typ()
			} else {
				// x is uniform, use x's type as element type
				elemType = x.typ()
			}
		}

		if elemType != nil {
			x.typ_ = NewVarying(elemType)
			return true
		}
	}

	return false
}

// isArithmetic reports whether op is an arithmetic operator
func isArithmetic(op syntax.Operator) bool {
	switch op {
	case syntax.Add, syntax.Sub, syntax.Mul, syntax.Div, syntax.Rem:
		return true
	case syntax.And, syntax.Or, syntax.Xor, syntax.AndNot:
		return true // bitwise operations
	case syntax.Shl, syntax.Shr:
		return true // shift operations
	}
	return false
}

// handleSPMDIndexing handles varying type propagation for indexing expressions.
// If the index is varying, the result should also be varying.
// Note: The index expression has already been evaluated by the caller, so we need
// to pass the evaluated index operand to avoid double evaluation and error reporting.
func (check *Checker) handleSPMDIndexing(x *operand, indexOperand *operand) {
	if !buildcfg.Experiment.SPMD {
		return
	}

	// Check if the index is varying (using already-evaluated operand)
	if check.isVaryingOperand(indexOperand) {
		// If the index is varying and the result type is not already varying,
		// wrap it in a varying type
		if x.typ() != nil && !check.isVaryingType(x.typ()) {
			x.typ_ = NewVarying(x.typ())
		}
	}
}

// isVaryingType checks if a type is a varying type
func (check *Checker) isVaryingType(typ Type) bool {
	if spmdType, ok := typ.(*SPMDType); ok {
		return spmdType.IsVarying()
	}
	return false
}