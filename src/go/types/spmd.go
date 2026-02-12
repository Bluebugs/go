// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// SPMDQualifier represents the SPMD type qualifier (uniform or varying).
type SPMDQualifier uint8

const (
	UniformQualifier SPMDQualifier = iota
	VaryingQualifier
)

// SPMDType represents a varying qualified type (lanes.Varying[T] or lanes.Varying[T, N]).
type SPMDType struct {
	qualifier  SPMDQualifier
	constraint int64
	elem       Type
}

func (s *SPMDType) Qualifier() SPMDQualifier { return s.qualifier }
func (s *SPMDType) IsUniform() bool          { return s.qualifier == UniformQualifier }
func (s *SPMDType) IsVarying() bool          { return s.qualifier == VaryingQualifier }
func (s *SPMDType) Constraint() int64        { return s.constraint }
func (s *SPMDType) Elem() Type               { return s.elem }
func (s *SPMDType) Underlying() Type         { return s.elem.Underlying() }
func (s *SPMDType) String() string           { return TypeString(s, nil) }
