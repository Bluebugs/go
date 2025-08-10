// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file handles SPMD pointer operation validations.

package types2

import (
	"cmd/compile/internal/syntax"
	"internal/buildcfg"
)

// validateSPMDPointerOperation validates pointer operations with SPMD types
// Returns (handled, valid, errorMsg) where:
// - handled: true if this was an SPMD-related operation
// - valid: true if the operation is valid
// - errorMsg: error message if operation is invalid
func (check *Checker) validateSPMDPointerOperation(x *operand, op syntax.Operator, operandExpr syntax.Expr) (handled, valid bool, errorMsg string) {
	if !buildcfg.Experiment.SPMD {
		return false, false, ""
	}

	switch op {
	case syntax.And: // Address operation (&)
		return check.validateSPMDAddressOperation(x, operandExpr)
	}

	return false, false, ""
}

// validateSPMDAddressOperation validates taking address of SPMD types
func (check *Checker) validateSPMDAddressOperation(x *operand, operandExpr syntax.Expr) (handled, valid bool, errorMsg string) {
	// Only reject if we're taking address of a varying variable directly (not an indexed expression)
	// Allow: &data[i] where i is varying (valid scatter/gather pattern)
	// Reject: &data where data is varying variable (invalid)
	
	if spmdType, ok := x.typ.(*SPMDType); ok && spmdType.qualifier == VaryingQualifier {
		// Check if this is a direct variable reference (not an indexed expression)
		if name, ok := operandExpr.(*syntax.Name); ok && name != nil {
			// This is taking address of a varying variable directly - forbidden
			return true, false, "cannot take address of varying variable"
		}
		// Otherwise, this is likely an indexed expression like data[i] - allow it
	}

	// Check for potential out-of-bounds access with varying pointer
	if indexExpr, ok := operandExpr.(*syntax.IndexExpr); ok {
		var idx operand
		check.expr(nil, &idx, indexExpr.Index)
		
		// Check if index expression involves multiplication that could cause out-of-bounds access
		if binExpr, ok := indexExpr.Index.(*syntax.Operation); ok && binExpr.Op == syntax.Mul {
			// Check if any operand is varying
			var left, right operand
			check.expr(nil, &left, binExpr.X)
			check.expr(nil, &right, binExpr.Y)
			
			// If either operand is varying, this could be out of bounds
			leftVarying := false
			rightVarying := false
			if spmdType, ok := left.typ.(*SPMDType); ok && spmdType.IsVarying() {
				leftVarying = true
			}
			if spmdType, ok := right.typ.(*SPMDType); ok && spmdType.IsVarying() {
				rightVarying = true
			}
			
			if leftVarying || rightVarying {
				return true, false, "potential out-of-bounds access with varying pointer"
			}
		}
	}

	return false, false, ""
}

// validateSPMDPointerArithmetic validates pointer arithmetic operations with varying pointers
func (check *Checker) validateSPMDPointerArithmetic(x *operand, op syntax.Operator) (handled, valid bool, errorMsg string) {
	if !buildcfg.Experiment.SPMD {
		return false, false, ""
	}

	// Check if operand is a varying pointer
	if spmdType, ok := x.typ.(*SPMDType); ok && spmdType.qualifier == VaryingQualifier {
		if ptr, ok := spmdType.elem.(*Pointer); ok && ptr != nil {
			// This is a varying pointer - arithmetic operations like ++ are not supported
			return true, false, "varying pointer arithmetic not supported in this context"
		}
	}

	return false, false, ""
}

// validateSPMDPointerAssignment validates pointer assignments between SPMD types
func (check *Checker) validateSPMDPointerAssignment(dstType, srcType Type) (handled, valid bool, errorMsg string) {
	if !buildcfg.Experiment.SPMD {
		return false, false, ""
	}

	dstSPMD, dstIsSPMD := dstType.(*SPMDType)
	srcSPMD, srcIsSPMD := srcType.(*SPMDType)

	// Case 1: Assigning varying pointer to uniform pointer (forbidden)
	if dstIsSPMD && srcIsSPMD {
		// Check if both are pointer types
		dstPtr, dstIsPtr := dstSPMD.elem.(*Pointer)
		srcPtr, srcIsPtr := srcSPMD.elem.(*Pointer)
		
		if dstIsPtr && srcIsPtr && dstPtr != nil && srcPtr != nil {
			// Varying pointer to uniform pointer assignment is forbidden
			if srcSPMD.qualifier == VaryingQualifier && dstSPMD.qualifier == UniformQualifier {
				return true, false, "cannot assign varying pointer to uniform variable"
			}
		}
	}

	// Case 2: Assigning varying pointer to regular pointer type (also forbidden)
	if !dstIsSPMD && srcIsSPMD {
		if _, dstIsPtr := dstType.(*Pointer); dstIsPtr {
			if srcSPMD.qualifier == VaryingQualifier {
				if _, srcIsPtr := srcSPMD.elem.(*Pointer); srcIsPtr {
					return true, false, "cannot assign varying pointer to uniform variable"
				}
			}
		}
	}

	return false, false, ""
}

// validateSPMDTypeSwitchCases checks for SPMD type switch restrictions
func (check *Checker) validateSPMDTypeSwitchCases(cases []syntax.Expr, hasDefault bool) (bool, string) {
	if !buildcfg.Experiment.SPMD {
		return false, ""
	}

	// Check if any case involves varying types
	hasVaryingTypes := false
	for _, e := range cases {
		if e != nil {
			T := check.varType(e)
			if T != nil {
				if spmdType, ok := T.(*SPMDType); ok && spmdType.IsVarying() {
					hasVaryingTypes = true
					break
				}
			}
		}
	}

	// If we have varying types and a default case, validate it
	if hasVaryingTypes && hasDefault {
		return true, "varying types in type switch must be handled explicitly"
	}

	return false, ""
}

// validateSPMDVaryingPointerAccess checks for potential out-of-bounds varying pointer access
func (check *Checker) validateSPMDVaryingPointerAccess(x *operand, index syntax.Expr) (bool, string) {
	if !buildcfg.Experiment.SPMD {
		return false, ""
	}

	// Check if we're taking address of an array element with varying index
	if x.mode == variable {
		if indexExpr, ok := index.(*syntax.IndexExpr); ok {
			var idx operand
			check.expr(nil, &idx, indexExpr.Index)
			
			// Check if index is a varying expression that could cause out-of-bounds access
			if idx.typ != nil {
				if spmdType, ok := idx.typ.(*SPMDType); ok && spmdType.IsVarying() {
					// This is a simplified check - in a real implementation, we would need
					// more sophisticated analysis to determine if the access is truly unsafe
					return true, "potential out-of-bounds access with varying pointer"
				}
			}
		}
	}

	return false, ""
}