//go:build goexperiment.spmd

// Package lanes provides cross-lane operations for SPMD programming.
// These functions enable data movement and communication between SIMD lanes.
//
// IMPORTANT: Most functions in this package are COMPILER BUILTINS that cannot be
// implemented in regular Go code. They must be handled specially by the compiler
// during compilation and replaced with appropriate SIMD instructions.
//
// ARCHITECTURE: Cross-lane operations (Broadcast, Rotate, Swizzle, ShiftLeft, ShiftRight)
// handle both regular Varying[T] and constrained Varying[T, N] types:
// - Regular Varying[T]: Direct builtin replacement by compiler
// - Constrained Varying[T, N]: Conversion using FromConstrained (returns array of Varying[T]),
//   then iteration over array applying builtin to each element, then reconstruction
//
// The Go source code here serves only for:
// 1. Type checking and validation during Phase 1
// 2. Documentation of the expected API
// 3. Constrained varying handling logic
// 4. Placeholder implementations that panic if somehow executed
package lanes

// Varying represents a value that differs across SIMD lanes (a vector value).
// The type checker special-cases this type:
// - Varying[T] is an unconstrained varying type
// - Varying[T, N] is a constrained varying type with N lanes (compiler magic, not standard generics)
type Varying[T any] struct{ _ [0]T }

// PHASE 1.8: Compiler builtin declarations for SPMD lane operations.
// These functions should never execute at runtime - they must be replaced by the compiler.

// Index returns the current lane index (0 to Count-1) in SPMD context.
// Can only be called within go for loops or SPMD functions (functions with varying parameters).
// COMPILER BUILTIN: This function cannot be implemented in Go - it must be handled
// by the compiler as a builtin that generates lane index vectors like [0,1,2,3].
func Index() Varying[int] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.Index is a compiler builtin and should be replaced during compilation")
}

// Count returns the number of SIMD lanes for the given varying type.
// This is determined at compile time based on the SIMD width and element type.
// COMPILER BUILTIN: Should be replaced with compile-time constant, but provides
// PoC implementation for Phase 1.8 testing until compiler handles it.
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
// COMPILER BUILTIN for regular varying types. Constrained varying handled in Phase 2.
func Broadcast[T any](value Varying[T], lane int) Varying[T] {
	// Direct builtin call - constrained Varying[T, N] support in Phase 2
	return broadcastBuiltin(value, lane)
}

// broadcastBuiltin is the actual compiler builtin for regular varying types only
func broadcastBuiltin[T any](value Varying[T], lane int) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.broadcastBuiltin is a compiler builtin and should be replaced during compilation")
}

// Rotate shifts values across lanes by the specified offset.
// Positive offset rotates right, negative rotates left.
// COMPILER BUILTIN for regular varying types. Constrained varying handled in Phase 2.
func Rotate[T any](value Varying[T], offset int) Varying[T] {
	// Direct builtin call - constrained Varying[T, N] support in Phase 2
	return rotateBuiltin(value, offset)
}

// rotateBuiltin is the actual compiler builtin for regular varying types only
func rotateBuiltin[T any](value Varying[T], offset int) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.rotateBuiltin is a compiler builtin and should be replaced during compilation")
}

// From converts a uniform slice to varying values.
// Each lane gets the corresponding slice element.
// COMPILER BUILTIN: This function cannot be implemented in Go - it must be handled
// by the compiler as a builtin intrinsic that generates SIMD load instructions.
func From[T any](data []T) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.From is a compiler builtin and should be replaced during compilation")
}

// FromConstrained converts constrained varying to unconstrained varying plus mask.
// Returns (values, mask) where mask indicates which lanes are active.
// COMPILER BUILTIN: This function converts Varying[T, 0] to array of Varying[T] values
func FromConstrained[T any](data Varying[T]) ([]Varying[T], []Varying[bool]) {
	// This is a compiler builtin - execution should never reach here
	// Phase 2: Compiler should replace with conversion that:
	// 1. Extracts the constraint size from Varying[T, N] type
	// 2. Creates array of Varying[T] with constraint elements
	// 3. Generates mask indicating which lanes are active
	panic("lanes.FromConstrained is a compiler builtin and should be replaced during compilation")
}

// ToConstrained converts unconstrained varying arrays back to constrained varying.
// COMPILER BUILTIN
func ToConstrained[T any](data []Varying[T], mask []Varying[bool], target Varying[T]) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	// Phase 2: Compiler should replace with conversion that:
	// 1. Takes array of Varying[T] and mask
	// 2. Constructs Varying[T, N] with appropriate constraint
	panic("lanes.ToConstrained is a compiler builtin and should be replaced during compilation")
}

// Swizzle performs arbitrary permutation of lane values based on indices.
// COMPILER BUILTIN for regular varying types. Constrained varying handled in Phase 2.
func Swizzle[T any](value Varying[T], indices Varying[int]) Varying[T] {
	// Direct builtin call - constrained Varying[T, N] support in Phase 2
	return swizzleBuiltin(value, indices)
}

// swizzleBuiltin is the actual compiler builtin for regular varying types only
func swizzleBuiltin[T any](value Varying[T], indices Varying[int]) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.swizzleBuiltin is a compiler builtin and should be replaced during compilation")
}

// ShiftLeft performs per-lane left shift operation.
// COMPILER BUILTIN for regular varying types. Constrained varying handled in Phase 2.
func ShiftLeft[T integer](value Varying[T], shift Varying[T]) Varying[T] {
	// Direct builtin call - constrained Varying[T, N] support in Phase 2
	return shiftLeftBuiltin(value, shift)
}

// shiftLeftBuiltin is the actual compiler builtin for regular varying types only
func shiftLeftBuiltin[T integer](value Varying[T], shift Varying[T]) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.shiftLeftBuiltin is a compiler builtin and should be replaced during compilation")
}

// ShiftRight performs per-lane right shift operation.
// COMPILER BUILTIN for regular varying types. Constrained varying handled in Phase 2.
func ShiftRight[T integer](value Varying[T], shift Varying[T]) Varying[T] {
	// Direct builtin call - constrained Varying[T, N] support in Phase 2
	return shiftRightBuiltin(value, shift)
}

// shiftRightBuiltin is the actual compiler builtin for regular varying types only
func shiftRightBuiltin[T integer](value Varying[T], shift Varying[T]) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	panic("lanes.shiftRightBuiltin is a compiler builtin and should be replaced during compilation")
}

// Helper functions for constrained varying conversion (compiler builtins)

// isConstrainedVarying checks if a varying value is actually a constrained varying (Varying[T, N])
// COMPILER BUILTIN: This should be replaced by the compiler with type inspection
func isConstrainedVarying[T any](value Varying[T]) bool {
	// This is a compiler builtin - execution should never reach here
	// For Phase 1.8 PoC, always return false (assume regular varying)
	// TODO Phase 2: Compiler should replace with actual type inspection
	return false
}

// convertConstrainedToVarying converts a constrained varying to regular varying
// PHASE 2 ARCHITECTURE: For constrained Varying[T, N] operations:
// 1. Use FromConstrained(value) to get array of Varying[T] elements
// 2. Iterate over the array applying builtin operation to each Varying[T]
// 3. Reconstruct result back into constrained Varying[T, N] format
func convertConstrainedToVarying[T any](value Varying[T]) Varying[T] {
	// This is a compiler builtin - execution should never reach here
	// Phase 1.8: Returns unchanged since we can't detect constrained varying yet
	// Phase 2: Compiler should replace with proper FromConstrained conversion
	return value
}

// Type constraints for generic functions
type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}
