// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends walk.go to support SPMD node types.
// With package-based types (lanes.Varying[T]), SPMDType AST node no longer exists.
// This file is kept as a placeholder for future SPMD-specific AST walking needs.

package syntax

// handleSPMDNode handles SPMD-specific AST nodes during tree walking.
// Currently a no-op since SPMDType AST node was removed in the migration
// to package-based types.
func (w walker) handleSPMDNode(n Node) bool {
	return false
}
