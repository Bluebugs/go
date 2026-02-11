// Test varying switch statement masking in SPMD loops

package spmd

import "lanes"

func testVaryingSwitch() {
	var data [16]int
	var result [16]int

	go for i := range 16 {
		var val lanes.Varying[int] = data[i]

		// Varying switch - all cases execute with per-lane masking
		switch val {
		case 0:
			result[i] = 100
		case 1:
			result[i] = 200
		case 2, 3:
			result[i] = 300
		default:
			result[i] = 400
		}
	}
}

func testVaryingSwitchWithModification() {
	var data [16]int
	var result [16]int

	go for i := range 16 {
		var val lanes.Varying[int] = data[i]
		var x lanes.Varying[int] = 0

		// Varying switch modifying local variable
		switch val {
		case 0:
			x = 10
		case 1:
			x = 20
		default:
			x = 30
		}

		result[i] = x
	}
}

func testVaryingSwitchNoDefault() {
	var data [16]int
	var result [16]int

	go for i := range 16 {
		var val lanes.Varying[int] = data[i]

		// Varying switch without default
		switch val {
		case 0:
			result[i] = 100
		case 1:
			result[i] = 200
		case 2:
			result[i] = 300
		}
	}
}

func testVaryingSwitchWithVaryingCases() {
	var data [16]int
	var result [16]int

	go for i := range 16 {
		var selector lanes.Varying[int] = data[i]
		var threshold lanes.Varying[int]
		// Create a varying threshold value
		if selector > 10 {
			threshold = 5
		} else {
			threshold = 15
		}

		// Mix of varying and scalar case values
		switch selector {
		case threshold: // varying case value - per-lane comparison
			result[i] = 100
		case 0: // scalar case value - auto-splatted
			result[i] = 200
		default:
			result[i] = 300
		}
	}
}
