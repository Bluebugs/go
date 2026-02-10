//go:build goexperiment.spmd

// Test SSA generation for varying arithmetic operations
package spmdtest

import "lanes"

// Test basic varying arithmetic generates vector operations
func testBasicVaryingArithmeticSSA() {
	// EXPECT SSA: OpVectorAdd (for varying addition)
	// EXPECT SSA: OpVectorMul (for varying multiplication)
	// EXPECT SSA: OpVectorSub (for varying subtraction)
	// EXPECT SSA: OpVectorDiv (for varying division)
	go for i := range 42 {
		var a lanes.Varying[int32] = 10 + i
		var b lanes.Varying[int32] = 20

		var sum lanes.Varying[int32] = a + b
		var product lanes.Varying[int32] = a * b
		var difference lanes.Varying[int32] = b - a
		var quotient lanes.Varying[int32] = b / a

		process(sum + product + difference + quotient)
	}
}

// Test varying floating-point arithmetic
func testVaryingFloatArithmeticSSA() {
	// EXPECT SSA: OpVectorAdd (for f32x4 addition)
	// EXPECT SSA: OpVectorMul (for f32x4 multiplication)
	// EXPECT SSA: OpVectorSub (for f32x4 subtraction)
	// EXPECT SSA: OpVectorDiv (for f32x4 division)
	go for i := range 10 {
		var a lanes.Varying[float32] = 3.14 + i
		var b lanes.Varying[float32] = 2.71

		var sum lanes.Varying[float32] = a + b
		var product lanes.Varying[float32] = a * b
		var difference lanes.Varying[float32] = a - b
		var quotient lanes.Varying[float32] = a / b

		process(sum + product + difference + quotient)
	}
}

// Test mixed type arithmetic with automatic promotion
func testMixedTypeArithmeticSSA() {
	// EXPECT SSA: OpVectorCvt (for type conversions)
	// EXPECT SSA: OpVectorAdd (for converted arithmetic)
	go for intVal := range 50 {
		var floatVal lanes.Varying[float32] = 3.14

		// Type conversion should generate vector conversion operations
		var mixed lanes.Varying[float32] = lanes.Varying[float32](intVal) + floatVal
		var converted lanes.Varying[int32] = lanes.Varying[int32](floatVal) + intVal

		process(lanes.Varying[int32](mixed) + converted)
	}
}

// Test uniform-to-varying broadcasts in arithmetic
func testUniformBroadcastArithmeticSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast for uniform-to-varying)
	// EXPECT SSA: OpVectorAdd (for arithmetic with broadcasted values)
	uniformVal := 100

	go for varyingVal := range 100 {
		// Uniform should be automatically broadcasted
		var result lanes.Varying[int32] = uniformVal + varyingVal
		var result2 lanes.Varying[int32] = varyingVal * uniformVal

		process(result + result2)
	}
}

// Test bitwise operations on varying values
func testVaryingBitwiseSSA() {
	// EXPECT SSA: OpVectorAnd (for varying bitwise AND)
	// EXPECT SSA: OpVectorOr (for varying bitwise OR)
	// EXPECT SSA: OpVectorXor (for varying bitwise XOR)
	// EXPECT SSA: OpVectorShl (for varying left shift)
	// EXPECT SSA: OpVectorShr (for varying right shift)
	go for a := range 32 {
		var b lanes.Varying[int32] = 0x00FF
		var shift lanes.Varying[int32] = 4

		var andResult lanes.Varying[int32] = a & b
		var orResult lanes.Varying[int32] = a | b
		var xorResult lanes.Varying[int32] = a ^ b
		var leftShift lanes.Varying[int32] = a << shift
		var rightShift lanes.Varying[int32] = a >> shift

		process(andResult + orResult + xorResult + leftShift + rightShift)
	}
}

// Test comparison operations generating boolean vectors
func testVaryingComparisonsSSA() {
	// EXPECT SSA: OpVectorEq (for varying equality)
	// EXPECT SSA: OpVectorLt (for varying less than)
	// EXPECT SSA: OpVectorGt (for varying greater than)
	// EXPECT SSA: OpVectorLeq (for varying less than or equal)
	// EXPECT SSA: OpVectorGeq (for varying greater than or equal)
	go for a := range 42 {
		var b lanes.Varying[int32] = 20

		var eq lanes.Varying[bool] = a == b
		var lt lanes.Varying[bool] = a < b
		var gt lanes.Varying[bool] = a > b
		var leq lanes.Varying[bool] = a <= b
		var geq lanes.Varying[bool] = a >= b

		// Use comparisons in conditional to test mask generation
		if eq || lt || gt || leq || geq {
			process(a)
		}
	}
}

// Test vector loads and stores
func testVaryingMemoryOpsSSA() {
	// EXPECT SSA: OpVectorLoad (for varying array access)
	// EXPECT SSA: OpVectorStore (for varying array assignment)
	var data [16]int32

	go for i := range data {
		// Load should generate vector load with varying indices
		value := data[i]

		// Store should generate vector store with varying indices
		data[i] = value * 2
	}

	go for _, d := range data {
		process(d)
	}
}

// Helper function
func process(x lanes.Varying[int]) {
	_ = x
}
