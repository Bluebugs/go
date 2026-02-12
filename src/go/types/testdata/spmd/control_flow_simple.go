// -goexperiment spmd

// Test basic control flow restrictions in SPMD contexts
package spmdtest

import "lanes"

// Test go for loop basic restrictions
func testGoForBasic() {
	// Valid go for loops - continue always allowed
	go for i := range 10 {
		if i > 5 {
			continue // OK: continue always allowed
		}
		processC(i)
	}

	// FORBIDDEN: break under varying condition
	go for i := range 10 {
		if i > 5 {
			break /* ERROR "break statement not allowed under varying conditions in SPMD for loop" */
		}
		processC(i)
	}

	// FORBIDDEN: return under varying condition
	go for i := range 10 {
		if i < 3 {
			return /* ERROR "return statement not allowed under varying conditions in SPMD for loop" */
		}
		processC(i)
	}

	// FORBIDDEN: nested go for loops
	go for i := range 10 {
		go /* ERROR "nested `go for` loop (prohibited for now)" */ for j := range 5 {
			processC(i + j)
		}
		_ = i // use i
	}
}

// Helper function
func processC(x lanes.Varying[int]) {
	_ = x
}
