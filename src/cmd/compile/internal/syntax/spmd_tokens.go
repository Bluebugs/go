//go:build goexperiment.spmd

// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file provides token checking utilities for SPMD types.

package syntax

// IsUniformToken reports whether t is the uniform keyword token.
func IsUniformToken(t token) bool {
	return t == _Uniform
}

// IsVaryingToken reports whether t is the varying keyword token.
func IsVaryingToken(t token) bool {
	return t == _Varying
}