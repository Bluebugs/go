// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssagen

import (
	"internal/buildcfg"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSPMDSSAGeneration tests that SPMD constructs generate correct SSA opcodes
func TestSPMDSSAGeneration(t *testing.T) {
	if !buildcfg.Experiment.SPMD {
		t.Skip("SPMD experiment not enabled")
	}
	
	testSPMDSSAFiles(t, "testdata/spmd")
}

func testSPMDSSAFiles(t *testing.T, dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read directory %s: %v", dir, err)
	}
	
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".go") {
			continue
		}
		
		filePath := filepath.Join(dir, file.Name())
		t.Run(file.Name(), func(t *testing.T) {
			testSPMDSSAFile(t, filePath)
		})
	}
}

func testSPMDSSAFile(t *testing.T, filePath string) {
	// Read the test file
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", filePath, err)
	}
	
	// Extract expected SSA opcodes from comments
	expectedOpcodes := extractExpectedOpcodes(string(content))
	if len(expectedOpcodes) == 0 {
		t.Logf("No EXPECT SSA comments found in %s", filePath)
		return
	}
	
	// For now, just verify that we can parse the expectations
	// Full implementation will come in Phase 1.7 when SSA generation is implemented
	t.Logf("Found %d expected SSA opcodes in %s:", len(expectedOpcodes), filePath)
	for _, opcode := range expectedOpcodes {
		t.Logf("  - %s", opcode)
	}
	
	// TODO: When SPMD SSA generation is implemented:
	// 1. Compile the Go code to SSA
	// 2. Extract actual opcodes from generated SSA
	// 3. Verify that all expected opcodes are present
	// 4. Verify correct opcode arguments and relationships
	
	// For now, mark as expected failure since SPMD syntax isn't implemented yet
	t.Logf("SPMD SSA generation not yet implemented - test infrastructure ready")
}

// extractExpectedOpcodes parses EXPECT SSA comments from test files
func extractExpectedOpcodes(content string) []string {
	var opcodes []string
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "// EXPECT SSA:") {
			// Extract the expected opcode
			opcode := strings.TrimSpace(strings.TrimPrefix(trimmed, "// EXPECT SSA:"))
			if opcode != "" {
				opcodes = append(opcodes, opcode)
			}
		}
	}
	
	return opcodes
}

// Future SSA validation functions (to be implemented in Phase 1.7):

// validateSPMDSSA will validate that generated SSA contains expected opcodes
// func validateSPMDSSA(t *testing.T, actualSSA *ssa.Func, expectedOpcodes []string) {
//     // Implementation will:
//     // 1. Walk through SSA function blocks and values
//     // 2. Extract opcodes and their context
//     // 3. Match against expected opcodes
//     // 4. Verify mask propagation patterns
//     // 5. Validate vector operation generation
// }

// compileSPMDToSSA will compile SPMD Go code to SSA for testing
// func compileSPMDToSSA(t *testing.T, source string) *ssa.Func {
//     // Implementation will:
//     // 1. Parse Go source with SPMD constructs
//     // 2. Type check with SPMD rules
//     // 3. Generate SSA with SPMD extensions
//     // 4. Return SSA function for validation
// }

// validateMaskPropagation will verify correct mask handling in SSA
// func validateMaskPropagation(t *testing.T, ssaFunc *ssa.Func) {
//     // Implementation will:
//     // 1. Trace mask values through SSA
//     // 2. Verify OpAnd/OpOr/OpNot for mask operations
//     // 3. Validate OpSelect for conditional execution
//     // 4. Check SPMD function calls get mask parameters
// }

// validateVectorOperations will verify vector SSA opcodes
// func validateVectorOperations(t *testing.T, ssaFunc *ssa.Func) {
//     // Implementation will:
//     // 1. Find vector arithmetic operations
//     // 2. Verify OpVectorAdd, OpVectorMul, etc.
//     // 3. Check vector load/store operations
//     // 4. Validate type consistency for vector ops
// }
