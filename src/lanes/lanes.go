// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.spmd

// Package lanes provides cross-lane operations for SPMD programming.
// These functions enable data movement and communication between SIMD lanes.
//
// IMPORTANT: All functions in this package are COMPILER BUILTINS that cannot be
// implemented in regular Go code. They must be handled specially by the compiler
// during compilation and replaced with appropriate SIMD instructions.
//
// The Go source code here serves only for:
// 1. Type checking and validation during Phase 1
// 2. Documentation of the expected API
// 3. Placeholder implementations that panic if somehow executed
package lanes

// Varying represents a value that differs across SIMD lanes (a vector value).
// The type checker special-cases this type:
// - Varying[T] is an unconstrained varying type
type Varying[T any] struct{ _ [0]T }

// PHASE 1.8: Compiler builtin declarations for SPMD lane operations.
// These functions should never execute at runtime - they must be replaced by the compiler.

// Index returns the current lane index (0 to Count-1) in SPMD context.
// Can only be called within go for loops or SPMD functions (functions with varying parameters).
// COMPILER BUILTIN: This function cannot be implemented in Go - it must be handled
// by the compiler as a builtin that generates lane index vectors like [0,1,2,3].
//
//go:noinline
func Index() Varying[int] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.Index is a compiler builtin and should be replaced during compilation")
}

// Count returns the number of SIMD lanes for the given varying type.
// This is determined at compile time based on the SIMD width and element type.
// COMPILER BUILTIN: Should be replaced with compile-time constant, but provides
// PoC implementation for Phase 1.8 testing until compiler handles it.
//
//go:noinline
func Count[T any](value Varying[T]) int {
	// Phase 1.8: Runtime type inspection for PoC - WASM SIMD128 calculation
	// Formula: 128 bits / (sizeof(T) * 8 bits) = lane count
	// TODO Phase 2: Compiler should replace with compile-time constant

	// Get the size of the underlying type T via runtime type inspection
	var zero T
	switch any(zero).(type) {
	case int8, uint8, bool:
		return 16 // 128/8 = 16 lanes
	case int16, uint16:
		return 8 // 128/16 = 8 lanes
	case int32, uint32, float32:
		return 4 // 128/32 = 4 lanes
	case int64, uint64, float64:
		return 2 // 128/64 = 2 lanes
	case int, uint, uintptr:
		// Platform dependent - assume 32-bit for WASM PoC
		return 4 // 128/32 = 4 lanes
	default:
		// For complex types, assume 32-bit size as reasonable default
		return 4 // Default fallback for PoC
	}
}

// Broadcast takes a value from the specified lane and broadcasts it to all lanes.
//
//go:noinline
func Broadcast[T any](value Varying[T], lane int) Varying[T] {
	return broadcastBuiltin(value, lane)
}

// broadcastBuiltin is the actual compiler builtin for Broadcast.
//
//go:noinline
func broadcastBuiltin[T any](value Varying[T], lane int) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.broadcastBuiltin is a compiler builtin and should be replaced during compilation")
}

// Rotate shifts values across lanes by the specified offset.
// Positive offset rotates right, negative rotates left.
//
//go:noinline
func Rotate[T any](value Varying[T], offset int) Varying[T] {
	return rotateBuiltin(value, offset)
}

// rotateBuiltin is the actual compiler builtin for Rotate.
//
//go:noinline
func rotateBuiltin[T any](value Varying[T], offset int) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.rotateBuiltin is a compiler builtin and should be replaced during compilation")
}

// From converts a uniform slice to varying values.
// Each lane gets the corresponding slice element.
// COMPILER BUILTIN: This function cannot be implemented in Go - it must be handled
// by the compiler as a builtin intrinsic that generates SIMD load instructions.
//
//go:noinline
func From[T any](data []T) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.From is a compiler builtin and should be replaced during compilation")
}

// Swizzle performs arbitrary permutation of lane values based on indices.
//
//go:noinline
func Swizzle[T any](value Varying[T], indices Varying[int]) Varying[T] {
	return swizzleBuiltin(value, indices)
}

// swizzleBuiltin is the actual compiler builtin for Swizzle.
//
//go:noinline
func swizzleBuiltin[T any](value Varying[T], indices Varying[int]) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.swizzleBuiltin is a compiler builtin and should be replaced during compilation")
}

