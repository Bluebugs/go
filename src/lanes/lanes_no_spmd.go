// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !goexperiment.spmd

// Package lanes provides cross-lane operations for SPMD programming.
// This file provides the Varying type stub when SPMD experiment is disabled.
package lanes

// Varying represents a value that differs across SIMD lanes.
// When SPMD is disabled, this type exists for compilation but
// SPMD-specific functions are unavailable and will cause compile errors.
type Varying[T any] struct{ _ [0]T }
