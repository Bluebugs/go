//go:build goexperiment.spmd

// Package reduce provides reduction operations for SPMD programming.
// These functions convert varying values to uniform values through various reduction operations.
package reduce

// FIXME: This is a stub implementation for Phase 1.4 type system validation.
// All functions will panic at runtime until Phase 2+ implementation.

// Add reduces varying values to their uniform sum.
func Add[T numeric](data varying[] T) uniform T {
	// FIXME: Implement in Phase 2 - SIMD reduction sum
	panic("reduce.Add() not implemented - stub for Phase 1.4 type validation")
}

// All returns true if all lane values are true.
func All(data varying[] bool) uniform bool {
	// FIXME: Implement in Phase 2 - SIMD reduction AND
	panic("reduce.All() not implemented - stub for Phase 1.4 type validation")
}

// Any returns true if any lane value is true.
func Any(data varying[] bool) uniform bool {
	// FIXME: Implement in Phase 2 - SIMD reduction OR
	panic("reduce.Any() not implemented - stub for Phase 1.4 type validation")
}

// Max reduces varying values to their uniform maximum.
func Max[T ordered](data varying[] T) uniform T {
	// FIXME: Implement in Phase 2 - SIMD reduction max
	panic("reduce.Max() not implemented - stub for Phase 1.4 type validation")
}

// Min reduces varying values to their uniform minimum.
func Min[T ordered](data varying[] T) uniform T {
	// FIXME: Implement in Phase 2 - SIMD reduction min
	panic("reduce.Min() not implemented - stub for Phase 1.4 type validation")
}

// Or performs bitwise OR reduction across lanes.
func Or[T integer](data varying[] T) uniform T {
	// FIXME: Implement in Phase 2 - SIMD bitwise OR reduction
	panic("reduce.Or() not implemented - stub for Phase 1.4 type validation")
}

// And performs bitwise AND reduction across lanes.
func And[T integer](data varying[] T) uniform T {
	// FIXME: Implement in Phase 2 - SIMD bitwise AND reduction
	panic("reduce.And() not implemented - stub for Phase 1.4 type validation")
}

// Xor performs bitwise XOR reduction across lanes.
func Xor[T integer](data varying[] T) uniform T {
	// FIXME: Implement in Phase 2 - SIMD bitwise XOR reduction
	panic("reduce.Xor() not implemented - stub for Phase 1.4 type validation")
}

// From converts varying values to a uniform array.
// Each lane value becomes an array element.
func From[T any](data varying[] T) []T {
	// FIXME: Implement in Phase 2 - varying to array conversion
	panic("reduce.From() not implemented - stub for Phase 1.4 type validation")
}

// Count returns the number of true values across lanes.
func Count(data varying[] bool) uniform int {
	// FIXME: Implement in Phase 2 - SIMD population count
	panic("reduce.Count() not implemented - stub for Phase 1.4 type validation")
}

// FindFirstSet returns the index of the first true value.
func FindFirstSet(data varying[] bool) uniform int {
	// FIXME: Implement in Phase 2 - SIMD find first set
	panic("reduce.FindFirstSet() not implemented - stub for Phase 1.4 type validation")
}

// Sum is an alias for Add for compatibility.
func Sum[T numeric](data varying[] T) uniform T {
	return Add(data)
}

// Mask creates a mask from boolean values.
func Mask(data varying[] bool) uniform int {
	// FIXME: Implement in Phase 2 - boolean mask creation
	panic("reduce.Mask() not implemented - stub for Phase 1.4 type validation")
}

// Mul reduces varying values to their uniform product.
func Mul[T numeric](data varying[] T) uniform T {
	// FIXME: Implement in Phase 2 - SIMD reduction multiply
	panic("reduce.Mul() not implemented - stub for Phase 1.4 type validation")
}

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