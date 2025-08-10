// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends stmt.go to implement SPMD type checking rules.

package types2

import (
	"cmd/compile/internal/syntax"
	"go/constant"
	"go/token"
	"internal/buildcfg"
	. "internal/types/errors"
)

// SPMD statement context flags
const (
	// SPMD context flags (extending stmtContext)
	inSPMDFor        stmtContext = 1 << (iota + 8) // inside SPMD go for loop
	varyingCondition                               // inside varying if statement
)

// SPMDControlFlowInfo tracks SPMD control flow context following ISPC approach
type SPMDControlFlowInfo struct {
	inSPMDLoop        bool // inside SPMD go for loop
	varyingDepth      int  // depth of nested varying if statements
	maskAltered       bool // true if any continue in varying context has occurred
	hasVaryingParams  bool // current function has varying parameters
}

// handleSPMDStatement processes SPMD-specific statement validation
func (check *Checker) handleSPMDStatement(s syntax.Stmt, ctxt stmtContext) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}

	switch s := s.(type) {
	case *syntax.ForStmt:
		if s.IsSpmd {
			check.spmdForStmt(s, ctxt)
			return true
		}
		// If this is a regular for loop inside SPMD context, mark it
		if ctxt&inSPMDFor != 0 {
			// Let the regular for loop handle itself, but mark the context
			return false // Don't handle here, let normal for processing continue
		}
	case *syntax.BranchStmt:
		if ctxt&inSPMDFor != 0 {
			// Check for goto statements in SPMD context
			if s.Tok == syntax.Goto {
				check.error(s, InvalidSPMDGoto, "goto statements not supported in SPMD context")
				return true
			}
			// Handle continue statements first for mask alteration tracking
			if s.Tok == syntax.Continue {
				// Track mask alteration when continue occurs in varying context
				if globalSPMDInfo.varyingDepth > 0 {
					globalSPMDInfo.maskAltered = true
				}
				// Continue is always allowed - let regular handling proceed
				return false
			}
			
			// Only apply SPMD restrictions to breaks that would escape regular control structures
			// If breakOk is set, let the regular Go logic handle them first
			if s.Tok == syntax.Break && ctxt&breakOk != 0 {
				return false // Let regular break handling proceed
			}
			// Apply SPMD restrictions for breaks/continues that would target the SPMD loop
			check.validateSPMDBranch(s, ctxt)
			return true
		}
	case *syntax.IfStmt:
		if ctxt&inSPMDFor != 0 {
			check.spmdIfStmt(s, ctxt)
			return true
		}
	case *syntax.ReturnStmt:
		if ctxt&inSPMDFor != 0 {
			check.validateSPMDReturn(s, ctxt)
			return true
		}
	case *syntax.SwitchStmt:
		if ctxt&inSPMDFor != 0 {
			check.spmdSwitchStmt(s, ctxt)
			return true
		}
	case *syntax.SelectStmt:
		if ctxt&inSPMDFor != 0 {
			check.error(s, InvalidSPMDSelect, "select statements not supported in SPMD context")
			return true
		}
	case *syntax.LabeledStmt:
		if ctxt&inSPMDFor != 0 {
			// Only reject labels that are goto targets, not labels for break/continue
			// Labels for control structures (for/switch) are allowed
			if _, isControlStruct := s.Stmt.(*syntax.ForStmt); !isControlStruct {
				if _, isSwitch := s.Stmt.(*syntax.SwitchStmt); !isSwitch {
					// This is a standalone label (goto target) - not allowed in SPMD
					check.error(s, InvalidSPMDGoto, "goto statements not supported in SPMD context")
					return true
				}
			}
		}
	}

	return false
}

