// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends stmt.go to implement SPMD type checking rules.

package types

import (
	"go/ast"
	"go/constant"
	"go/token"
	"internal/buildcfg"
	. "internal/types/errors"
)

// SPMD statement context flags
const (
	inSPMDFor        stmtContext = 1 << (iota + 8)
	varyingCondition
)

func (check *Checker) handleSPMDStatement(s ast.Stmt, ctxt stmtContext) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}

	switch s := s.(type) {
	case *ast.RangeStmt:
		if s.IsSpmd {
			check.spmdForStmt(s, ctxt)
			return true
		}
	case *ast.BranchStmt:
		if ctxt&inSPMDFor != 0 {
			if s.Tok == token.GOTO {
				check.error(s, InvalidSPMDGoto, "goto statements not supported in SPMD context")
				return true
			}
			if s.Tok == token.CONTINUE {
				if check.spmdInfo.varyingDepth > 0 {
					check.spmdInfo.maskAltered = true
				}
				return false
			}
			if s.Tok == token.BREAK && ctxt&breakOk != 0 {
				return false
			}
			check.validateSPMDBranch(s, ctxt)
			return true
		}
	case *ast.IfStmt:
		if ctxt&inSPMDFor != 0 {
			check.spmdIfStmt(s, ctxt)
			return true
		}
	case *ast.ReturnStmt:
		if ctxt&inSPMDFor != 0 {
			check.validateSPMDReturn(s, ctxt)
			return true
		}
	case *ast.SwitchStmt:
		if ctxt&inSPMDFor != 0 {
			check.spmdSwitchStmt(s, ctxt)
			return true
		}
	case *ast.SelectStmt:
		if ctxt&inSPMDFor != 0 {
			check.error(s, InvalidSPMDSelect, "select statements not supported in SPMD context")
			return true
		}
	case *ast.LabeledStmt:
		if ctxt&inSPMDFor != 0 {
			if _, isFor := s.Stmt.(*ast.ForStmt); !isFor {
				if _, isRange := s.Stmt.(*ast.RangeStmt); !isRange {
					if _, isSwitch := s.Stmt.(*ast.SwitchStmt); !isSwitch {
						check.error(s, InvalidSPMDGoto, "goto statements not supported in SPMD context")
						return true
					}
				}
			}
		}
	}

	return false
}

func (check *Checker) spmdForStmt(s *ast.RangeStmt, ctxt stmtContext) {
	if ctxt&inSPMDFor != 0 {
		check.error(s, InvalidNestedSPMDFor, "nested `go for` loop (prohibited for now)")
		return
	}

	inner := ctxt | continueOk | inSPMDFor

	oldSPMDInfo := check.spmdInfo
	check.spmdInfo = SPMDControlFlowInfo{
		inSPMDLoop:       true,
		varyingDepth:     0,
		maskAltered:      false,
		hasVaryingParams: oldSPMDInfo.hasVaryingParams,
		varyingElemSizes: nil,
	}
	defer func() { check.spmdInfo = oldSPMDInfo }()

	check.spmdRangeStmt(inner, s)
}

func (check *Checker) spmdRangeStmt(inner stmtContext, s *ast.RangeStmt) {
	var expr operand
	check.expr(nil, &expr, s.X)
	if expr.mode() == invalid {
		return
	}

	if s.Constraint != nil {
		var constraintOp operand
		check.expr(nil, &constraintOp, s.Constraint)
		if constraintOp.mode() != constant_ {
			check.error(s.Constraint, InvalidConstVal, "constraint must be a constant")
			return
		}
	}

	check.openScope(s, "range")
	defer check.closeScope()

	if s.Tok == token.DEFINE {
		var vars []*Var
		lhs := [2]ast.Expr{s.Key, s.Value}

		for i, lhsExpr := range lhs {
			if lhsExpr == nil {
				continue
			}
			var obj *Var
			if ident, _ := lhsExpr.(*ast.Ident); ident != nil {
				name := ident.Name
				obj = NewVar(ident.Pos(), check.pkg, name, nil)
				check.recordDef(ident, obj)
				if name != "_" {
					vars = append(vars, obj)
				}
			} else {
				check.errorf(lhsExpr, InvalidSyntaxTree, "cannot declare %s", lhsExpr)
				obj = NewVar(lhsExpr.Pos(), check.pkg, "_", nil)
			}

			if i == 0 && s.Key != nil {
				obj.typ = NewVarying(Typ[Int])
			} else if i == 1 && s.Value != nil {
				obj.typ = NewVarying(expr.typ())
			}
			if obj.typ == nil {
				obj.typ = Typ[Invalid]
			}
		}

		if len(vars) > 0 {
			scopePos := s.Body.Pos()
			for _, obj := range vars {
				check.declare(check.scope, nil, obj, scopePos)
			}
		} else {
			check.error(s, NoNewVar, "no new variables on left side of :=")
		}
	} else {
		check.error(s, InvalidSyntaxTree, "SPMD range with assignment not yet supported")
	}

	check.openScope(s.Body, "block")
	defer check.closeScope()
	check.stmtList(inner, s.Body.List)

	check.spmdInfo.varyingElemSizes = append(check.spmdInfo.varyingElemSizes, check.getTypeSize(Typ[Int]))
	s.LaneCount = check.computeEffectiveLaneCount(&check.spmdInfo)
}

