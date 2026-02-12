// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "internal/buildcfg"

func (c *comparer) handleSPMDTypeIdentical(x, y Type, p *ifacePair) (handled, identical bool) {
	if !buildcfg.Experiment.SPMD {
		return false, false
	}

	if spmdX, ok := x.(*SPMDType); ok {
		if spmdY, ok := y.(*SPMDType); ok {
			return true, (spmdX.qualifier == spmdY.qualifier &&
				spmdX.constraint == spmdY.constraint &&
				c.identical(spmdX.elem, spmdY.elem, p))
		}
		return true, false
	}
	if _, ok := y.(*SPMDType); ok {
		return true, false
	}
	return false, false
}
