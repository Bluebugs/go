// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.spmd

// Package reduce provides reduction operations for SPMD programming.
// These functions convert varying values to uniform values through various reduction operations.
package reduce

import "lanes"

// FIXME: This is a stub implementation for Phase 1.4 type system validation.
// All functions will panic at runtime until Phase 2+ implementation.
//
// NOTE: Avoiding generic function calls due to compiler bug - generic SPMD functions
// calling other generic SPMD functions cause "unreachable" panic.

// Add reduces varying values to their uniform sum.
//go:noinline
func Add[T numeric](data lanes.Varying[T]) T {
	// FIXME: Implement in Phase 2 - SIMD reduction sum
	panic("reduce.Add() not implemented - stub for Phase 1.4 type validation")
}

// All returns true if all lane values are true.
//go:noinline
func All(data lanes.Varying[bool]) bool {
	// FIXME: Implement in Phase 2 - SIMD reduction AND
	panic("reduce.All() not implemented - stub for Phase 1.4 type validation")
}

// Any returns true if any lane value is true.
//go:noinline
func Any(data lanes.Varying[bool]) bool {
	// FIXME: Implement in Phase 2 - SIMD reduction OR
	panic("reduce.Any() not implemented - stub for Phase 1.4 type validation")
}

// Max reduces varying values to their uniform maximum.
//go:noinline
func Max[T ordered](data lanes.Varying[T]) T {
	// FIXME: Implement in Phase 2 - SIMD reduction max
	panic("reduce.Max() not implemented - stub for Phase 1.4 type validation")
}

// Min reduces varying values to their uniform minimum.
//go:noinline
func Min[T ordered](data lanes.Varying[T]) T {
	// FIXME: Implement in Phase 2 - SIMD reduction min
	panic("reduce.Min() not implemented - stub for Phase 1.4 type validation")
}

// Or performs bitwise OR reduction across lanes.
//go:noinline
func Or[T integer](data lanes.Varying[T]) T {
	// FIXME: Implement in Phase 2 - SIMD bitwise OR reduction
	panic("reduce.Or() not implemented - stub for Phase 1.4 type validation")
}

// And performs bitwise AND reduction across lanes.
//go:noinline
func And[T integer](data lanes.Varying[T]) T {
	// FIXME: Implement in Phase 2 - SIMD bitwise AND reduction
	panic("reduce.And() not implemented - stub for Phase 1.4 type validation")
}

// Xor performs bitwise XOR reduction across lanes.
//go:noinline
func Xor[T integer](data lanes.Varying[T]) T {
	// FIXME: Implement in Phase 2 - SIMD bitwise XOR reduction
	panic("reduce.Xor() not implemented - stub for Phase 1.4 type validation")
}

// From converts varying values to a uniform array.
// Each lane value becomes an array element.
//go:noinline
func From[T any](data lanes.Varying[T]) []T {
	// FIXME: Implement in Phase 2 - varying to array conversion
	panic("reduce.From() not implemented - stub for Phase 1.4 type validation")
}

// Count returns the number of true values across lanes.
//go:noinline
func Count(data lanes.Varying[bool]) int {
	// FIXME: Implement in Phase 2 - SIMD population count
	panic("reduce.Count() not implemented - stub for Phase 1.4 type validation")
}

// FindFirstSet returns the index of the first true value.
//go:noinline
func FindFirstSet(data lanes.Varying[bool]) int {
	// FIXME: Implement in Phase 2 - SIMD find first set
	panic("reduce.FindFirstSet() not implemented - stub for Phase 1.4 type validation")
}

// Mask creates a mask from boolean values.
//go:noinline
func Mask(data lanes.Varying[bool]) int {
	// FIXME: Implement in Phase 2 - boolean mask creation
	panic("reduce.Mask() not implemented - stub for Phase 1.4 type validation")
}

// Mul reduces varying values to their uniform product.
//go:noinline
func Mul[T numeric](data lanes.Varying[T]) T {
	// FIXME: Implement in Phase 2 - SIMD reduction multiply
	panic("reduce.Mul() not implemented - stub for Phase 1.4 type validation")
}

// NOTE: Sum alias removed due to compiler bug with generic SPMD function calls.
// Users should call Add() directly instead of Sum().

// Type constraints for generic functions
type numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string
}