// spmdForStmt validates SPMD for loop statements
func (check *Checker) spmdForStmt(s *syntax.ForStmt, ctxt stmtContext) {
	// Check for nested go for loops
	if ctxt&inSPMDFor != 0 {
		check.error(s, InvalidNestedSPMDFor, "nested `go for` loop (prohibited for now)")
		return
	}

	// Set SPMD context for inner statements
	inner := ctxt | continueOk | inSPMDFor
	// Note: breakOk is conditionally set based on varying control flow

	// Initialize SPMD control flow tracking
	oldSPMDInfo := globalSPMDInfo
	globalSPMDInfo = SPMDControlFlowInfo{
		inSPMDLoop:       true,
		varyingDepth:     0,
		maskAltered:      false,
		hasVaryingParams: oldSPMDInfo.hasVaryingParams, // Preserve from outer scope
	}
	defer func() { globalSPMDInfo = oldSPMDInfo }()

	// Handle SPMD range clauses specially
	if rclause, _ := s.Init.(*syntax.RangeClause); rclause != nil {
		// extract sKey, sValue, sExtra from the range clause
		sKey := rclause.Lhs            // possibly nil
		var sValue, sExtra syntax.Expr // possibly nil
		if p, _ := sKey.(*syntax.ListExpr); p != nil {
			if len(p.ElemList) < 2 {
				check.error(s, InvalidSyntaxTree, "invalid lhs in range clause")
				return
			}
			// len(p.ElemList) >= 2
			sKey = p.ElemList[0]
			sValue = p.ElemList[1]
			if len(p.ElemList) > 2 {
				// delay error reporting until we know more
				sExtra = p.ElemList[2]
			}
		}
		// Use SPMD-specific range statement handling
		check.spmdRangeStmt(inner, s, s, sKey, sValue, sExtra, rclause.X, rclause.Def, rclause.Constraint)
		return // Don't process body again
	} else {
		// Process the SPMD loop body with regular for scope
		check.openScope(s, "for")
		defer check.closeScope()

		// Handle regular init/cond/post statements
		if s.Init != nil {
			check.simpleStmt(s.Init)
		}

		if s.Cond != nil {
			var x operand
			check.expr(nil, &x, s.Cond)
			if x.mode == invalid {
				return
			}
			if !allBoolean(x.typ) {
				check.error(&x, InvalidCond, "non-boolean condition in for statement")
			}
		}

		if s.Post != nil {
			check.simpleStmt(s.Post)
		}

		// Check SPMD loop body with SPMD context
		// Process block statements to track mask alteration across statements
		blockStmt := s.Body
		check.openScope(blockStmt, "block")
		defer check.closeScope()
		check.stmtList(inner, blockStmt.List)
	}
}

// spmdRangeStmt type-checks SPMD range statements with constraint support
func (check *Checker) spmdRangeStmt(inner stmtContext, s, rangeStmt syntax.Stmt, sKey, sValue, sExtra, x syntax.Expr, def bool, constraint syntax.Expr) {
	// For Phase 1.4, handle SPMD range clauses similarly to regular range clauses
	// but with SPMD-specific variable typing
	
	// Type-check the range expression
	var expr operand
	check.expr(nil, &expr, x)
	if expr.mode == invalid {
		return
	}
	
	// Handle constraint if present (for varying[n] syntax)
	if constraint != nil {
		var constraintOp operand
		check.expr(nil, &constraintOp, constraint)
		if constraintOp.mode != constant_ {
			check.error(constraint, InvalidConstVal, "constraint must be a constant")
			return
		}
		// TODO: Validate constraint value and use it for SPMD code generation
	}

	// Open scope for SPMD range variables (following regular range pattern)
	check.openScope(s.(*syntax.ForStmt), "range")
	defer check.closeScope()

	// Create SPMD-typed range variables following the regular rangeStmt pattern
	// In SPMD context, range variables should be varying by default
	if def {
		// Short variable declaration (:=)
		var vars []*Var
		lhs := [2]syntax.Expr{sKey, sValue} // sKey, sValue may be nil
		
		for i, lhsExpr := range lhs {
			if lhsExpr == nil {
				continue
			}
			
			// determine lhs variable
			var obj *Var
			if ident, _ := lhsExpr.(*syntax.Name); ident != nil {
				// declare new variable
				name := ident.Value
				obj = newVar(LocalVar, ident.Pos(), check.pkg, name, nil)
				check.recordDef(ident, obj)
				// _ variables don't count as new variables
				if name != "_" {
					vars = append(vars, obj)
				}
			} else {
				check.errorf(lhsExpr, InvalidSyntaxTree, "cannot declare %s", lhsExpr)
				obj = newVar(LocalVar, lhsExpr.Pos(), check.pkg, "_", nil) // dummy variable
			}
			assert(obj.typ == nil)
			
			// Set SPMD varying types for iteration variables
			if i == 0 && sKey != nil {
				// Key is varying int in SPMD range
				obj.typ = NewVarying(Typ[Int])
			} else if i == 1 && sValue != nil {
				// Value is varying element type in SPMD range
				obj.typ = NewVarying(expr.typ)
			}
			assert(obj.typ != nil)
		}
		
		// declare variables in scope
		if len(vars) > 0 {
			scopePos := s.(*syntax.ForStmt).Body.Pos()
			for _, obj := range vars {
				check.declare(check.scope, nil /* recordDef already called */, obj, scopePos)
			}
		} else {
			check.error(s, NoNewVar, "no new variables on left side of :=")
		}
	} else {
		// Regular assignment (=) - not yet implemented for SPMD
		check.error(s, InvalidSyntaxTree, "SPMD range with assignment not yet supported")
	}
	
	if sExtra != nil {
		check.error(sExtra, InvalidIterVar, "too many variables in range clause")
	}
	
	// Type-check the loop body with SPMD context
	// Process block statements to track mask alteration across statements
	forStmt := s.(*syntax.ForStmt)
	blockStmt := forStmt.Body
	check.openScope(blockStmt, "block")
	defer check.closeScope()
	check.stmtList(inner, blockStmt.List)
}

