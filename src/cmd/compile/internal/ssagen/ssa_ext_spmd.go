// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssagen

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/ssa"
	"cmd/internal/src"
)

// spmdForStmt generates SSA for an SPMD "go for" loop.
//
// Phase 1.10c: Vectorized loop structure. The walk/range.go phase has
// already lowered the range clause into Init/Cond/Post/Body with the correct
// laneCount stride. We generate the same loop structure as scalar loops:
//
//	Init: hv1 = 0; hn = N                          (from walk)
//	Cond: hv1 < hn                                  (from walk, bounds check)
//	Body: i = OSPMDAdd(OSPMDSplat(hv1), OSPMDLaneIndex()) (from walk)
//	      user code                                 (from walk)
//	Post: hv1 = hv1 + laneCount                     (from walk)
//
// When N is not a multiple of laneCount, the last iteration processes
// fewer than laneCount elements. Tail masking will be added in Phase 1.10e.
func (s *state) spmdForStmt(n *ir.ForStmt) {
	laneCount := n.LaneCount
	if laneCount <= 1 {
		// No vectorization benefit, fall back to scalar loop.
		s.scalarForStmt(n)
		return
	}

	bCond := s.f.NewBlock(ssa.BlockPlain)
	bBody := s.f.NewBlock(ssa.BlockPlain)
	bIncr := s.f.NewBlock(ssa.BlockPlain)
	bEnd := s.f.NewBlock(ssa.BlockPlain)

	// ensure empty for loops have correct position; issue #30167
	bBody.Pos = n.Pos()

	// first, jump to condition test
	b := s.endBlock()
	b.AddEdgeTo(bCond)

	// generate code to test condition
	s.startBlock(bCond)
	if n.Cond != nil {
		s.condBranch(n.Cond, bBody, bEnd, 1)
	} else {
		b := s.endBlock()
		b.Kind = ssa.BlockPlain
		b.AddEdgeTo(bBody)
	}

	// set up for continue/break in body
	prevContinue := s.continueTo
	prevBreak := s.breakTo
	s.continueTo = bIncr
	s.breakTo = bEnd
	var lab *ssaLabel
	if sym := n.Label; sym != nil {
		lab = s.label(sym)
		lab.continueTarget = bIncr
		lab.breakTarget = bEnd
	}

	// generate body
	s.startBlock(bBody)
	s.stmtList(n.Body)

	// tear down continue/break
	s.continueTo = prevContinue
	s.breakTo = prevBreak
	if lab != nil {
		lab.continueTarget = nil
		lab.breakTarget = nil
	}

	// done with body, goto incr
	if b := s.endBlock(); b != nil {
		b.AddEdgeTo(bIncr)
	}

	// generate incr (walk already set correct stride)
	s.startBlock(bIncr)
	if n.Post != nil {
		s.stmt(n.Post)
	}
	if b := s.endBlock(); b != nil {
		b.AddEdgeTo(bCond)
		// It can happen that bIncr ends in a block containing only VARKILL,
		// and that muddles the debugging experience.
		if b.Pos == src.NoXPos {
			b.Pos = bCond.Pos
		}
	}

	s.startBlock(bEnd)
}

// scalarForStmt generates a standard scalar for loop (same as normal OFOR).
// Used as fallback when laneCount <= 1 or vectorization is not beneficial.
func (s *state) scalarForStmt(n *ir.ForStmt) {
	bCond := s.f.NewBlock(ssa.BlockPlain)
	bBody := s.f.NewBlock(ssa.BlockPlain)
	bIncr := s.f.NewBlock(ssa.BlockPlain)
	bEnd := s.f.NewBlock(ssa.BlockPlain)

	bBody.Pos = n.Pos()

	b := s.endBlock()
	b.AddEdgeTo(bCond)

	s.startBlock(bCond)
	if n.Cond != nil {
		s.condBranch(n.Cond, bBody, bEnd, 1)
	} else {
		b := s.endBlock()
		b.Kind = ssa.BlockPlain
		b.AddEdgeTo(bBody)
	}

	prevContinue := s.continueTo
	prevBreak := s.breakTo
	s.continueTo = bIncr
	s.breakTo = bEnd
	var lab *ssaLabel
	if sym := n.Label; sym != nil {
		lab = s.label(sym)
		lab.continueTarget = bIncr
		lab.breakTarget = bEnd
	}

	s.startBlock(bBody)
	s.stmtList(n.Body)

	s.continueTo = prevContinue
	s.breakTo = prevBreak
	if lab != nil {
		lab.continueTarget = nil
		lab.breakTarget = nil
	}

	if b := s.endBlock(); b != nil {
		b.AddEdgeTo(bIncr)
	}

	s.startBlock(bIncr)
	if n.Post != nil {
		s.stmt(n.Post)
	}
	if b := s.endBlock(); b != nil {
		b.AddEdgeTo(bCond)
		if b.Pos == src.NoXPos {
			b.Pos = bCond.Pos
		}
	}

	s.startBlock(bEnd)
}
