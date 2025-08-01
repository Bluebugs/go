//go:build goexperiment.spmd

// Test pointer operation validations with varying types
package spmdtest

import (
	"reduce"
	"unsafe"
)

// Test basic pointer operations with varying
func testBasicPointerOperations() {
	var data [16]int
	var vPtr varying *int
	
	// Valid: varying pointer to uniform data
	go for i := range 4 {
		vPtr = &data[i]  // Each lane points to different element
		value := *vPtr   // Dereference varying pointer
		process(value)
	}
	
	// Valid: uniform pointer in SPMD context
	uPtr := &data[0]
	go for i := range 4 {
		value := *uPtr  // Same pointer accessed by all lanes
		process(value + int(i))
	}
}

// Test pointer arithmetic with varying
func testVaryingPointerArithmetic() {
	var data [16]int
	basePtr := &data[0]
	
	go for i := range 8 {
		// Valid: pointer arithmetic with varying offset
		vOffset := varying uintptr(i)
		vPtr := (*int)(unsafe.Add(unsafe.Pointer(basePtr), vOffset*unsafe.Sizeof(int(0))))
		value := *vPtr
		process(value)
	}
}

// Test invalid pointer operations
func testInvalidPointerOperations() {
	var data varying int
	
	// ERROR "cannot take address of varying variable"
	ptr := &data
	
	// ERROR "varying pointer arithmetic not supported in this context"
	var vPtr varying *int
	vPtr++  // Increment varying pointer (not supported)
	
	_ = ptr
}

// Test pointer assignment rules
func testPointerAssignmentRules() {
	var data [16]int
	var uPtr uniform *int = &data[0]
	var vPtr varying *int
	
	// Valid: uniform to varying pointer assignment
	vPtr = uPtr  // Broadcast uniform pointer to all lanes
	
	// ERROR "cannot assign varying pointer to uniform variable"
	uPtr = vPtr
	
	_ = vPtr
}

// Test pointer function parameters
func testPointerFunctionParameters() {
	var data [16]int
	
	go for i := range 4 {
		vPtr := &data[i]
		
		// Valid: pass varying pointer to SPMD function
		result := processPtrSPMD(vPtr)
		process(result)
		
		// Valid: pass varying pointer to uniform function
		processPtrUniform(vPtr)  // Should work with first lane
	}
}

// SPMD function taking varying pointer
func processPtrSPMD(ptr varying *int) varying int {
	return *ptr * 2
}

// Uniform function taking uniform pointer
func processPtrUniform(ptr uniform *int) {
	value := *ptr
	process(value)
}

// Test pointer to varying types
func testPointerToVaryingTypes() {
	// ERROR "pointer to varying type not supported"
	var invalidPtr *varying int
	
	// ERROR "array of pointers to varying not supported"
	var invalidArray [4]*varying int
	
	_, _ = invalidPtr, invalidArray
}

// Test slice operations with varying pointers
func testSliceOperationsWithVarying() {
	var data []int = make([]int, 16)
	
	go for i := range 4 {
		// Valid: varying slice indexing
		vIndex := varying int(i * 2)
		value := data[vIndex]  // Each lane accesses different index
		process(value)
		
		// Valid: slice pointer operations
		vPtr := &data[vIndex]
		result := *vPtr
		process(result)
	}
}

// Test nil pointer handling
func testNilPointerHandling() {
	var vPtr varying *int
	
	go for i := range 4 {
		// Valid: nil checking with varying pointer
		if vPtr != nil {
			value := *vPtr
			process(value)
		}
		
		// Valid: comparison with nil
		isNil := vPtr == nil
		if reduce.Any(isNil) {
			// Handle nil lanes
			continue
		}
	}
}

// Test interface{} with varying pointers
func testInterfaceWithVaryingPointers() {
	var data [16]int
	
	go for i := range 4 {
		vPtr := &data[i]
		
		// Valid: varying pointer as interface{}
		var iface interface{} = vPtr
		
		// Type switch with varying pointer
		switch v := iface.(type) {
		case varying *int:
			value := *v
			process(value)
		default:
			// ERROR "varying types in type switch must be handled explicitly"
		}
	}
}

// Test memory safety with varying pointers
func testMemorySafetyVaryingPointers() {
	var data [16]int
	
	go for i := range 4 {
		// Valid: bounds checking
		if i < len(data) {
			vPtr := &data[i]
			value := *vPtr
			process(value)
		}
		
		// ERROR "potential out-of-bounds access with varying pointer"
		vPtr := &data[i*4]  // Could exceed bounds depending on lane
		value := *vPtr
		process(value)
	}
}

// Helper function
func process(x int) {
	_ = x
}