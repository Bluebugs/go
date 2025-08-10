// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends call.go to add SPMD-specific call expression validation.

package types2

import (
	"cmd/compile/internal/syntax"
	"internal/buildcfg"
	. "internal/types/errors"
)

// validateSPMDFunctionCall validates SPMD-specific function call restrictions
func (check *Checker) validateSPMDFunctionCall(call *syntax.CallExpr, x *operand) {
	if !buildcfg.Experiment.SPMD {
		return
	}

	// Check for lanes.Index() context restrictions
	if check.isLanesIndexCall(call) {
		check.validateLanesIndexContext(call)
	}
}

// isLanesIndexCall checks if this is a call to lanes.Index()
func (check *Checker) isLanesIndexCall(call *syntax.CallExpr) bool {
	// Check if this is a selector expression (pkg.func)
	sel, ok := call.Fun.(*syntax.SelectorExpr)
	if !ok {
		return false
	}

	// Check if the selector is "Index"
	if sel.Sel.Value != "Index" {
		return false
	}

	// Check if the package is "lanes"
	if name, ok := sel.X.(*syntax.Name); ok {
		if obj := check.lookup(name.Value); obj != nil {
			if pkg, ok := obj.(*PkgName); ok {
				return pkg.imported.name == "lanes"
			}
		}
	}

	return false
}

// validateLanesIndexContext validates that lanes.Index() is called in SPMD context
func (check *Checker) validateLanesIndexContext(call *syntax.CallExpr) {
	// lanes.Index() can only be called within SPMD contexts:
	// 1. Inside go for loops
	// 2. Inside SPMD functions (functions with varying parameters)
	
	// Check if we're in a go for loop (SPMD context)
	if globalSPMDInfo.inSPMDLoop {
		return // Valid: inside go for loop
	}

	// Check if we're in an SPMD function
	if check.isInSPMDFunction() {
		return // Valid: inside SPMD function
	}

	// Invalid: lanes.Index() called outside SPMD context
	check.error(call, InvalidSPMDCall, "lanes.Index() can only be called in SPMD context")
}

// isInSPMDFunction checks if we're currently inside an SPMD function
func (check *Checker) isInSPMDFunction() bool {
	// Check if current function signature has varying parameters
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
func (check *Checker) validateSPMDMakeType(arg syntax.Expr, T Type) {
	if !buildcfg.Experiment.SPMD {
		return
	}

	// Map validation is now handled by validateSPMDTypeRestrictions in type validation
	// No additional validation needed here
}