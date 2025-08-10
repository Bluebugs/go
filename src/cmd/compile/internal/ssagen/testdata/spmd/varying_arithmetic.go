//go:build goexperiment.spmd

// Test SSA generation for varying arithmetic operations
package spmdtest

// Test basic varying arithmetic generates vector operations
func testBasicVaryingArithmeticSSA() {
	// EXPECT SSA: OpVectorAdd (for varying addition)
	// EXPECT SSA: OpVectorMul (for varying multiplication)
	// EXPECT SSA: OpVectorSub (for varying subtraction)
	// EXPECT SSA: OpVectorDiv (for varying division)
	go for i := range 42 {
		var a varying int32 = 10 + i
		var b varying int32 = 20
	
		var sum varying int32 = a + b
		var product varying int32 = a * b
		var difference varying int32 = b - a
		var quotient varying int32 = b / a
	
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
		var a varying float32 = 3.14 + i
		var b varying float32 = 2.71
	
		var sum varying float32 = a + b
		var product varying float32 = a * b
		var difference varying float32 = a - b
		var quotient varying float32 = a / b
	
		process(sum + product + difference + quotient)
	}
}

// Test mixed type arithmetic with automatic promotion
func testMixedTypeArithmeticSSA() {
	// EXPECT SSA: OpVectorCvt (for type conversions)
	// EXPECT SSA: OpVectorAdd (for converted arithmetic)
	go for intVal := range 50 {
		var floatVal varying float32 = 3.14
	
		// Type conversion should generate vector conversion operations
		var mixed varying float32 = varying float32(intVal) + floatVal
		var converted varying int32 = varying int32(floatVal) + intVal
	
		process(varying int32(mixed) + converted)
	}
}

// Test uniform-to-varying broadcasts in arithmetic
func testUniformBroadcastArithmeticSSA() {
	// EXPECT SSA: OpCall (to lanes.Broadcast for uniform-to-varying)
	// EXPECT SSA: OpVectorAdd (for arithmetic with broadcasted values)
	uniformVal := 100

	go for varyingVal := range 100 {
		// Uniform should be automatically broadcasted
		var result varying int32 = uniformVal + varyingVal
		var result2 varying int32 = varyingVal * uniformVal
	
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
		var b varying int32 = 0x00FF
		var shift varying int32 = 4
	
		var andResult varying int32 = a & b
		var orResult varying int32 = a | b
		var xorResult varying int32 = a ^ b
		var leftShift varying int32 = a << shift
		var rightShift varying int32 = a >> shift
	
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
		var b varying int32 = 20
	
		var eq varying bool = a == b
		var lt varying bool = a < b
		var gt varying bool = a > b
		var leq varying bool = a <= b
		var geq varying bool = a >= b
	
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
func process(x varying int) {
	_ = x
}