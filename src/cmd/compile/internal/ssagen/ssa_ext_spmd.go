// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssagen

import (
	"cmd/compile/internal/ir"
)

// spmdForStmt generates SSA for an SPMD "go for" loop.
// Phase 1.10a: Initial stub that fatals - will be replaced with
// scalar fallback in Phase 1.10b and vectorized code in Phase 1.10c.
func (s *state) spmdForStmt(n *ir.ForStmt) {
	s.Fatalf("SPMD go for loop SSA generation not yet implemented (LaneCount=%d)", n.LaneCount)
}