// spmdIfStmt handles if statements within SPMD context
func (check *Checker) spmdIfStmt(s *syntax.IfStmt, ctxt stmtContext) {
	// Check if the condition is varying
	var x operand
	check.expr(nil, &x, s.Cond)

	isVaryingCondition := false
	if x.mode != invalid && x.typ != nil {
		if spmdType, ok := x.typ.(*SPMDType); ok && spmdType.qualifier == VaryingQualifier {
			isVaryingCondition = true
		}
	}

	// Track varying depth
	if isVaryingCondition {
		globalSPMDInfo.varyingDepth++
		defer func() { globalSPMDInfo.varyingDepth-- }()
	}

	// Set break/continue permissions based on varying depth
	inner := ctxt
	if globalSPMDInfo.varyingDepth > 0 || globalSPMDInfo.maskAltered {
		// Remove break permission in varying context
		inner &^= breakOk
	}

	// Process if statement branches
	check.stmt(inner, s.Then)
	if s.Else != nil {
		check.stmt(inner, s.Else)
	}
}

// validateSPMDBranch validates break/continue/return statements in SPMD context
func (check *Checker) validateSPMDBranch(s *syntax.BranchStmt, ctxt stmtContext) {
	// Only apply SPMD restrictions to statements that target the SPMD loop
	// Regular for loops inside SPMD go for loops should follow normal Go rules
	
	switch s.Tok {
	case syntax.Break, syntax.Continue:
		// Check if this break/continue targets the SPMD go for loop
		// If it has a label, we'd need to check what it targets
		// For unlabeled breaks/continues, they target the innermost loop
		
		// If we're inside a regular for loop, the break/continue targets that loop, not the SPMD loop
		// So we should only apply SPMD restrictions if there's no intervening regular for loop
		
		// For now, only apply restrictions when we're in varying context
		// This allows regular for loops with uniform conditions to work normally
		if s.Tok == syntax.Break && (globalSPMDInfo.varyingDepth > 0 || globalSPMDInfo.maskAltered) {
			if globalSPMDInfo.maskAltered {
				check.error(s, InvalidSPMDBreak, "break statement not allowed after continue in varying context in SPMD for loop")
			} else {
				check.error(s, InvalidSPMDBreak, "break statement not allowed under varying conditions in SPMD for loop")
			}
		}
		
		// Continue statements are handled earlier in handleSPMDStatement
		// No additional processing needed here
	}
}

// validateSPMDReturn validates return statements in SPMD context
func (check *Checker) validateSPMDReturn(s *syntax.ReturnStmt, ctxt stmtContext) {
	if ctxt&inSPMDFor != 0 && (globalSPMDInfo.varyingDepth > 0 || globalSPMDInfo.maskAltered) {
		if globalSPMDInfo.maskAltered {
			check.error(s, InvalidSPMDReturn, "return statement not allowed after continue in varying context in SPMD for loop")
		} else {
			check.error(s, InvalidSPMDReturn, "return statement not allowed under varying conditions in SPMD for loop")
		}
	}
}

// validateSPMDFunction checks SPMD function restrictions
func (check *Checker) validateSPMDFunction(name *syntax.Name, sig *Signature, body *syntax.BlockStmt) {
	if !buildcfg.Experiment.SPMD {
		return
	}

	// Check if function has varying parameters (is SPMD function)
	hasSPMDParams := check.hasSPMDParameters(sig)

	if hasSPMDParams {
		// Private function restriction: public functions cannot have varying parameters
		// Exception: lanes and reduce packages are allowed to export SPMD functions
		if name != nil && name.Value != "" && token.IsExported(name.Value) {
			pkgPath := check.pkg.path
			if pkgPath != "lanes" && pkgPath != "reduce" {
				check.error(name, InvalidSPMDFunction, "public functions cannot have varying parameters (except in lanes/reduce packages)")
			}
		}

		// SPMD functions cannot contain go for loops
		if body != nil {
			check.checkNoGoForInSPMDFunction(body)
		}
	}
}

