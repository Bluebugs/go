// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// SPMDQualifier represents the SPMD type qualifier (uniform or varying).
type SPMDQualifier uint8

const (
	UniformQualifier SPMDQualifier = iota
	VaryingQualifier
)

// SPMDType represents a varying qualified type (lanes.Varying[T]).
//
// The optional `lanes` field is set by the SSA predication pass to a
// non-zero canonical lane count when the value is materialized inside
// an SPMD loop scope (e.g., a Varying[int] inside a `go for` over
// []float64 carries lanes=2 on WASM SIMD128 to match the loop's
// iteration width). The type-checker only ever produces width-free
// instances (lanes=0); width-fixed instances exist only post
// type-checking, mutated in place via the SSA predication pass.
//
// TinyGo's getLLVMType reads Lanes() during Varying[T] materialization
// to emit the LLVM vector at the canonical width. When lanes=0
// (abstract), TinyGo falls back to its existing element-natural /
// function-min derivation.
type SPMDType struct {
	qualifier SPMDQualifier
	elem      Type
	lanes     int // 0 = abstract; > 0 = width-fixed (SSA predication pass)
}

// NewVarying returns a new abstract varying type for the given element
// type. The type-checker uses this constructor; its Lanes() returns 0.
func NewVarying(elem Type) *SPMDType {
	return &SPMDType{qualifier: VaryingQualifier, elem: elem}
}

// NewVaryingWithLanes returns a new width-fixed varying type. Used by
// the SSA predication pass to mutate in-loop Varying values' types to
// carry the surrounding loop's canonical lane count. TinyGo reads
// Lanes() during getLLVMType to materialize the LLVM vector at the
// right width. Type-checker callers should not use this constructor —
// they produce abstract types via NewVarying.
func NewVaryingWithLanes(elem Type, lanes int) *SPMDType {
	return &SPMDType{qualifier: VaryingQualifier, elem: elem, lanes: lanes}
}

// Qualifier returns the SPMD qualifier (uniform or varying).
func (s *SPMDType) Qualifier() SPMDQualifier { return s.qualifier }

// IsUniform reports whether the type is uniform.
func (s *SPMDType) IsUniform() bool { return s.qualifier == UniformQualifier }

// IsVarying reports whether the type is varying.
func (s *SPMDType) IsVarying() bool { return s.qualifier == VaryingQualifier }

// Elem returns the element type of the SPMD type.
func (s *SPMDType) Elem() Type { return s.elem }

// Lanes returns the canonical lane count of a width-fixed SPMD type.
// Returns 0 for abstract (type-checker-produced) instances. Set by the
// SSA predication pass via NewVaryingWithLanes; read by TinyGo's
// getLLVMType to materialize the LLVM vector at the right width.
func (s *SPMDType) Lanes() int { return s.lanes }

// Underlying returns the underlying type of the SPMD type.
// For SPMD types, the underlying type is the element type.
func (s *SPMDType) Underlying() Type { return s.elem.Underlying() }

// String returns a string representation of the SPMD type.
func (s *SPMDType) String() string { return TypeString(s, nil) }

// ----------------------------------------------------------------------------
// Type compatibility and conversion utilities for SPMD types

// IsVaryingType reports whether t is a varying type.
func IsVaryingType(t Type) bool {
	if s, ok := t.(*SPMDType); ok {
		return s.IsVarying()
	}
	return false
}

// IsSPMDType reports whether t is an SPMD type (uniform or varying).
func IsSPMDType(t Type) bool {
	_, ok := t.(*SPMDType)
	return ok
}

// UnderlyingType returns the underlying type, unwrapping SPMD qualifiers.
func UnderlyingType(t Type) Type {
	if s, ok := t.(*SPMDType); ok {
		return s.Elem()
	}
	return t
}

// CanAssignSPMD reports whether a value of type src can be assigned to a variable of type dst
// according to SPMD assignment rules.
// - varying can be assigned to varying of the same underlying type
// - regular types can be assigned to varying (broadcast)
// - varying cannot be assigned to regular types (except via reduce operations)
func CanAssignSPMD(dst, src Type) bool {
	dstSPMD, dstIsSPMD := dst.(*SPMDType)
	srcSPMD, srcIsSPMD := src.(*SPMDType)

	// If neither is SPMD, use regular type compatibility
	if !dstIsSPMD && !srcIsSPMD {
		return Identical(dst, src)
	}

	// Get underlying types for comparison
	dstUnderlying := UnderlyingType(dst)
	srcUnderlying := UnderlyingType(src)

	// Check underlying type compatibility
	if !Identical(dstUnderlying, srcUnderlying) {
		return false
	}

	// SPMD assignment rules
	if dstIsSPMD && srcIsSPMD {
		// Both are SPMD types
		if dstSPMD.IsUniform() && srcSPMD.IsVarying() {
			return false // varying cannot be assigned to uniform
		}
		return true // uniform to uniform, varying to varying, uniform to varying all OK
	} else if dstIsSPMD && !srcIsSPMD {
		// dst is SPMD, src is regular Go type - treat as uniform
		return dstSPMD.IsUniform() || dstSPMD.IsVarying() // can assign to both uniform and varying
	} else if !dstIsSPMD && srcIsSPMD {
		// dst is regular Go type, src is SPMD
		return srcSPMD.IsUniform() // can only assign uniform to regular type
	}

	return false
}
