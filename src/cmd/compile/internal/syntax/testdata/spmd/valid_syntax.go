//go:build goexperiment.spmd

// Valid SPMD syntax that should parse successfully
// when GOEXPERIMENT=spmd is enabled
package spmdtest

import (
	"lanes"
	"reduce"
)

// Valid variable declarations using lanes.Varying[T]
func validDeclarations() {
	var a int = 42
	var b lanes.Varying[float32]
	var e []lanes.Varying[int]

	_ = a
	_, _, _ = a, b, e
}

// Valid go for loop syntax
func validGoFor() {
	// Basic go for with range
	go for i := range 10 {
		process(i)
	}

	// go for with slice
	data := []int{1, 2, 3, 4}
	go for _, value := range data {
		process(value)
	}

	// Infinite go for loop
	go for {
		process(0) // infinite loop
	}
}

// Valid built-in function usage
func validBuiltins() {
	go for i := range 8 {
		var data lanes.Varying[int] = i

		// lanes functions
		count := lanes.Count(data)
		index := lanes.Index()
		broadcast := lanes.Broadcast(data, 0)
		rotated := lanes.Rotate(data, 1)

		// reduce functions
		sum := reduce.Add(data)
		all := reduce.All(data > 5)
		any := reduce.Any(data > 5)

		_, _, _, _, _, _, _ = count, index, broadcast, rotated, sum, all, any
	}
}

// Valid control flow (ISPC-based rules)
func validControlFlow() {
	threshold := 7

	go for i := range 10 {
		// VALID: return/break under uniform conditions
		if threshold < 0 {
			return // valid under uniform condition
		}

		if threshold > 100 {
			break // valid under uniform condition
		}

		// VALID: continue always allowed
		if i > 5 { // varying condition
			continue // always valid in go for
		}

		process(i)
	}
}

func process(x lanes.Varying[int]) {
	// Valid SPMD function with varying parameter
	_ = x
}
