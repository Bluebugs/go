// -goexperiment spmd

// Test type switch on universal constrained varying types
package spmdtest

import "lanes"

// VALID: type switch on Varying[T, 0] with matching elem type cases
func typeSwitchValid(data lanes.Varying[int32, 0]) {
	switch v := data.(type) {
	case lanes.Varying[int32, 4]:
		_ = v * 2
	case lanes.Varying[int32, 8]:
		_ = v + 1
	default:
	}
}

// INVALID: wrong elem type in case
func typeSwitchWrongElem(data lanes.Varying[int32, 0]) {
	switch data.(type) {
	case lanes /* ERROR "element type mismatch" */ .Varying[float32, 4]:
	}
}

// INVALID: type switch on non-universal constrained
func typeSwitchNonUniversal(data lanes.Varying[int32, 4]) {
	switch data /* ERROR "is not an interface" */ .(type) {
	case lanes.Varying[int32, 4]:
	}
}

// INVALID: unconstrained varying case (constraint == -1)
func typeSwitchUnconstrainedCase(data lanes.Varying[int32, 0]) {
	switch data.(type) {
	case lanes /* ERROR "type switch on" */ .Varying[int32]:
	}
}
