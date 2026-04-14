// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends call.go to add SPMD-specific call expression validation.

package types

import (
	"go/ast"
	"go/constant"
	"go/token"
	"internal/buildcfg"
	. "internal/types/errors"
)

const (
	spmdFeatureDisabled   = 1 << iota // spmd experiment flag not enabled
	spmdInvalidContext                // call in invalid SPMD context
	spmdGroupSizeMismatch             // groupSize doesn't divide lane count
)

func (check *Checker) getSPMDInfo() *SPMDControlFlowInfo {
	return &check.spmdInfo
}

// validateSPMDFunctionCall validates SPMD-specific function call restrictions
func (check *Checker) validateSPMDFunctionCall(call *ast.CallExpr, x *operand) {
	if !buildcfg.Experiment.SPMD {
		return
	}

	if check.isLanesIndexCall(call) {
		check.validateLanesIndexContext(call)
	}

	if check.isCrossLaneWithinCall(call) {
		check.validateCrossLaneWithinCall(call)
	}
}

// isLanesIndexCall checks if this is a call to lanes.Index()
func (check *Checker) isLanesIndexCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "Index" {
		return false
	}

	if name, ok := sel.X.(*ast.Ident); ok {
		if obj := check.lookup(name.Name); obj != nil {
			if pkg, ok := obj.(*PkgName); ok {
				return pkg.imported.name == "lanes"
			}
		}
	}

	return false
}

// validateLanesIndexContext validates that lanes.Index() is called in SPMD context
func (check *Checker) validateLanesIndexContext(call *ast.CallExpr) {
	if check.spmdInfo.inSPMDLoop {
		return
	}

	if check.isInSPMDFunction() {
		return
	}

	check.error(call, InvalidSPMDCall, "lanes.Index() can only be called in SPMD context")
}

// isCrossLaneWithinCall checks if this is a call to lanes.*Within functions
func (check *Checker) isCrossLaneWithinCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	switch sel.Sel.Name {
	case "RotateWithin", "SwizzleWithin", "ShiftLeftWithin", "ShiftRightWithin":
		// Valid names
	default:
		return false
	}

	if name, ok := sel.X.(*ast.Ident); ok {
		if obj := check.lookup(name.Name); obj != nil {
			if pkg, ok := obj.(*PkgName); ok {
				return pkg.imported.name == "lanes"
			}
		}
	}

	return false
}

// validateCrossLaneWithinCall validates that *Within functions are called in valid SPMD context
func (check *Checker) validateCrossLaneWithinCall(call *ast.CallExpr) {
	sel := call.Fun.(*ast.SelectorExpr)
	funcName := "lanes." + sel.Sel.Name

	// Check if in SPMD context (either in SPMD loop or SPMD function)
	inSPMDContext := check.isInSPMDLoop() || check.isInSPMDFunction()
	if !inSPMDContext {
		check.errorf(call, InvalidSPMDCall, "%s can only be called in SPMD context (go for loop or SPMD function)", funcName)
		return
	}

	// Get the groupSize argument (third argument)
	if len(call.Args) < 3 {
		return // Will be caught by regular type checking
	}

	groupSizeArg := call.Args[2]
	groupSize := check.groupSizeFromExpr(groupSizeArg)

	if groupSize <= 0 {
		return // Will be caught by regular type checking
	}

	// Get effective lane count
	laneCount := check.getEffectiveLaneCount()

	if laneCount <= 1 {
		check.errorf(call, InvalidSPMDCall, "%s requires SIMD mode (-simd=false is not supported for this function)", funcName)
		return
	}

	if laneCount%groupSize != 0 {
		check.errorf(call, InvalidSPMDCall, "%s: groupSize (%d) must evenly divide lane count (%d)", funcName, groupSize, laneCount)
	}
}

func (check *Checker) isInSPMDLoop() bool {
	// Check if we're inside an SPMD loop by looking at the spmdInfo context
	return check.spmdInfo.inSPMDLoop
}

// groupSizeFromExpr extracts the groupSize constant from an expression
func (check *Checker) groupSizeFromExpr(expr ast.Expr) int64 {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.INT {
			if n, ok := constant.Int64Val(constant.MakeFromLiteral(e.Value, token.INT, 0)); ok {
				return n
			}
		}
	case *ast.Ident:
		// Try to resolve constant
		if obj := check.lookup(e.Name); obj != nil {
			if con, ok := obj.(*Const); ok {
				if n, ok := constant.Int64Val(con.Val()); ok {
					return n
				}
			}
		}
	}
	return -1
}

// getEffectiveLaneCount returns the effective lane count in current SPMD context
// Returns 1 if in scalar mode (detected by simdRegisterSize() <= 1)
func (check *Checker) getEffectiveLaneCount() int64 {
	// Check if we're in scalar mode first
	if check.simdRegisterSize() <= 1 {
		return 1
	}

	// In SPMD context, compute effective lane count from varying types in scope
	if check.spmdInfo.inSPMDLoop || check.isInSPMDFunction() {
		return check.computeEffectiveLaneCount(&check.spmdInfo)
	}

	// Not in SPMD context - use default based on register size
	// This handles the case where *Within functions are called outside SPMD context
	// (which should be caught by context validation first)
	return check.simdRegisterSize() / 4 // default assuming int32 elements
}

// isInSPMDFunction checks if we're currently inside an SPMD function
func (check *Checker) isInSPMDFunction() bool {
	if check.sig != nil {
		if check.sig.params != nil {
			for _, param := range check.sig.params.vars {
				if spmdType, ok := param.typ.(*SPMDType); ok && spmdType.IsVarying() {
					return true
				}
			}
		}
	}
	return false
}

// validateSPMDMakeType validates SPMD restrictions for make() calls
func (check *Checker) validateSPMDMakeType(arg ast.Expr, T Type) {
	if !buildcfg.Experiment.SPMD {
		return
	}
}

// spmdWrapFieldType returns the SPMD-adjusted type for a struct field access.
// When the receiver is *Varying[S] (a uniform pointer to a varying struct),
// field access should return Varying[fieldType], not fieldType.
// For all other receivers, the field type is returned unchanged.
func spmdWrapFieldType(receiverType, fieldType Type) Type {
	if !buildcfg.Experiment.SPMD {
		return fieldType
	}
	ptr, ok := receiverType.(*Pointer)
	if !ok {
		return fieldType
	}
	if _, ok := ptr.Elem().(*SPMDType); !ok {
		return fieldType
	}
	if _, alreadyVarying := fieldType.(*SPMDType); alreadyVarying {
		return fieldType
	}
	return NewVarying(fieldType)
}
