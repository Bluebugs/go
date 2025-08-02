//go:build goexperiment.spmd

// Test SSA generation for go for loops
package spmdtest

// Test basic go for loop generates correct SSA opcodes
func testBasicGoForSSA() {
	// EXPECT SSA: OpPhi (for loop counter)
	// EXPECT SSA: OpVectorLoad (for varying data loading)
	// EXPECT SSA: OpVectorAdd (for varying arithmetic)
	// EXPECT SSA: OpVectorStore (for varying data storing)
	go for i := range 16 {
		var data varying int32 = varying int32(i) * 2
		process(int(data))
	}
}

// Test go for loop with masking generates mask operations
func testGoForWithMaskingSSA() {
	// EXPECT SSA: OpPhi (for loop counter and mask)
	// EXPECT SSA: OpAnd (for mask intersection with condition)
	// EXPECT SSA: OpSelect (for conditional execution with mask)
	// EXPECT SSA: OpOr (for mask combination)
	go for i := range 16 {
		var condition varying bool = i > 8
		if condition {
			var result varying int32 = varying int32(i) * 3
			process(int(result))
		}
	}
}

// Test constrained go for loop generates unrolling or chunking
func testConstrainedGoForSSA() {
	// EXPECT SSA: OpPhi (for chunk iteration)
	// EXPECT SSA: OpVectorLoad (for constrained varying data)
	// EXPECT SSA: OpCall (to lanes.FromConstrained)
	go for i := range[4] 16 {
		var data varying[4] int32
		process(int(data[0]))
		_ = i
	}
}

// Test nested control flow in go for generates complex mask tracking
func testNestedControlFlowSSA() {
	// EXPECT SSA: OpPhi (for outer mask)
	// EXPECT SSA: OpAnd (for inner condition mask)
	// EXPECT SSA: OpSelect (for conditional varying operations)
	// EXPECT SSA: OpNot (for negating conditions)
	go for i := range 16 {
		var outer varying bool = i < 8
		if outer {
			var inner varying bool = i%2 == 0
			if inner {
				var result varying int32 = varying int32(i) + 10
				process(int(result))
			}
		}
	}
}

// Test loop with continue generates mask updates
func testGoForContinueSSA() {
	// EXPECT SSA: OpPhi (for continue mask tracking)
	// EXPECT SSA: OpOr (for accumulating continue conditions)
	// EXPECT SSA: OpAndNot (for excluding continued lanes)
	go for i := range 16 {
		if i%3 == 0 {
			continue
		}
		var data varying int32 = varying int32(i) * 4
		process(int(data))
	}
}

// Helper function
func process(x int) {
	_ = x
}