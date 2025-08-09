//go:build goexperiment.spmd

// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2


// SPMDQualifier represents the SPMD type qualifier (uniform or varying).
type SPMDQualifier uint8

const (
	UniformQualifier SPMDQualifier = iota
	VaryingQualifier
)

// SPMDType represents a uniform or varying qualified type.
type SPMDType struct {
	qualifier  SPMDQualifier
	constraint int64 // -1 for no constraint, 0 for universal ([]), >0 for numeric constraint
	elem       Type
}

// NewUniform returns a new uniform type for the given element type.
func NewUniform(elem Type) *SPMDType {
	return &SPMDType{qualifier: UniformQualifier, constraint: -1, elem: elem}
}

// NewVarying returns a new varying type for the given element type.
func NewVarying(elem Type) *SPMDType {
	return &SPMDType{qualifier: VaryingQualifier, constraint: -1, elem: elem}
}

// NewVaryingConstrained returns a new constrained varying type for the given element type and constraint.
// constraint: -1 for no constraint, 0 for universal ([]), >0 for numeric constraint
func NewVaryingConstrained(elem Type, constraint int64) *SPMDType {
	return &SPMDType{qualifier: VaryingQualifier, constraint: constraint, elem: elem}
}

// Qualifier returns the SPMD qualifier (uniform or varying).
func (s *SPMDType) Qualifier() SPMDQualifier { return s.qualifier }

// IsUniform reports whether the type is uniform.
func (s *SPMDType) IsUniform() bool { return s.qualifier == UniformQualifier }

// IsVarying reports whether the type is varying.
func (s *SPMDType) IsVarying() bool { return s.qualifier == VaryingQualifier }

// Constraint returns the varying constraint.
// -1 for no constraint, 0 for universal ([]), >0 for numeric constraint.
func (s *SPMDType) Constraint() int64 { return s.constraint }

// IsConstrained reports whether the varying type has a constraint.
func (s *SPMDType) IsConstrained() bool { return s.constraint >= 0 }

// IsUniversalConstrained reports whether the varying type has a universal constraint ([]).
func (s *SPMDType) IsUniversalConstrained() bool { return s.constraint == 0 }

// Elem returns the element type of the SPMD type.
func (s *SPMDType) Elem() Type { return s.elem }

// Underlying returns the underlying type of the SPMD type.
// For SPMD types, the underlying type is the element type.
func (s *SPMDType) Underlying() Type { return s.elem.Underlying() }

// String returns a string representation of the SPMD type.
func (s *SPMDType) String() string { return TypeString(s, nil) }

// ----------------------------------------------------------------------------
// Type compatibility and conversion utilities for SPMD types

// IsUniformType reports whether t is a uniform type.
func IsUniformType(t Type) bool {
	if s, ok := t.(*SPMDType); ok {
		return s.IsUniform()
	}
	return false
}

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

// CanAssign reports whether a value of type src can be assigned to a variable of type dst
// according to SPMD assignment rules.
// - varying can be assigned to varying of the same underlying type
// - uniform can be assigned to uniform of the same underlying type
// - uniform can be assigned to varying (broadcast)
// - varying cannot be assigned to uniform
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