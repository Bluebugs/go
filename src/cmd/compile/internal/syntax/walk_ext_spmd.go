// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends walk.go to support SPMD node types.

package syntax

import "internal/buildcfg"

// spmdWalker extends the walker to handle SPMD node types.
// This function is called from the default case in walker.node()
// and will only work when SPMD experiment is enabled.
func (w walker) handleSPMDNode(n Node) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}

	switch n := n.(type) {
	case *SPMDType:
		// Walk the element type
		w.node(n.Elem)
		return true
	default:
		return false
	}
}
