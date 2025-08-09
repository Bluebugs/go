//go:build !goexperiment.spmd

// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

// IsUniformToken stub for when SPMD experiment is disabled.
func IsUniformToken(t token) bool {
	return false
}

// IsVaryingToken stub for when SPMD experiment is disabled.
func IsVaryingToken(t token) bool {
	return false
}