//go:build !goexperiment.spmd

// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

// handleSPMDTypeIdentical stub for when SPMD experiment is disabled.
func (c *comparer) handleSPMDTypeIdentical(x, y Type, p *ifacePair) (bool, bool) {
	// SPMD types don't exist when experiment is disabled
	return false, false
}