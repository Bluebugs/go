// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends expr.go to handle SPMD expression type propagation.

package types

import (
	"go/token"
	"internal/buildcfg"
)

// handleSPMDComparison handles comparison operations for SPMD types.
// Returns true if SPMD handling was applied, false otherwise.
func (check *Checker) handleSPMDComparison(x *operand, y *operand, op token.Token) bool {
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
func (check *Checker) handleSPMDBinaryExpr(x *operand, y *operand, op token.Token) bool {
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
func isArithmetic(op token.Token) bool {
	switch op {
	case token.ADD, token.SUB, token.MUL, token.QUO, token.REM:
		return true
	case token.AND, token.OR, token.XOR, token.AND_NOT:
		return true // bitwise operations
	case token.SHL, token.SHR:
		return true // shift operations
	}
	return false
}

// handleSPMDAddrOf handles &Varying[T] producing Varying[*T] (per-lane pointer vector).
// Returns true if SPMD handling was applied and x has been updated, false otherwise.
func (check *Checker) handleSPMDAddrOf(x *operand) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}
	spmdType, ok := x.typ().(*SPMDType)
	if !ok || !spmdType.IsVarying() {
		return false
	}
	x.mode_ = value
	x.typ_ = NewVarying(&Pointer{base: spmdType.Elem()})
	return true
}

// handleSPMDIndirect handles *Varying[*T] producing Varying[T] (per-lane scatter/gather).
// When dereferencing a varying pointer vector, each lane independently dereferences
// its own pointer, producing a varying value. Returns true if SPMD handling was
// applied and x has been updated, false otherwise.
func (check *Checker) handleSPMDIndirect(x *operand) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}
	spmdType, ok := x.typ().(*SPMDType)
	if !ok || !spmdType.IsVarying() {
		return false
	}
	ptr, ok := spmdType.Elem().(*Pointer)
	if !ok {
		return false
	}
	// *Varying[*T] → Varying[T]: per-lane dereference (scatter on LHS, gather on RHS).
	x.mode_ = variable
	x.typ_ = NewVarying(ptr.base)
	return true
}

// handleSPMDIndexing handles varying type propagation for indexing expressions.
// If the index is varying, the result should also be varying.
func (check *Checker) handleSPMDIndexing(x *operand, indexType Type) {
	if !buildcfg.Experiment.SPMD {
		return
	}

	// Check if the index is varying
	if IsVaryingType(indexType) {
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
