// -goexperiment spmd

// Test type-checking for lanes.CompactStore.
package spmdtest

import "lanes"

// Valid: CompactStore in SPMD loop with correct types
func testCompactStoreValid(dst []byte, src []byte) {
	go for i, ch := range src {
		_ = i
		mask := ch > 0
		n := lanes.CompactStore(dst, ch, mask)
		_ = n
	}
}

// Valid: CompactStore in SPMD function body
func testCompactStoreInFunc(dst []int32, v lanes.Varying[int32], mask lanes.Varying[bool]) {
	n := lanes.CompactStore(dst, v, mask)
	_ = n
}

// ERROR: CompactStore with wrong slice element type
func testCompactStoreTypeMismatch(dst []int32, src []byte) {
	go for _, ch := range src {
		mask := ch > 0
		_ = lanes.CompactStore(dst, ch /* ERROR "does not match inferred type" */, mask)
	}
}