// ShiftLeft performs a cross-lane left shift, filling vacated lanes with zero.
//
//go:noinline
func ShiftLeft[T integer](value Varying[T], shift Varying[T]) Varying[T] {
	return shiftLeftBuiltin(value, shift)
}

// shiftLeftBuiltin is the actual compiler builtin for ShiftLeft.
//
//go:noinline
func shiftLeftBuiltin[T integer](value Varying[T], shift Varying[T]) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.shiftLeftBuiltin is a compiler builtin and should be replaced during compilation")
}

// ShiftRight performs a cross-lane right shift, filling vacated lanes with zero.
//
//go:noinline
func ShiftRight[T integer](value Varying[T], shift Varying[T]) Varying[T] {
	return shiftRightBuiltin(value, shift)
}

// shiftRightBuiltin is the actual compiler builtin for ShiftRight.
//
//go:noinline
func shiftRightBuiltin[T integer](value Varying[T], shift Varying[T]) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.shiftRightBuiltin is a compiler builtin and should be replaced during compilation")
}

// RotateWithin rotates values within independent groups of groupSize lanes.
// Groups: lanes [0..groupSize-1], [groupSize..2*groupSize-1], etc.
// groupSize must be a compile-time constant that evenly divides the lane count.
//
//go:noinline
func RotateWithin[T any](v Varying[T], offset int, groupSize int) Varying[T] {
	panic("lanes.RotateWithin is a compiler builtin and should be replaced during compilation")
}

// ShiftLeftWithin shifts values left within independent groups, filling with zero.
// groupSize must be a compile-time constant that evenly divides the lane count.
//
//go:noinline
func ShiftLeftWithin[T any](v Varying[T], amount int, groupSize int) Varying[T] {
	panic("lanes.ShiftLeftWithin is a compiler builtin and should be replaced during compilation")
}

// ShiftRightWithin shifts values right within independent groups, filling with zero.
// groupSize must be a compile-time constant that evenly divides the lane count.
//
//go:noinline
func ShiftRightWithin[T any](v Varying[T], amount int, groupSize int) Varying[T] {
	panic("lanes.ShiftRightWithin is a compiler builtin and should be replaced during compilation")
}

// SwizzleWithin permutes values within independent groups using indices.
// groupSize must be a compile-time constant that evenly divides the lane count.
//
//go:noinline
func SwizzleWithin[T any](v Varying[T], indices Varying[int], groupSize int) Varying[T] {
	panic("lanes.SwizzleWithin is a compiler builtin and should be replaced during compilation")
}

// =====================================================================
// FP math primitives
// Each function is an element-wise floating-point operation lowered to
// the corresponding LLVM vector intrinsic (@llvm.<op>.vN{f32,f64}).
// =====================================================================

// Sqrt returns the per-lane square root of value.
// NaN for negative inputs per IEEE 754. Lowered to @llvm.sqrt.vN{f32,f64}.
//
//go:noinline
func Sqrt[T floatingPoint](value Varying[T]) Varying[T] {
	return sqrtBuiltin(value)
}

//go:noinline
func sqrtBuiltin[T floatingPoint](value Varying[T]) Varying[T] {
	panic("lanes.sqrtBuiltin is a compiler builtin and should be replaced during compilation")
}

// Abs returns the per-lane absolute value. Lowered to @llvm.fabs.vN{f32,f64}.
//
//go:noinline
func Abs[T floatingPoint](value Varying[T]) Varying[T] {
	return absBuiltin(value)
}

//go:noinline
func absBuiltin[T floatingPoint](value Varying[T]) Varying[T] {
	panic("lanes.absBuiltin is a compiler builtin and should be replaced during compilation")
}

// Floor returns the per-lane largest integer value <= value (round toward -inf).
// Lowered to @llvm.floor.vN{f32,f64}.
//
//go:noinline
func Floor[T floatingPoint](value Varying[T]) Varying[T] {
	return floorBuiltin(value)
}

//go:noinline
func floorBuiltin[T floatingPoint](value Varying[T]) Varying[T] {
	panic("lanes.floorBuiltin is a compiler builtin and should be replaced during compilation")
}

