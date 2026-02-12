//go:build goexperiment.spmd

// Test SSA generation for reduce package builtin interception
package spmdtest

import (
	"lanes"
	"reduce"
)

// Test reduce.Add generates OpSPMDReduceAdd
func testReduceAddSSA() {
	// EXPECT SSA: OpSPMDReduceAdd (reduce.Add interception)
	var total int
	go for i := range 16 {
		total = reduce.Add(i)
	}
	_ = total
}

// Test reduce.All generates OpSPMDMaskAllTrue
func testReduceAllSSA() {
	// EXPECT SSA: OpSPMDMaskAllTrue (reduce.All interception)
	go for i := range 16 {
		cond := i > 5
		if reduce.All(cond) {
			continue
		}
	}
}

// Test reduce.Any generates OpSPMDMaskAnyTrue
func testReduceAnySSA() {
	// EXPECT SSA: OpSPMDMaskAnyTrue (reduce.Any interception)
	go for i := range 16 {
		cond := i > 10
		if reduce.Any(cond) {
			continue
		}
	}
}

// Test reduce.Min and reduce.Max generate OpSPMDReduceMin/Max
func testReduceMinMaxSSA() {
	// EXPECT SSA: OpSPMDReduceMin (reduce.Min interception)
	// EXPECT SSA: OpSPMDReduceMax (reduce.Max interception)
	var minVal, maxVal int
	go for i := range 16 {
		minVal = reduce.Min(i)
		maxVal = reduce.Max(i)
	}
	_, _ = minVal, maxVal
}

// Test reduce.Mul generates OpSPMDReduceMul
func testReduceMulSSA() {
	// EXPECT SSA: OpSPMDReduceMul (reduce.Mul interception)
	var product int
	go for i := range 16 {
		product = reduce.Mul(i + 1)
	}
	_ = product
}

// Test bitwise reductions generate correct opcodes
func testReduceBitwiseSSA() {
	// EXPECT SSA: OpSPMDReduceOr (reduce.Or interception)
	// EXPECT SSA: OpSPMDReduceAnd (reduce.And interception)
	// EXPECT SSA: OpSPMDReduceXor (reduce.Xor interception)
	go for i := range 16 {
		_ = reduce.Or(i)
		_ = reduce.And(i)
		_ = reduce.Xor(i)
	}
}

// Test reduce.All and reduce.Any in conditional context
func testReduceConditionalSSA() {
	// EXPECT SSA: OpSPMDMaskAllTrue (reduce.All in condition)
	// EXPECT SSA: OpSPMDMaskAnyTrue (reduce.Any in condition)
	go for i := range 16 {
		cond := i > 8
		if reduce.All(cond) {
			continue
		}
		if reduce.Any(cond) {
			continue
		}
	}
}

// Test multiple reduce operations in same loop
func testMultipleReduceSSA() {
	// EXPECT SSA: OpSPMDReduceAdd (reduce.Add)
	// EXPECT SSA: OpSPMDReduceMax (reduce.Max)
	// EXPECT SSA: OpSPMDMaskAnyTrue (reduce.Any)
	go for i := range 32 {
		sum := reduce.Add(i)
		max := reduce.Max(i)
		_ = sum
		_ = max
		if reduce.Any(i > 0) {
			continue
		}
	}
}

// Test reduce combined with lanes operations
func testReduceWithLanesSSA() {
	// EXPECT SSA: OpSPMDBroadcastLane (lanes.Broadcast)
	// EXPECT SSA: OpSPMDReduceAdd (reduce.Add)
	go for i := range 16 {
		b := lanes.Broadcast(i, 0)
		result := b + i
		_ = reduce.Add(result)
	}
}