// hasSPMDParameters checks if function signature has varying parameters
func (check *Checker) hasSPMDParameters(sig *Signature) bool {
	if sig.params == nil {
		return false
	}

	for _, param := range sig.params.vars {
		if spmdType, ok := param.typ.(*SPMDType); ok && spmdType.qualifier == VaryingQualifier {
			return true
		}
	}
	return false
}

// checkNoGoForInSPMDFunction validates that SPMD functions don't contain go for loops
func (check *Checker) checkNoGoForInSPMDFunction(body *syntax.BlockStmt) {
	syntax.Inspect(body, func(n syntax.Node) bool {
		if forStmt, ok := n.(*syntax.ForStmt); ok && forStmt.IsSpmd {
			check.error(forStmt, InvalidSPMDFunction, "functions with varying parameters cannot contain go for loops")
		}
		return true
	})
}

// hasGoForInSPMDFunction checks if SPMD functions contain go for loops (returns boolean)
func (check *Checker) hasGoForInSPMDFunction(body *syntax.BlockStmt) bool {
	hasGoFor := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if forStmt, ok := n.(*syntax.ForStmt); ok && forStmt.IsSpmd {
			hasGoFor = true
			return false // Stop inspection once found
		}
		return true
	})
	return hasGoFor
}

// spmdSwitchStmt handles switch statements in SPMD context
func (check *Checker) spmdSwitchStmt(s *syntax.SwitchStmt, ctxt stmtContext) {
	// Check if this is a type switch first (just like regular switch processing)
	if g, _ := s.Tag.(*syntax.TypeSwitchGuard); g != nil {
		// This is a type switch - delegate to regular type switch processing with SPMD context
		check.typeSwitchStmt(ctxt|inTypeSwitch, s, g)
		return
	}
	
	// This is an expression switch - handle SPMD-specific logic
	var x operand
	if s.Tag != nil {
		check.expr(nil, &x, s.Tag)
		// By checking assignment of x to an invisible temporary
		// (as a compiler would), we get all the relevant checks.
		check.assignment(&x, nil, "switch expression")
		if x.mode != invalid && !Comparable(x.typ) && !hasNil(x.typ) {
			check.errorf(&x, InvalidExprSwitch, "cannot switch on %s (%s is not comparable)", &x, x.typ)
			x.mode = invalid
		}
	} else {
		// spec: "A missing switch expression is
		// equivalent to the boolean value true."
		x.mode = constant_
		x.typ = Typ[Bool]
		x.val = constant.MakeBool(true)
		// TODO(gri) should have a better position here
		pos := s.Rbrace
		if len(s.Body) > 0 {
			pos = s.Body[0].Pos()
		}
		x.expr = syntax.NewName(pos, "true")
	}

	// Check if the switch expression is varying
	isVaryingSwitch := check.isVaryingOperand(&x)

	// Handle SPMD-specific switch restrictions
	check.multipleSwitchDefaults(s.Body)

	// Track the original varying depth
	originalVaryingDepth := globalSPMDInfo.varyingDepth
	
	// If this is a varying switch, increment varying depth
	if isVaryingSwitch {
		globalSPMDInfo.varyingDepth++
	}
	
	defer func() {
		// Restore varying depth when exiting switch
		globalSPMDInfo.varyingDepth = originalVaryingDepth
	}()

	seen := make(valueMap) // map of seen case values to positions and types
	for i, clause := range s.Body {
		if clause == nil {
			check.error(clause, InvalidSyntaxTree, "incorrect expression switch case")
			continue
		}
		inner := ctxt
		if i+1 < len(s.Body) {
			inner |= fallthroughOk
		} else {
			inner |= finalSwitchCase
		}
		check.caseValues(&x, syntax.UnpackListExpr(clause.Cases), seen)
		check.openScope(clause, "case")
		check.stmtList(inner, clause.Body)
		check.closeScope()
	}
}

// Add SPMD info to Checker if not already present
func (check *Checker) initSPMDInfo() {
	// Initialize spmdInfo field if it doesn't exist
	// This would typically be added to the Checker struct
}
