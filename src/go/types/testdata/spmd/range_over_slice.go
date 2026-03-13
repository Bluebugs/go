// -goexperiment spmd

// Test SPMD range-over-slice and range-over-array type inference.
// The value variable must receive Varying[elem], not Varying[container].
package spmdtest

import "lanes"

// Range over a slice: value variable gets Varying[int32], not Varying[[]int32]
func rangeOverSlice(data []int32) {
	go for i, v := range data {
		var _ lanes.Varying[int] = i    // key is Varying[int]
		var _ lanes.Varying[int32] = v  // value must be Varying[int32]
	}
}

// Range over an array: value variable gets Varying[float64]
func rangeOverArray(data [16]float64) {
	go for _, v := range data {
		var _ lanes.Varying[float64] = v
	}
}

// Range over a slice with only key variable
func rangeOverSliceKeyOnly(data []int) {
	go for i := range data {
		var _ lanes.Varying[int] = i
	}
}

// Range over an integer (existing behavior, should still work)
func rangeOverInt() {
	go for i := range 16 {
		_ = i
	}
}

// Range over a slice of pointers: pointer element type is allowed
func rangeOverPtrSlice(data []*int32) {
	go for _, v := range data {
		_ = v
	}
}

// Range over a slice of slices: aggregate element type lowers to [1 x sliceStruct]
// (serial execution, laneCount=1). No longer a type-checker error.
func rangeOverSliceOfSlices(data [][]int) {
	go for _, v := range data {
		_ = v
	}
}

// Range over a slice of structs: aggregate element type lowers to [1 x struct]
// (serial execution, laneCount=1). No longer a type-checker error.
type Point struct{ X, Y int32 }

func rangeOverStructSlice(data []Point) {
	go for _, v := range data {
		_ = v
	}
}
