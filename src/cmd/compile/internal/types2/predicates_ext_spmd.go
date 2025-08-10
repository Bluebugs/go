// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends predicates.go to handle SPMD types.

package types2

import "internal/buildcfg"

// handleSPMDTypeIdentical handles SPMD type identity comparison.
// Returns (handled, identical) where handled indicates if SPMD types were compared,
// and identical indicates if they are identical.
func (c *comparer) handleSPMDTypeIdentical(x, y Type, p *ifacePair) (handled, identical bool) {
	if !buildcfg.Experiment.SPMD {
		return false, false
	}

	if spmdX, ok := x.(*SPMDType); ok {
		if spmdY, ok := y.(*SPMDType); ok {
			// Two SPMD types are identical if they have the same qualifier, constraint, and element type
			return true, (spmdX.qualifier == spmdY.qualifier &&
				spmdX.constraint == spmdY.constraint &&
				c.identical(spmdX.elem, spmdY.elem, p))
		}
		// One is SPMD, the other isn't - they're not identical
		return true, false
	}
	if _, ok := y.(*SPMDType); ok {
		// y is SPMD but x isn't - they're not identical
		return true, false
	}
	// Neither is SPMD type
	return false, false
}