// Ceil returns the per-lane smallest integer value >= value (round toward +inf).
// Lowered to @llvm.ceil.vN{f32,f64}.
//
//go:noinline
func Ceil[T floatingPoint](value Varying[T]) Varying[T] {
	return ceilBuiltin(value)
}

//go:noinline
func ceilBuiltin[T floatingPoint](value Varying[T]) Varying[T] {
	panic("lanes.ceilBuiltin is a compiler builtin and should be replaced during compilation")
}

// Round returns the per-lane value rounded to the nearest integer, with ties
// rounded away from zero (matches Go's math.Round semantics).
// Lowered to @llvm.round.vN{f32,f64}.
//
//go:noinline
func Round[T floatingPoint](value Varying[T]) Varying[T] {
	return roundBuiltin(value)
}

//go:noinline
func roundBuiltin[T floatingPoint](value Varying[T]) Varying[T] {
	panic("lanes.roundBuiltin is a compiler builtin and should be replaced during compilation")
}

// Trunc returns the per-lane value with its fractional part removed
// (rounded toward zero). Lowered to @llvm.trunc.vN{f32,f64}.
//
//go:noinline
func Trunc[T floatingPoint](value Varying[T]) Varying[T] {
	return truncBuiltin(value)
}

//go:noinline
func truncBuiltin[T floatingPoint](value Varying[T]) Varying[T] {
	panic("lanes.truncBuiltin is a compiler builtin and should be replaced during compilation")
}

// Min returns the per-lane minimum of a and b. IEEE 754 minNum semantics:
// when one operand is NaN, returns the other; matches Go's built-in min().
// Note: this is the PAIRWISE per-lane min. To collapse a Varying into a scalar
// minimum, use reduce.Min. Lowered to @llvm.minnum.vN{f32,f64}.
//
//go:noinline
func Min[T floatingPoint](a, b Varying[T]) Varying[T] {
	return minBuiltin(a, b)
}

//go:noinline
func minBuiltin[T floatingPoint](a, b Varying[T]) Varying[T] {
	panic("lanes.minBuiltin is a compiler builtin and should be replaced during compilation")
}

// Max returns the per-lane maximum of a and b. IEEE 754 maxNum semantics:
// when one operand is NaN, returns the other; matches Go's built-in max().
// Note: this is the PAIRWISE per-lane max. To collapse a Varying into a scalar
// maximum, use reduce.Max. Lowered to @llvm.maxnum.vN{f32,f64}.
//
//go:noinline
func Max[T floatingPoint](a, b Varying[T]) Varying[T] {
	return maxBuiltin(a, b)
}

//go:noinline
func maxBuiltin[T floatingPoint](a, b Varying[T]) Varying[T] {
	panic("lanes.maxBuiltin is a compiler builtin and should be replaced during compilation")
}

// FMA returns the per-lane fused multiply-add: a*b + c.
//
// Rounding depends on the backend:
//   - x86 (+fma), and any target with hardware vector FMA: exact IEEE 754
//     single-rounding (no intermediate rounding of the product), matching
//     Go's math.FMA (added in Go 1.14). Lowered to @llvm.fma.vN{f32,f64}.
//   - WASM with +relaxed-simd: lowered to @llvm.wasm.relaxed.madd
//     (f32x4.relaxed_madd / f64x2.relaxed_madd), whose rounding is
//     implementation-defined — the runtime may use a single fused rounding
//     or two roundings (unfused mul+add). This avoids scalarizing to per-lane
//     fmaf libcalls, which base WASM SIMD128 would otherwise require since it
//     has no vector FMA instruction.
//
// Use FMA when you want the compiler to emit a single multiply-add
// instruction; rely on its exact single-rounding only on backends that
// guarantee it.
//
//go:noinline
func FMA[T floatingPoint](a, b, c Varying[T]) Varying[T] {
	return fmaBuiltin(a, b, c)
}

//go:noinline
func fmaBuiltin[T floatingPoint](a, b, c Varying[T]) Varying[T] {
	panic("lanes.fmaBuiltin is a compiler builtin and should be replaced during compilation")
}

// Type constraints for generic functions
type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// floatingPoint constrains lanes.* math builtins to IEEE 754 binary float types.
// Mirrors Go's stdlib math package, which handles both float32 and float64 internally.
type floatingPoint interface {
	~float32 | ~float64
}
