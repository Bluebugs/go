// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file provides token checking utilities for SPMD types.
// They will only identify SPMD types if the SPMD experiment is enabled at runtime

package syntax

import "internal/buildcfg"

// IsUniformToken reports whether t is the uniform keyword token.
func IsUniformToken(t token) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}

	return t == _Uniform
}

// IsVaryingToken reports whether t is the varying keyword token.
func IsVaryingToken(t token) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}

	return t == _Varying
}
