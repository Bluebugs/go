// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends call.go to add SPMD-specific call expression validation.

package types

import (
	"go/ast"
	"internal/buildcfg"
	. "internal/types/errors"
)

// validateSPMDFunctionCall validates SPMD-specific function call restrictions
func (check *Checker) validateSPMDFunctionCall(call *ast.CallExpr, x *operand) {
	if !buildcfg.Experiment.SPMD {
		return
	}

	if check.isLanesIndexCall(call) {
		check.validateLanesIndexContext(call)
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
