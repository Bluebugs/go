// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends check.go to add SPMD support to the Checker struct.

package types2

import (
	"cmd/compile/internal/syntax"
	"internal/buildcfg"
	. "internal/types/errors"
)

// Extensions to Checker for SPMD support
type CheckerSPMDExtension struct {
	spmdInfo SPMDControlFlowInfo // SPMD control flow tracking
}

// getSPMDInfo returns the SPMD control flow info for the checker
func (check *Checker) getSPMDInfo() *SPMDControlFlowInfo {
	if !buildcfg.Experiment.SPMD {
		return nil
	}

	// For now, we'll use a global variable approach
	// In a full implementation, this would be a field in Checker
	return &globalSPMDInfo
}

// Global SPMD info (temporary solution until Checker struct can be modified)
var globalSPMDInfo SPMDControlFlowInfo

// Helper to access SPMD info
var spmdInfo *SPMDControlFlowInfo = &globalSPMDInfo

// validateSPMDFunctionSignature validates SPMD function signature constraints
func (check *Checker) validateSPMDFunctionSignature(fdecl *syntax.FuncDecl, sig *Signature) {
	if !buildcfg.Experiment.SPMD {
		return
	}

	// Check if function has varying parameters
	hasVaryingParams := false
	if sig.params != nil {
		for _, param := range sig.params.vars {
			if spmdType, ok := param.typ.(*SPMDType); ok && spmdType.IsVarying() {
				hasVaryingParams = true
				break
			}
		}
	}

	// Rule 1: Public functions cannot have varying parameters
	// Exception: Allow public functions in lanes and reduce packages
	if hasVaryingParams && fdecl.Name.Value != "" {
		// Check if function name starts with uppercase (public)
		if fdecl.Name.Value[0] >= 'A' && fdecl.Name.Value[0] <= 'Z' {
			// Allow public varying functions in lanes and reduce packages
			pkgName := check.pkg.name
			if pkgName != "lanes" && pkgName != "reduce" {
				check.error(fdecl, InvalidSPMDFunction, "public functions cannot have varying parameters")
			}
		}
	}

	// Rule 2: Public functions cannot return varying types (separate from parameter rule)
	// Exception: Allow public functions in lanes and reduce packages
	if !hasVaryingParams && fdecl.Name.Value != "" && fdecl.Name.Value[0] >= 'A' && fdecl.Name.Value[0] <= 'Z' {
		// Check if function returns varying types
		if sig.results != nil {
			for _, result := range sig.results.vars {
				if spmdType, ok := result.typ.(*SPMDType); ok && spmdType.IsVarying() {
					// Allow public varying return functions in lanes and reduce packages
					pkgName := check.pkg.name
					if pkgName != "lanes" && pkgName != "reduce" {
						check.error(fdecl, InvalidSPMDFunction, "public functions cannot return varying types")
						return
					}
				}
			}
		}
	}

	// Rule 3: Functions with varying parameters cannot contain go for loops
	if hasVaryingParams && fdecl.Body != nil {
		if check.hasGoForInSPMDFunction(fdecl.Body) {
			check.error(fdecl, InvalidSPMDFunction, "functions with varying parameters cannot contain go for loops")
		}
	}

	// Note: Return statement validation is handled in stmt.go during return statement processing
}

// SIMD register capacity validation constants
const (
	// SIMD128 capacity: 128 bits = 16 bytes
	simd128CapacityBytes = 16

	// Constrained varying capacity limit: 512 bits = 64 bytes
	constrainedVaryingCapacityBytes = 64

	// Default lane count for capacity calculations
	defaultLaneCount = 4

	// For PoC, be more permissive with capacity - allow larger types for testing
	// In practice, this would be configurable per target architecture
	pocCapacityMultiplier = 4 // Allow 4x SIMD128 capacity for PoC testing
)

// checkSPMDFunctionCapacity validates that varying parameters don't exceed SIMD capacity
func (check *Checker) checkSPMDFunctionCapacity(fdecl *syntax.FuncDecl, sig *Signature) {
	if !buildcfg.Experiment.SPMD {
		return
	}

	if sig.params != nil {
		for _, param := range sig.params.vars {
			if spmdType, ok := param.typ.(*SPMDType); ok && spmdType.IsVarying() {
				size := check.calculateVaryingTypeCapacity(spmdType)

				// Check individual parameter capacity - each parameter must be ≤ 16 bytes
				maxCapacity := int64(simd128CapacityBytes) // 16 bytes per parameter
				if size > maxCapacity {
					check.error(fdecl, InvalidSPMDFunction, "SPMD function parameter capacity exceeded")
					return
				}
			}
		}
	}

	// Note: No total function capacity limit - only per-parameter limit applies
}

// checkGoForCapacity validates SIMD capacity for go for loops
func (check *Checker) checkGoForCapacity(stmt *syntax.ForStmt) {
	if !buildcfg.Experiment.SPMD {
		return
	}

	// Collect all varying variables declared/used in the loop body
	capacity := check.calculateGoForCapacity(stmt.Body)

	if capacity > simd128CapacityBytes {
		check.error(stmt, InvalidSPMDFor, "SIMD register capacity exceeded")
	}
}

// calculateVaryingTypeCapacity calculates the SIMD capacity needed for a varying type
func (check *Checker) calculateVaryingTypeCapacity(spmdType *SPMDType) int64 {
	elementSize := check.getTypeSize(spmdType.elem)

	// For constrained varying, multiply by constraint
	if spmdType.IsConstrained() {
		constraintValue := spmdType.Constraint()
		if constraintValue == 0 {
			// Universal constraint ([]) - use default lane count
			return elementSize * defaultLaneCount
		}
		return elementSize * constraintValue
	}

	// For unconstrained varying, multiply by default lane count
	return elementSize * defaultLaneCount
}

// getTypeSize returns the size in bytes of a type
func (check *Checker) getTypeSize(typ Type) int64 {
	switch t := typ.Underlying().(type) {
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
			return 8 // assuming 64-bit architecture
		default:
			return 8 // conservative default
		}
	case *Array:
		elemSize := check.getTypeSize(t.elem)
		return elemSize * t.len
	case *Slice:
		return 24 // slice header size (3 * 8 bytes)
	case *Pointer:
		return 8 // pointer size
	default:
		return 8 // conservative default
	}
}

// calculateGoForCapacity calculates total capacity needed for all varying variables in a go for loop
func (check *Checker) calculateGoForCapacity(body syntax.Stmt) int64 {
	// This is a simplified implementation
	// In practice, we'd need to walk the AST and collect all varying variable declarations
	// For now, return a placeholder that allows basic validation
	return 0 // TODO: Implement AST walking to collect varying variables
}
