// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssagen

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"strings"
)

// spmdLoopMaskState tracks per-lane continue/break masks for loops inside SPMD context.
type spmdLoopMaskState struct {
	continueMaskVar ir.Node            // synthetic var for continue mask (reset per iteration)
	breakMaskVar    ir.Node            // synthetic var for break mask (persists across iterations), nil for go for
	entryMask       *ssa.Value         // mask on loop entry
	parent          *spmdLoopMaskState // for nested loops
}

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
// fewer than laneCount elements. spmdBodyWithTailMask handles tail masking
// by computing a per-iteration mask (laneIndex < bound) to disable
// out-of-bounds lanes.
func (s *state) spmdForStmt(n *ir.ForStmt) {
	laneCount := n.LaneCount
	if laneCount <= 1 {
		// No vectorization benefit, fall back to scalar loop.
		s.scalarForStmt(n)
		return
	}

	// Save and set SPMD context
	prevInSPMD := s.inSPMDLoop
	prevMask := s.spmdMask
	s.inSPMDLoop = true
	// Initialize all-true mask: SPMDSplat(true)
	boolType := types.Types[types.TBOOL]
	s.spmdMask = s.newValue1(ssa.OpSPMDSplat, boolType, s.constBool(true))

	// Set up continue mask for go for body (break is forbidden under varying in go for)
	continueMaskVar := typecheck.TempAt(n.Pos(), s.curfn, boolType)
	prevLoopMasks := s.spmdLoopMasks
	s.spmdLoopMasks = &spmdLoopMaskState{
		continueMaskVar: continueMaskVar,
		entryMask:       s.spmdMask,
		parent:          prevLoopMasks,
	}

	defer func() {
		s.inSPMDLoop = prevInSPMD
		s.spmdMask = prevMask
		s.spmdLoopMasks = prevLoopMasks
	}()

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
	s.spmdBodyWithTailMask(n)

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

// spmdBodyWithTailMask generates the SPMD loop body with per-iteration tail masking.
// For a loop `go for i := range N` with laneCount L, when N is not a multiple of L,
// the last iteration has lane indices that exceed N. This function computes a tail
// mask (laneIndex < bound) each iteration and applies it to the execution mask so
// out-of-bounds lanes are disabled.
//
// The body generation proceeds as:
//  1. Reset s.spmdMask to all-true (clean slate each iteration)
//  2. Execute n.Body[0] — the SPMD index assignment (i = SPMDAdd(SPMDSplat(hv1), SPMDLaneIndex()))
//  3. Compute tailMask = SPMDLess(laneIndices, SPMDSplat(bound))
//  4. Apply s.spmdMask = SPMDMaskAnd(allTrue, tailMask)
//  5. Execute n.Body[1:] — remaining user code with correct mask
func (s *state) spmdBodyWithTailMask(n *ir.ForStmt) {
	boolType := types.Types[types.TBOOL]

	// Reset mask to all-true at start of each iteration.
	s.spmdMask = s.newValue1(ssa.OpSPMDSplat, boolType, s.constBool(true))

	// Reset continue mask for this iteration
	if s.spmdLoopMasks != nil {
		s.vars[s.spmdLoopMasks.continueMaskVar] = s.newValue1(ssa.OpSPMDSplat, boolType, s.constBool(false))
	}

	if n.Cond == nil || len(n.Body) == 0 {
		s.stmtList(n.Body)
		return
	}

	// Execute the first body statement (SPMD index assignment: i = SPMDAdd(SPMDSplat(hv1), SPMDLaneIndex()))
	firstStmt := n.Body[0]
	s.stmt(firstStmt)

	// Validate SPMD assignment pattern: i = SPMDAdd(SPMDSplat(hv1), SPMDLaneIndex())
	// If the pattern doesn't match, skip tail masking (safe fallback: all-true mask).
	assignStmt, ok := firstStmt.(*ir.AssignStmt)
	if !ok || assignStmt.Y == nil || assignStmt.Y.Op() != ir.OSPMDAdd {
		s.stmtList(n.Body[1:])
		return
	}

	// Validate condition structure: hv1 < hn (generated by walk/range.go)
	condExpr, ok := n.Cond.(*ir.BinaryExpr)
	if !ok || condExpr.Op() != ir.OLT {
		s.stmtList(n.Body[1:])
		return
	}

	// Get SSA value of the loop variable (just assigned above)
	loopVar := assignStmt.X
	iVal := s.variable(loopVar, loopVar.Type())

	// Extract bound from Cond: hv1 < hn → hn is the Y operand
	boundVal := s.expr(condExpr.Y)

	// Splat bound to all lanes and compare: tailMask = (i < bound)
	splatBound := s.newValue1(ssa.OpSPMDSplat, loopVar.Type(), boundVal)
	tailMask := s.newValue2(ssa.OpSPMDLess, boolType, iVal, splatBound)

	// Apply tail mask to execution mask
	s.spmdMask = s.newValue2(ssa.OpSPMDMaskAnd, boolType, s.spmdMask, tailMask)

	// Execute remaining body statements with tail-masked execution
	s.stmtList(n.Body[1:])
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

// spmdIfStmt generates SSA for a varying-condition if/else inside an SPMD loop.
// Instead of branching, both branches execute with different lane masks,
// and modified variables are merged using SPMDSelect.
func (s *state) spmdIfStmt(n *ir.IfStmt) {
	// Evaluate the varying condition (including any init statements)
	s.stmtList(n.Cond.Init())
	condValue := s.expr(n.Cond)
	boolType := types.Types[types.TBOOL]

	savedMask := s.spmdMask

	// Compute branch masks
	trueMask := s.newValue2(ssa.OpSPMDMaskAnd, boolType, savedMask, condValue)
	falseMask := s.newValue2(ssa.OpSPMDMaskAndNot, boolType, savedMask, condValue)

	// Snapshot vars before branches
	snapshot := s.snapshotVars()

	s.spmdVaryingDepth++

	// Execute true branch with trueMask
	s.spmdMask = trueMask
	s.stmtList(n.Body)
	trueVars := s.snapshotVars()

	// Restore to pre-branch state, execute false branch with falseMask
	s.restoreVarsSnapshot(snapshot)
	s.spmdMask = falseMask
	if len(n.Else) > 0 {
		s.stmtList(n.Else)
	}
	// s.vars now holds false-branch state

	// Merge modified variables using SPMDSelect
	s.spmdMergeVars(condValue, snapshot, trueVars)

	s.spmdVaryingDepth--

	// Restore mask, excluding lanes that continued/broke during this if
	s.spmdMask = savedMask
	s.spmdExcludeBranchMasks()
}

// snapshotVars returns a shallow copy of the current variable map.
func (s *state) snapshotVars() map[ir.Node]*ssa.Value {
	snap := make(map[ir.Node]*ssa.Value, len(s.vars))
	for k, v := range s.vars {
		snap[k] = v
	}
	return snap
}

// restoreVarsSnapshot restores s.vars from a snapshot.
func (s *state) restoreVarsSnapshot(snap map[ir.Node]*ssa.Value) {
	s.vars = make(map[ir.Node]*ssa.Value, len(snap))
	for k, v := range snap {
		s.vars[k] = v
	}
}

// spmdMergeVars merges variables after both branches of a varying if.
// condValue is the varying bool condition.
// snapshot is the pre-branch variable state.
// trueVars is the variable state after the true branch.
// s.vars currently holds the false-branch state.
//
// For each variable modified in either branch:
//   merged = SPMDSelect(cond, trueVal, falseVal)
func (s *state) spmdMergeVars(condValue *ssa.Value, snapshot, trueVars map[ir.Node]*ssa.Value) {
	// Collect all variables that were modified in either branch
	for varNode, trueVal := range trueVars {
		origVal := snapshot[varNode]
		falseVal := s.vars[varNode]
		if falseVal == nil {
			falseVal = origVal
		}
		if origVal == nil && falseVal == nil {
			// New variable only exists in true branch; not visible after merge
			continue
		}
		if trueVal == origVal && falseVal == origVal {
			continue // not modified in either branch
		}
		if trueVal == falseVal {
			s.vars[varNode] = trueVal // same value, no select needed
			continue
		}
		// Different values: merge with SPMDSelect(cond, trueVal, falseVal)
		s.vars[varNode] = s.newValue3(ssa.OpSPMDSelect, trueVal.Type, condValue, trueVal, falseVal)
	}
	// Also check false-only modifications (vars modified in false branch but not in true)
	for varNode, falseVal := range s.vars {
		if _, inTrue := trueVars[varNode]; inTrue {
			continue // already handled above
		}
		origVal := snapshot[varNode]
		if falseVal == origVal {
			continue // not modified
		}
		trueVal := origVal
		if trueVal == nil {
			continue // new variable only in false branch, keep it
		}
		s.vars[varNode] = s.newValue3(ssa.OpSPMDSelect, falseVal.Type, condValue, trueVal, falseVal)
	}
}

// spmdSwitchStmt generates SSA for a varying-condition switch inside an SPMD loop.
// All case branches execute with per-lane masking; modified variables are merged
// using SPMDSelect, following the same pattern as spmdIfStmt.
func (s *state) spmdSwitchStmt(n *ir.SwitchStmt) {
	savedMask := s.spmdMask
	boolType := types.Types[types.TBOOL]

	// Evaluate the tag expression (varying value being switched on)
	tagVal := s.expr(n.Tag)

	// Snapshot vars before any case body execution
	snapshot := s.snapshotVars()

	// Phase 1: Compute per-case masks (mutually exclusive).
	// remainingMask tracks lanes not yet claimed by any case.
	remainingMask := savedMask
	caseMasks := make([]*ssa.Value, len(n.Cases))
	defaultIdx := -1

	for i, clause := range n.Cases {
		if len(clause.List) == 0 {
			// default case — gets remaining mask after all explicit cases
			defaultIdx = i
			continue
		}

		// Build OR of equalities for multi-value cases: case val1, val2, ...
		var caseCond *ssa.Value
		for _, caseExpr := range clause.List {
			caseVal := s.expr(caseExpr)
			// Determine if the case value is varying (produced by SPMD ops)
			// or scalar (needs splatting to all lanes).
			var eq *ssa.Value
			if isVaryingSPMDValue(caseVal) {
				// Varying case value: compare per-lane directly
				eq = s.newValue2(ssa.OpSPMDEqual, boolType, tagVal, caseVal)
			} else {
				// Scalar case value: splat to all lanes then compare
				splatVal := s.newValue1(ssa.OpSPMDSplat, tagVal.Type, caseVal)
				eq = s.newValue2(ssa.OpSPMDEqual, boolType, tagVal, splatVal)
			}
			if caseCond == nil {
				caseCond = eq
			} else {
				caseCond = s.newValue2(ssa.OpSPMDMaskOr, boolType, caseCond, eq)
			}
		}

		// Intersect with remaining mask (lanes not yet claimed)
		caseMasks[i] = s.newValue2(ssa.OpSPMDMaskAnd, boolType, remainingMask, caseCond)
		// Remove claimed lanes from remaining
		remainingMask = s.newValue2(ssa.OpSPMDMaskAndNot, boolType, remainingMask, caseCond)
	}

	if defaultIdx >= 0 {
		caseMasks[defaultIdx] = remainingMask
	}

	// Phase 2: Execute each case body and accumulate merged results.
	// Start with the pre-switch snapshot as the accumulated state.
	accumulated := s.snapshotVars()

	for i, clause := range n.Cases {
		if caseMasks[i] == nil {
			continue // case with no mask (shouldn't happen, but be defensive)
		}

		// Each case starts from pre-switch state
		s.restoreVarsSnapshot(snapshot)
		s.spmdMask = caseMasks[i]

		// Execute case body
		s.spmdVaryingDepth++
		s.stmtList(clause.Body)
		s.spmdVaryingDepth--

		// Merge: for each var modified by this case, select between case value and accumulated value
		for varNode, caseVal := range s.vars {
			accVal := accumulated[varNode]
			if accVal == nil {
				// New variable only in this case; carry forward
				accumulated[varNode] = caseVal
				continue
			}
			if caseVal == accVal {
				continue // not modified
			}
			accumulated[varNode] = s.newValue3(ssa.OpSPMDSelect, caseVal.Type, caseMasks[i], caseVal, accVal)
		}
	}

	s.vars = accumulated
	s.spmdMask = savedMask
	s.spmdExcludeBranchMasks()
}

// isVaryingSPMDValue reports whether v represents a varying (per-lane) value
// by checking if it was produced by an SPMD operation.
func isVaryingSPMDValue(v *ssa.Value) bool {
	switch v.Op {
	case ssa.OpSPMDSplat, ssa.OpSPMDLaneIndex, ssa.OpSPMDLaneCount,
		ssa.OpSPMDAdd, ssa.OpSPMDSub, ssa.OpSPMDMul, ssa.OpSPMDDiv, ssa.OpSPMDMod,
		ssa.OpSPMDNeg, ssa.OpSPMDAnd, ssa.OpSPMDOr, ssa.OpSPMDXor, ssa.OpSPMDNot,
		ssa.OpSPMDShl, ssa.OpSPMDShr,
		ssa.OpSPMDEqual, ssa.OpSPMDNotEqual,
		ssa.OpSPMDLess, ssa.OpSPMDLessEqual, ssa.OpSPMDGreater, ssa.OpSPMDGreaterEqual,
		ssa.OpSPMDSelect,
		ssa.OpSPMDMaskAnd, ssa.OpSPMDMaskOr, ssa.OpSPMDMaskAndNot, ssa.OpSPMDMaskNot,
		ssa.OpSPMDMaskAllTrue, ssa.OpSPMDMaskAnyTrue, ssa.OpSPMDMaskAllFalse,
		ssa.OpSPMDLoad, ssa.OpSPMDStore, ssa.OpSPMDMaskedLoad, ssa.OpSPMDMaskedStore,
		ssa.OpSPMDReduceAdd, ssa.OpSPMDReduceMul, ssa.OpSPMDReduceMin, ssa.OpSPMDReduceMax,
		ssa.OpSPMDReduceAnd, ssa.OpSPMDReduceOr, ssa.OpSPMDReduceXor,
		ssa.OpSPMDBroadcastLane, ssa.OpSPMDRotate, ssa.OpSPMDSwizzle,
		ssa.OpSPMDShiftLeft, ssa.OpSPMDShiftRight:
		return true
	}
	return false
}

// spmdBuiltinCall intercepts calls to lanes and reduce package functions
// during SSA generation inside SPMD loops. If the call matches a known
// builtin, it is replaced with the corresponding SPMD SSA opcode.
// Returns nil if the call is not a recognized SPMD builtin.
func (s *state) spmdBuiltinCall(n *ir.CallExpr) *ssa.Value {
	if n.Fun.Op() != ir.ONAME {
		return nil
	}
	sym := n.Fun.(*ir.Name).Sym()
	if sym == nil || sym.Pkg == nil {
		return nil
	}
	pkg := sym.Pkg.Path
	fn := sym.Name

	// Strip type parameters from stenciled generic function names.
	// e.g. "Add[int]" -> "Add"
	if idx := strings.IndexByte(fn, '['); idx >= 0 {
		fn = fn[:idx]
	}

	// Only intercept unconstrained (constraint == -1) and universal (constraint == 0) calls.
	// Numerically constrained calls (constraint > 0) need loop decomposition (future phase).
	for _, arg := range n.Args {
		if argType := arg.Type(); argType != nil && argType.Kind() == types.TSPMD {
			if c := argType.SPMDConstraint(); c > 0 {
				return nil // constrained varying - fall through to normal call
			}
		}
	}

	switch pkg {
	case "lanes":
		return s.spmdLanesBuiltin(n, fn)
	case "reduce":
		return s.spmdReduceBuiltin(n, fn)
	}
	return nil
}

// spmdLanesBuiltin handles interception of lanes package function calls.
// Maps 7 functions to SPMD SSA opcodes:
//   - Index()           -> OpSPMDLaneIndex
//   - Count(v)          -> OpSPMDLaneCount
//   - Broadcast(v,lane) -> OpSPMDBroadcastLane
//   - Rotate(v,offset)  -> OpSPMDRotate
//   - Swizzle(v,idx)    -> OpSPMDSwizzle
//   - ShiftLeft(v,fill) -> OpSPMDShiftLeft
//   - ShiftRight(v,fill)-> OpSPMDShiftRight
func (s *state) spmdLanesBuiltin(n *ir.CallExpr, fn string) *ssa.Value {
	switch fn {
	case "Index":
		// No args, no validation needed
		return s.newValue0(ssa.OpSPMDLaneIndex, n.Type())

	case "Count":
		// Count takes 1 arg (for type inference) but ignores it
		if len(n.Args) != 1 {
			return nil
		}
		return s.newValue0(ssa.OpSPMDLaneCount, n.Type())

	case "Broadcast":
		if len(n.Args) != 2 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue2(ssa.OpSPMDBroadcastLane, n.Type(), args[0], args[1])

	case "Rotate":
		if len(n.Args) != 2 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue2(ssa.OpSPMDRotate, n.Type(), args[0], args[1])

	case "Swizzle":
		if len(n.Args) != 2 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue2(ssa.OpSPMDSwizzle, n.Type(), args[0], args[1])

	case "ShiftLeft", "shiftLeftBuiltin":
		if len(n.Args) != 2 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue2(ssa.OpSPMDShiftLeft, n.Type(), args[0], args[1])

	case "ShiftRight", "shiftRightBuiltin":
		if len(n.Args) != 2 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue2(ssa.OpSPMDShiftRight, n.Type(), args[0], args[1])

	case "broadcastBuiltin":
		if len(n.Args) != 2 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue2(ssa.OpSPMDBroadcastLane, n.Type(), args[0], args[1])

	case "rotateBuiltin":
		if len(n.Args) != 2 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue2(ssa.OpSPMDRotate, n.Type(), args[0], args[1])

	case "swizzleBuiltin":
		if len(n.Args) != 2 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue2(ssa.OpSPMDSwizzle, n.Type(), args[0], args[1])
	}
	// Unrecognized (From, FromConstrained, ToConstrained) - fall through to normal call
	return nil
}

// spmdReduceBuiltin handles interception of reduce package function calls.
// Maps 9 functions to SPMD SSA opcodes:
//   - Add(v)  -> OpSPMDReduceAdd
//   - Mul(v)  -> OpSPMDReduceMul
//   - Max(v)  -> OpSPMDReduceMax
//   - Min(v)  -> OpSPMDReduceMin
//   - Or(v)   -> OpSPMDReduceOr
//   - And(v)  -> OpSPMDReduceAnd
//   - Xor(v)  -> OpSPMDReduceXor
//   - All(v)  -> OpSPMDMaskAllTrue
//   - Any(v)  -> OpSPMDMaskAnyTrue
func (s *state) spmdReduceBuiltin(n *ir.CallExpr, fn string) *ssa.Value {
	switch fn {
	case "Add":
		if len(n.Args) != 1 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue1(ssa.OpSPMDReduceAdd, n.Type(), args[0])

	case "Mul":
		if len(n.Args) != 1 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue1(ssa.OpSPMDReduceMul, n.Type(), args[0])

	case "Max":
		if len(n.Args) != 1 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue1(ssa.OpSPMDReduceMax, n.Type(), args[0])

	case "Min":
		if len(n.Args) != 1 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue1(ssa.OpSPMDReduceMin, n.Type(), args[0])

	case "Or":
		if len(n.Args) != 1 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue1(ssa.OpSPMDReduceOr, n.Type(), args[0])

	case "And":
		if len(n.Args) != 1 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue1(ssa.OpSPMDReduceAnd, n.Type(), args[0])

	case "Xor":
		if len(n.Args) != 1 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue1(ssa.OpSPMDReduceXor, n.Type(), args[0])

	case "All":
		if len(n.Args) != 1 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue1(ssa.OpSPMDMaskAllTrue, n.Type(), args[0])

	case "Any":
		if len(n.Args) != 1 {
			return nil
		}
		args := s.intrinsicArgs(n)
		return s.newValue1(ssa.OpSPMDMaskAnyTrue, n.Type(), args[0])
	}
	// Unrecognized (From, Count, FindFirstSet, Mask) - fall through to normal call
	return nil
}

// spmdMaskedBranchStmt handles continue/break inside varying context (spmdIfStmt/spmdSwitchStmt).
// Instead of ending the block, it accumulates into mask variables and zeros spmdMask.
func (s *state) spmdMaskedBranchStmt(n *ir.BranchStmt) {
	boolType := types.Types[types.TBOOL]
	lm := s.spmdLoopMasks

	// Accumulate into continue mask (for both continue and break)
	curCont := s.variable(lm.continueMaskVar, boolType)
	newCont := s.newValue2(ssa.OpSPMDMaskOr, boolType, curCont, s.spmdMask)
	s.vars[lm.continueMaskVar] = newCont

	// For break: also accumulate into break mask (persists across iterations)
	if n.Op() == ir.OBREAK && lm.breakMaskVar != nil {
		curBreak := s.variable(lm.breakMaskVar, boolType)
		newBreak := s.newValue2(ssa.OpSPMDMaskOr, boolType, curBreak, s.spmdMask)
		s.vars[lm.breakMaskVar] = newBreak
	}

	// Zero out mask so remaining statements in this branch are skipped
	s.spmdMask = s.newValue1(ssa.OpSPMDSplat, boolType, s.constBool(false))
}

// spmdExcludeBranchMasks subtracts accumulated continue/break masks from s.spmdMask.
// Called after spmdIfStmt/spmdSwitchStmt restores the saved mask.
func (s *state) spmdExcludeBranchMasks() {
	if s.spmdLoopMasks == nil {
		return
	}
	boolType := types.Types[types.TBOOL]
	lm := s.spmdLoopMasks

	contVal := s.variable(lm.continueMaskVar, boolType)
	s.spmdMask = s.newValue2(ssa.OpSPMDMaskAndNot, boolType, s.spmdMask, contVal)

	if lm.breakMaskVar != nil {
		breakVal := s.variable(lm.breakMaskVar, boolType)
		s.spmdMask = s.newValue2(ssa.OpSPMDMaskAndNot, boolType, s.spmdMask, breakVal)
	}
}

// spmdRegularForStmt generates SSA for a regular for loop inside SPMD context.
// Sets up continue/break mask tracking. Uniform continue/break still use block jumps;
// varying continue/break (inside spmdIfStmt) use mask accumulation.
func (s *state) spmdRegularForStmt(n *ir.ForStmt) {
	boolType := types.Types[types.TBOOL]
	entryMask := s.spmdMask

	// Create synthetic variables for mask tracking
	continueMaskVar := typecheck.TempAt(n.Pos(), s.curfn, boolType)
	breakMaskVar := typecheck.TempAt(n.Pos(), s.curfn, boolType)

	// Initialize break mask to all-false before loop
	s.vars[breakMaskVar] = s.newValue1(ssa.OpSPMDSplat, boolType, s.constBool(false))

	// Push loop mask state
	prevLoopMasks := s.spmdLoopMasks
	s.spmdLoopMasks = &spmdLoopMaskState{
		continueMaskVar: continueMaskVar,
		breakMaskVar:    breakMaskVar,
		entryMask:       entryMask,
		parent:          prevLoopMasks,
	}
	defer func() { s.spmdLoopMasks = prevLoopMasks }()

	// Standard 4-block loop structure
	bCond := s.f.NewBlock(ssa.BlockPlain)
	bBody := s.f.NewBlock(ssa.BlockPlain)
	bIncr := s.f.NewBlock(ssa.BlockPlain)
	bEnd := s.f.NewBlock(ssa.BlockPlain)
	bBody.Pos = n.Pos()

	b := s.endBlock()
	b.AddEdgeTo(bCond)

	// Condition block
	s.startBlock(bCond)
	if n.Cond != nil {
		s.condBranch(n.Cond, bBody, bEnd, 1)
	} else {
		b := s.endBlock()
		b.Kind = ssa.BlockPlain
		b.AddEdgeTo(bBody)
	}

	// Set up continue/break targets (for uniform jumps)
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

	// Body block
	s.startBlock(bBody)

	// Reset continue mask and compute active mask at iteration start
	s.vars[continueMaskVar] = s.newValue1(ssa.OpSPMDSplat, boolType, s.constBool(false))
	breakVal := s.variable(breakMaskVar, boolType)
	s.spmdMask = s.newValue2(ssa.OpSPMDMaskAndNot, boolType, entryMask, breakVal)

	// Execute body
	s.stmtList(n.Body)

	// Tear down
	s.continueTo = prevContinue
	s.breakTo = prevBreak
	if lab != nil {
		lab.continueTarget = nil
		lab.breakTarget = nil
	}

	if b := s.endBlock(); b != nil {
		b.AddEdgeTo(bIncr)
	}

	// Increment block
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

	// End block: restore mask excluding broken lanes
	s.startBlock(bEnd)
	breakValEnd := s.variable(breakMaskVar, boolType)
	s.spmdMask = s.newValue2(ssa.OpSPMDMaskAndNot, boolType, entryMask, breakValEnd)
}
