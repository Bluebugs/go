//go:build goexperiment.spmd

// Test SSA generation for lanes package builtin interception
package spmdtest

import "lanes"

// Test lanes.Broadcast generates OpSPMDBroadcastLane
func testBroadcastSSA() {
	// EXPECT SSA: OpSPMDBroadcastLane (lanes.Broadcast interception)
	go for i := range 16 {
		b := lanes.Broadcast(i, 0)
		processVarying(b)
	}
}

// Test lanes.Broadcast from different lanes
func testBroadcastLanesSSA() {
	// EXPECT SSA: OpSPMDBroadcastLane (lanes.Broadcast with lane 1)
	// EXPECT SSA: OpSPMDBroadcastLane (lanes.Broadcast with lane 2)
	go for i := range 16 {
		b1 := lanes.Broadcast(i, 1)
		b2 := lanes.Broadcast(i, 2)
		processVarying(b1 + b2)
	}
}

// Test lanes.Rotate generates OpSPMDRotate
func testRotateSSA() {
	// EXPECT SSA: OpSPMDRotate (lanes.Rotate interception)
	go for i := range 16 {
		r := lanes.Rotate(i, 1)
		processVarying(r)
	}
}

// Test lanes.Swizzle generates OpSPMDSwizzle
func testSwizzleSSA() {
	// EXPECT SSA: OpSPMDSwizzle (lanes.Swizzle interception)
	go for i := range 16 {
		idx := lanes.Index()
		s := lanes.Swizzle(i, idx)
		processVarying(s)
	}
}

// Test lanes.ShiftLeft generates OpSPMDShiftLeft
func testShiftLeftSSA() {
	// EXPECT SSA: OpSPMDShiftLeft (lanes.ShiftLeft interception)
	go for i := range 16 {
		var fill lanes.Varying[int] = 0
		l := lanes.ShiftLeft(i, fill)
		processVarying(l)
	}
}

// Test lanes.ShiftRight generates OpSPMDShiftRight
func testShiftRightSSA() {
	// EXPECT SSA: OpSPMDShiftRight (lanes.ShiftRight interception)
	go for i := range 16 {
		var fill lanes.Varying[int] = 0
		r := lanes.ShiftRight(i, fill)
		processVarying(r)
	}
}

// Test combined cross-lane operations
func testCombinedCrossLaneSSA() {
	// EXPECT SSA: OpSPMDBroadcastLane (lanes.Broadcast)
	// EXPECT SSA: OpSPMDRotate (lanes.Rotate)
	// EXPECT SSA: OpSPMDSwizzle (lanes.Swizzle)
	go for i := range 16 {
		b := lanes.Broadcast(i, 0)
		r := lanes.Rotate(i, 1)
		idx := lanes.Index()
		s := lanes.Swizzle(i, idx)
		processVarying(b + r + s)
	}
}

// Test cross-lane operations with arithmetic
func testCrossLaneWithArithmeticSSA() {
	// EXPECT SSA: OpSPMDBroadcastLane (lanes.Broadcast in expression)
	// EXPECT SSA: OpSPMDAdd (varying arithmetic)
	go for i := range 16 {
		b := lanes.Broadcast(i, 0)
		result := b + i*2
		processVarying(result)
	}
}

// Helper function to consume varying values
func processVarying(x lanes.Varying[int]) {
	_ = x
}
