// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends typexpr.go to handle SPMD types in the main typInternal function.

package types

import (
	"go/ast"
	"internal/buildcfg"
	. "internal/types/errors"
)

// handleSPMDIndexExpr checks if an IndexExpr is lanes.Varying[T] or lanes.Varying[T, N]
// and creates the corresponding SPMDType. Returns (type, true) if handled.
func (check *Checker) handleSPMDIndexExpr(ix *indexedExpr, def *TypeName) (Type, bool) {
	if !buildcfg.Experiment.SPMD {
		return nil, false
	}

	if !check.isLanesVaryingExpr(ix.x) {
		return nil, false
	}

	return check.processLanesVaryingType(ix, def)
}

// isLanesVaryingExpr checks if expr resolves to lanes.Varying
func (check *Checker) isLanesVaryingExpr(x ast.Expr) bool {
	// Case 1: Package-qualified lanes.Varying (from importing packages)
	if sel, ok := x.(*ast.SelectorExpr); ok {
		// Check the selector is "Varying"
		if sel.Sel == nil || sel.Sel.Name != "Varying" {
			return false
		}

		// Check the package is "lanes"
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return false
		}

		// Look up the name to see if it's an import of the "lanes" package
		obj := check.lookup(ident.Name)
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
	if ident, ok := x.(*ast.Ident); ok && ident.Name == "Varying" {
		if check.pkg.path == "lanes" {
			return true
		}
	}

	return false
}

// processLanesVaryingType creates an SPMDType from lanes.Varying[T]
func (check *Checker) processLanesVaryingType(ix *indexedExpr, def *TypeName) (Type, bool) {
	args := ix.indices // Already unpacked

	if len(args) != 1 {
		check.errorf(ix.orig, InvalidSPMDType, "lanes.Varying takes exactly one type argument, got %d", len(args))
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

	// Create the SPMD type (always varying - uniform is implicit via regular Go types)
	typ := NewVarying(elem)

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