func (check *Checker) spmdIfStmt(s *ast.IfStmt, ctxt stmtContext) {
	var x operand
	check.expr(nil, &x, s.Cond)

	isVaryingCondition := false
	if x.mode() != invalid && x.typ() != nil {
		if spmdType, ok := x.typ().(*SPMDType); ok && spmdType.qualifier == VaryingQualifier {
			isVaryingCondition = true
		}
	}

	if isVaryingCondition {
		check.spmdInfo.varyingDepth++
		defer func() { check.spmdInfo.varyingDepth-- }()
	}

	inner := ctxt
	if check.spmdInfo.varyingDepth > 0 || check.spmdInfo.maskAltered {
		inner &^= breakOk
	}

	check.stmt(inner, s.Body)
	if s.Else != nil {
		check.stmt(inner, s.Else)
	}
}

func (check *Checker) validateSPMDBranch(s *ast.BranchStmt, ctxt stmtContext) {
	switch s.Tok {
	case token.BREAK, token.CONTINUE:
		if s.Tok == token.BREAK && (check.spmdInfo.varyingDepth > 0 || check.spmdInfo.maskAltered) {
			if check.spmdInfo.maskAltered {
				check.error(s, InvalidSPMDBreak, "break statement not allowed after continue in varying context in SPMD for loop")
			} else {
				check.error(s, InvalidSPMDBreak, "break statement not allowed under varying conditions in SPMD for loop")
			}
		}
	}
}

func (check *Checker) validateSPMDReturn(s *ast.ReturnStmt, ctxt stmtContext) {
	if ctxt&inSPMDFor != 0 && (check.spmdInfo.varyingDepth > 0 || check.spmdInfo.maskAltered) {
		if check.spmdInfo.maskAltered {
			check.error(s, InvalidSPMDReturn, "return statement not allowed after continue in varying context in SPMD for loop")
		} else {
			check.error(s, InvalidSPMDReturn, "return statement not allowed under varying conditions in SPMD for loop")
		}
	}
}

func (check *Checker) validateSPMDFunction(name *ast.Ident, sig *Signature, body *ast.BlockStmt) {
	if !buildcfg.Experiment.SPMD {
		return
	}
	hasSPMDParams := check.hasSPMDParameters(sig)
	if hasSPMDParams {
		if name != nil && name.Name != "" && token.IsExported(name.Name) {
			pkgPath := check.pkg.path
			if pkgPath != "lanes" && pkgPath != "reduce" {
				check.error(name, InvalidSPMDFunc, "public functions cannot have varying parameters (except in lanes/reduce packages)")
			}
		}
		if body != nil {
			check.checkNoGoForInSPMDFunction(body)
		}
	}
}

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

func (check *Checker) checkNoGoForInSPMDFunction(body *ast.BlockStmt) {
	ast.Inspect(body, func(n ast.Node) bool {
		if rangeStmt, ok := n.(*ast.RangeStmt); ok && rangeStmt.IsSpmd {
			check.error(rangeStmt, InvalidSPMDFunc, "functions with varying parameters cannot contain go for loops")
		}
		return true
	})
}

func (check *Checker) spmdSwitchStmt(s *ast.SwitchStmt, ctxt stmtContext) {
	inner := ctxt | breakOk
	check.openScope(s, "switch")
	defer check.closeScope()

	check.simpleStmt(s.Init)

	var x operand
	if s.Tag != nil {
		check.expr(nil, &x, s.Tag)
		check.assignment(&x, nil, "switch expression")
		if x.mode() != invalid && !Comparable(x.typ()) && !hasNil(x.typ()) {
			check.errorf(&x, InvalidExprSwitch, "cannot switch on %s (%s is not comparable)", &x, x.typ())
			x.mode_ = invalid
		}
	} else {
		x.mode_ = constant_
		x.typ_ = Typ[Bool]
		x.val = constant.MakeBool(true)
		x.expr = &ast.Ident{NamePos: s.Body.Lbrace, Name: "true"}
	}

	isVaryingSwitch := check.isVaryingOperand(&x)

	originalVaryingDepth := check.spmdInfo.varyingDepth
	if isVaryingSwitch {
		check.spmdInfo.varyingDepth++
	}
	defer func() {
		check.spmdInfo.varyingDepth = originalVaryingDepth
	}()

	check.multipleDefaults(s.Body.List)

	seen := make(valueMap)
	for i, c := range s.Body.List {
		clause, _ := c.(*ast.CaseClause)
		if clause == nil {
			check.error(c, InvalidSyntaxTree, "incorrect expression switch case")
			continue
		}
		check.caseValues(&x, clause.List, seen)
		check.openScope(clause, "case")
		caseInner := inner
		if i+1 < len(s.Body.List) {
			caseInner |= fallthroughOk
		} else {
			caseInner |= finalSwitchCase
		}
		check.stmtList(caseInner, clause.Body)
		check.closeScope()
	}
}
