// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/ast"
	"go/token"
	"internal/buildcfg"
	. "internal/types/errors"
)

// SPMDControlFlowInfo tracks SPMD control flow context following ISPC approach
type SPMDControlFlowInfo struct {
	inSPMDLoop         bool
	varyingDepth       int
	maskAltered        bool
	hasVaryingParams   bool
	varyingElemSizes   []int64
	effectiveLaneCount int64
}

const simd128CapacityBytes = 16

func (check *Checker) validateSPMDFunctionSignature(fdecl *ast.FuncDecl, sig *Signature) {
	if !buildcfg.Experiment.SPMD {
		return
	}
	hasVaryingParams := false
	if sig.params != nil {
		for _, param := range sig.params.vars {
			if spmdType, ok := param.typ.(*SPMDType); ok && spmdType.IsVarying() {
				hasVaryingParams = true
				break
			}
		}
	}
	if hasVaryingParams && fdecl.Name.Name != "" {
		if token.IsExported(fdecl.Name.Name) {
			pkgName := check.pkg.name
			if pkgName != "lanes" && pkgName != "reduce" {
				check.error(fdecl, InvalidSPMDFunc, "public functions cannot have varying parameters")
			}
		}
	}
	if !hasVaryingParams && fdecl.Name.Name != "" && token.IsExported(fdecl.Name.Name) {
		if sig.results != nil {
			for _, result := range sig.results.vars {
				if spmdType, ok := result.typ.(*SPMDType); ok && spmdType.IsVarying() {
					pkgName := check.pkg.name
					if pkgName != "lanes" && pkgName != "reduce" {
						check.error(fdecl, InvalidSPMDFunc, "public functions cannot return varying types")
						return
					}
				}
			}
		}
	}
	if hasVaryingParams && fdecl.Body != nil {
		if check.hasGoForInBody(fdecl.Body) {
			check.error(fdecl, InvalidSPMDFunc, "functions with varying parameters cannot contain go for loops")
		}
	}
}

func (check *Checker) hasGoForInBody(body *ast.BlockStmt) bool {
	hasGoFor := false
	ast.Inspect(body, func(n ast.Node) bool {
		if rangeStmt, ok := n.(*ast.RangeStmt); ok && rangeStmt.IsSpmd {
			hasGoFor = true
			return false
		}
		return true
	})
	return hasGoFor
}

func (check *Checker) laneCountForType(elem Type) int64 {
	elemSize := check.getTypeSize(elem)
	if elemSize <= 0 {
		return 4
	}
	lc := int64(simd128CapacityBytes) / elemSize
	if lc <= 0 {
		return 1
	}
	return lc
}

func (check *Checker) calculateVaryingTypeCapacity(spmdType *SPMDType) int64 {
	elementSize := check.getTypeSize(spmdType.elem)
	return elementSize * check.laneCountForType(spmdType.elem)
}

func (check *Checker) computeEffectiveLaneCount(info *SPMDControlFlowInfo) int64 {
	if len(info.varyingElemSizes) == 0 {
		return int64(simd128CapacityBytes) / 4
	}
	maxElemSize := int64(0)
	for _, size := range info.varyingElemSizes {
		if size > maxElemSize {
			maxElemSize = size
		}
	}
	if maxElemSize <= 0 {
		return 4
	}
	return int64(simd128CapacityBytes) / maxElemSize
}

func (check *Checker) computeFunctionLaneCount(sig *Signature) int64 {
	if sig.params == nil {
		return 0
	}
	maxElemSize := int64(0)
	found := false
	for _, param := range sig.params.vars {
		if spmdType, ok := param.typ.(*SPMDType); ok && spmdType.IsVarying() {
			elemSize := check.getTypeSize(spmdType.elem)
			if elemSize > maxElemSize {
				maxElemSize = elemSize
				found = true
			}
		}
	}
	if !found {
		return 0
	}
	return int64(simd128CapacityBytes) / maxElemSize
}

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
		case Int, Uint, Uintptr, UnsafePointer:
			// Platform-dependent sizes: use the configured Sizes
			// to get the correct value for the target architecture
			// (e.g., 4 bytes on WASM32, 8 bytes on amd64).
			return check.conf.sizeof(typ)
		default:
			return 8
		}
	case *Array:
		return check.getTypeSize(t.elem) * t.len
	case *Slice:
		return check.conf.sizeof(typ)
	case *Pointer:
		return check.conf.sizeof(typ)
	default:
		return check.conf.sizeof(typ)
	}
}
