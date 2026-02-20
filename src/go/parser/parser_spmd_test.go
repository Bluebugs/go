// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"go/ast"
	"go/token"
	"internal/buildcfg"
	"runtime"
	"testing"
)

func setGOEXPERIMENT(goexperiment string) func() {
	exp, err := buildcfg.ParseGOEXPERIMENT(runtime.GOOS, runtime.GOARCH, goexperiment)
	if err != nil {
		panic(err)
	}
	old := buildcfg.Experiment
	buildcfg.Experiment = *exp
	return func() { buildcfg.Experiment = old }
}

func firstFuncStmt(f *ast.File) ast.Stmt {
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok {
			if len(fn.Body.List) > 0 {
				return fn.Body.List[0]
			}
		}
	}
	return nil
}

func firstGenDecl(f *ast.File) *ast.GenDecl {
	for _, d := range f.Decls {
		if gd, ok := d.(*ast.GenDecl); ok {
			return gd
		}
	}
	return nil
}

func TestSPMDParser(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		spmd  bool
		check func(*testing.T, *ast.File)
	}{
		{
			name: "basic go for range",
			src:  "package p\nvar x int\nfunc f() {\n\tgo for i := range 16 { x++ }\n}",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				rs, ok := stmt.(*ast.RangeStmt)
				if !ok {
					t.Fatalf("expected *ast.RangeStmt, got %T", stmt)
				}
				if !rs.IsSpmd {
					t.Error("expected IsSpmd = true")
				}
				if rs.Key == nil {
					t.Error("expected Key != nil")
				} else if ident, ok := rs.Key.(*ast.Ident); !ok || ident.Name != "i" {
					t.Errorf("expected Key = i, got %v", rs.Key)
				}
				if rs.Tok != token.DEFINE {
					t.Errorf("expected Tok = DEFINE, got %v", rs.Tok)
				}
				if rs.Constraint != nil {
					t.Errorf("expected Constraint = nil, got %v", rs.Constraint)
				}
			},
		},
		{
			name: "bare range",
			src:  "package p\nvar x int\nfunc f() {\n\tgo for range 16 { x++ }\n}",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				rs, ok := stmt.(*ast.RangeStmt)
				if !ok {
					t.Fatalf("expected *ast.RangeStmt, got %T", stmt)
				}
				if !rs.IsSpmd {
					t.Error("expected IsSpmd = true")
				}
				if rs.Key != nil {
					t.Errorf("expected Key = nil, got %v", rs.Key)
				}
				if rs.Constraint != nil {
					t.Errorf("expected Constraint = nil, got %v", rs.Constraint)
				}
			},
		},
		{
			name: "constrained range",
			src:  "package p\nvar x int\nfunc f() {\n\tgo for i := range[4] 16 { x++ }\n}",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				rs, ok := stmt.(*ast.RangeStmt)
				if !ok {
					t.Fatalf("expected *ast.RangeStmt, got %T", stmt)
				}
				if !rs.IsSpmd {
					t.Error("expected IsSpmd = true")
				}
				if rs.Key == nil {
					t.Error("expected Key != nil")
				} else if ident, ok := rs.Key.(*ast.Ident); !ok || ident.Name != "i" {
					t.Errorf("expected Key = i, got %v", rs.Key)
				}
				if rs.Constraint == nil {
					t.Fatal("expected Constraint != nil")
				}
				if lit, ok := rs.Constraint.(*ast.BasicLit); !ok || lit.Value != "4" {
					t.Errorf("expected Constraint = 4, got %v", rs.Constraint)
				}
			},
		},
		{
			name: "key-value range",
			src:  "package p\nvar s []int\nfunc f() {\n\tgo for i, v := range s { _ = v }\n}",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				rs, ok := stmt.(*ast.RangeStmt)
				if !ok {
					t.Fatalf("expected *ast.RangeStmt, got %T", stmt)
				}
				if !rs.IsSpmd {
					t.Error("expected IsSpmd = true")
				}
				if rs.Key == nil {
					t.Error("expected Key != nil")
				} else if ident, ok := rs.Key.(*ast.Ident); !ok || ident.Name != "i" {
					t.Errorf("expected Key = i, got %v", rs.Key)
				}
				if rs.Value == nil {
					t.Error("expected Value != nil")
				} else if ident, ok := rs.Value.(*ast.Ident); !ok || ident.Name != "v" {
					t.Errorf("expected Value = v, got %v", rs.Value)
				}
			},
		},
		{
			name: "bare constrained range",
			src:  "package p\nvar x int\nfunc f() {\n\tgo for range[4] 16 { x++ }\n}",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				rs, ok := stmt.(*ast.RangeStmt)
				if !ok {
					t.Fatalf("expected *ast.RangeStmt, got %T", stmt)
				}
				if !rs.IsSpmd {
					t.Error("expected IsSpmd = true")
				}
				if rs.Key != nil {
					t.Errorf("expected Key = nil, got %v", rs.Key)
				}
				if rs.Constraint == nil {
					t.Fatal("expected Constraint != nil")
				}
				if lit, ok := rs.Constraint.(*ast.BasicLit); !ok || lit.Value != "4" {
					t.Errorf("expected Constraint = 4, got %v", rs.Constraint)
				}
			},
		},
		{
			name: "universal constraint (empty brackets)",
			src:  "package p\nvar x int\nfunc f() {\n\tgo for range[] 16 { x++ }\n}",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				rs, ok := stmt.(*ast.RangeStmt)
				if !ok {
					t.Fatalf("expected *ast.RangeStmt, got %T", stmt)
				}
				if !rs.IsSpmd {
					t.Error("expected IsSpmd = true")
				}
				if rs.Constraint != nil {
					t.Errorf("expected Constraint = nil (universal), got %v", rs.Constraint)
				}
			},
		},
		{
			name: "regular for-range (not SPMD)",
			src:  "package p\nvar x int\nfunc f() {\n\tfor i := range 16 { x++ }\n}",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				rs, ok := stmt.(*ast.RangeStmt)
				if !ok {
					t.Fatalf("expected *ast.RangeStmt, got %T", stmt)
				}
				if rs.IsSpmd {
					t.Error("expected IsSpmd = false for regular for-range")
				}
			},
		},
		{
			name: "go goroutine (no regression)",
			src:  "package p\nfunc f() {\n\tgo func() {}()\n}",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				_, ok := stmt.(*ast.GoStmt)
				if !ok {
					t.Fatalf("expected *ast.GoStmt, got %T", stmt)
				}
			},
		},
		{
			name: "go goroutine without SPMD experiment",
			src:  "package p\nfunc f() {\n\tgo func() {}()\n}",
			spmd: false,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				_, ok := stmt.(*ast.GoStmt)
				if !ok {
					t.Fatalf("expected *ast.GoStmt, got %T", stmt)
				}
			},
		},
		{
			name: "assign variant (= instead of :=)",
			src:  "package p\nvar i int\nvar x int\nfunc f() {\n\tgo for i = range 16 { x++ }\n}",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				rs, ok := stmt.(*ast.RangeStmt)
				if !ok {
					t.Fatalf("expected *ast.RangeStmt, got %T", stmt)
				}
				if !rs.IsSpmd {
					t.Error("expected IsSpmd = true")
				}
				if rs.Tok != token.ASSIGN {
					t.Errorf("expected Tok = ASSIGN, got %v", rs.Tok)
				}
			},
		},
		{
			name: "constrained with assign",
			src:  "package p\nvar i int\nvar x int\nfunc f() {\n\tgo for i = range[8] 16 { x++ }\n}",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				rs, ok := stmt.(*ast.RangeStmt)
				if !ok {
					t.Fatalf("expected *ast.RangeStmt, got %T", stmt)
				}
				if !rs.IsSpmd {
					t.Error("expected IsSpmd = true")
				}
				if rs.Tok != token.ASSIGN {
					t.Errorf("expected Tok = ASSIGN, got %v", rs.Tok)
				}
				if rs.Constraint == nil {
					t.Fatal("expected Constraint != nil")
				}
				if lit, ok := rs.Constraint.(*ast.BasicLit); !ok || lit.Value != "8" {
					t.Errorf("expected Constraint = 8, got %v", rs.Constraint)
				}
			},
		},
		{
			name: "constrained type var decl",
			src:  "package p\nimport \"lanes\"\nvar x lanes.Varying[uint32, 2]",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				gd := firstGenDecl(f)
				if gd == nil {
					t.Fatal("expected GenDecl")
				}
				// Find the var decl (skip import)
				var vd *ast.GenDecl
				for _, d := range f.Decls {
					if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.VAR {
						vd = g
						break
					}
				}
				if vd == nil {
					t.Fatal("expected var GenDecl")
				}
				vs := vd.Specs[0].(*ast.ValueSpec)
				ile, ok := vs.Type.(*ast.IndexListExpr)
				if !ok {
					t.Fatalf("expected *ast.IndexListExpr, got %T", vs.Type)
				}
				if len(ile.Indices) != 2 {
					t.Fatalf("expected 2 indices, got %d", len(ile.Indices))
				}
				// First index should be type ident "uint32"
				if id, ok := ile.Indices[0].(*ast.Ident); !ok || id.Name != "uint32" {
					t.Errorf("expected first index = uint32, got %v", ile.Indices[0])
				}
				// Second index should be integer literal "2"
				if lit, ok := ile.Indices[1].(*ast.BasicLit); !ok || lit.Value != "2" {
					t.Errorf("expected second index = 2, got %v", ile.Indices[1])
				}
			},
		},
		{
			name: "constrained type int4",
			src:  "package p\nimport \"lanes\"\nvar x lanes.Varying[int, 4]",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				var vd *ast.GenDecl
				for _, d := range f.Decls {
					if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.VAR {
						vd = g
						break
					}
				}
				if vd == nil {
					t.Fatal("expected var GenDecl")
				}
				vs := vd.Specs[0].(*ast.ValueSpec)
				ile, ok := vs.Type.(*ast.IndexListExpr)
				if !ok {
					t.Fatalf("expected *ast.IndexListExpr, got %T", vs.Type)
				}
				if len(ile.Indices) != 2 {
					t.Fatalf("expected 2 indices, got %d", len(ile.Indices))
				}
				if lit, ok := ile.Indices[1].(*ast.BasicLit); !ok || lit.Value != "4" {
					t.Errorf("expected second index = 4, got %v", ile.Indices[1])
				}
			},
		},
		{
			name: "constrained func param",
			src:  "package p\nimport \"lanes\"\nfunc f(x lanes.Varying[float32, 8]) {}",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				var fn *ast.FuncDecl
				for _, d := range f.Decls {
					if fd, ok := d.(*ast.FuncDecl); ok {
						fn = fd
						break
					}
				}
				if fn == nil {
					t.Fatal("expected FuncDecl")
				}
				params := fn.Type.Params.List
				if len(params) != 1 {
					t.Fatalf("expected 1 param, got %d", len(params))
				}
				ile, ok := params[0].Type.(*ast.IndexListExpr)
				if !ok {
					t.Fatalf("expected *ast.IndexListExpr, got %T", params[0].Type)
				}
				if len(ile.Indices) != 2 {
					t.Fatalf("expected 2 indices, got %d", len(ile.Indices))
				}
				if lit, ok := ile.Indices[1].(*ast.BasicLit); !ok || lit.Value != "8" {
					t.Errorf("expected second index = 8, got %v", ile.Indices[1])
				}
			},
		},
		{
			name: "constrained return type",
			src:  "package p\nimport \"lanes\"\nvar x lanes.Varying[int32, 4]\nfunc f() lanes.Varying[int32, 4] { return x }",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				var fn *ast.FuncDecl
				for _, d := range f.Decls {
					if fd, ok := d.(*ast.FuncDecl); ok {
						fn = fd
						break
					}
				}
				if fn == nil {
					t.Fatal("expected FuncDecl")
				}
				results := fn.Type.Results.List
				if len(results) != 1 {
					t.Fatalf("expected 1 result, got %d", len(results))
				}
				ile, ok := results[0].Type.(*ast.IndexListExpr)
				if !ok {
					t.Fatalf("expected *ast.IndexListExpr, got %T", results[0].Type)
				}
				if len(ile.Indices) != 2 {
					t.Fatalf("expected 2 indices, got %d", len(ile.Indices))
				}
				if lit, ok := ile.Indices[1].(*ast.BasicLit); !ok || lit.Value != "4" {
					t.Errorf("expected second index = 4, got %v", ile.Indices[1])
				}
			},
		},
		{
			name: "constrained type alias",
			src:  "package p\nimport \"lanes\"\ntype V = lanes.Varying[uint16, 8]",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				var td *ast.GenDecl
				for _, d := range f.Decls {
					if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.TYPE {
						td = g
						break
					}
				}
				if td == nil {
					t.Fatal("expected type GenDecl")
				}
				ts := td.Specs[0].(*ast.TypeSpec)
				ile, ok := ts.Type.(*ast.IndexListExpr)
				if !ok {
					t.Fatalf("expected *ast.IndexListExpr, got %T", ts.Type)
				}
				if len(ile.Indices) != 2 {
					t.Fatalf("expected 2 indices, got %d", len(ile.Indices))
				}
				if lit, ok := ile.Indices[1].(*ast.BasicLit); !ok || lit.Value != "8" {
					t.Errorf("expected second index = 8, got %v", ile.Indices[1])
				}
			},
		},
		{
			name: "constrained byte16",
			src:  "package p\nimport \"lanes\"\nvar x lanes.Varying[byte, 16]",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				var vd *ast.GenDecl
				for _, d := range f.Decls {
					if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.VAR {
						vd = g
						break
					}
				}
				if vd == nil {
					t.Fatal("expected var GenDecl")
				}
				vs := vd.Specs[0].(*ast.ValueSpec)
				ile, ok := vs.Type.(*ast.IndexListExpr)
				if !ok {
					t.Fatalf("expected *ast.IndexListExpr, got %T", vs.Type)
				}
				if len(ile.Indices) != 2 {
					t.Fatalf("expected 2 indices, got %d", len(ile.Indices))
				}
				if id, ok := ile.Indices[0].(*ast.Ident); !ok || id.Name != "byte" {
					t.Errorf("expected first index = byte, got %v", ile.Indices[0])
				}
				if lit, ok := ile.Indices[1].(*ast.BasicLit); !ok || lit.Value != "16" {
					t.Errorf("expected second index = 16, got %v", ile.Indices[1])
				}
			},
		},
		{
			name: "constrained type conversion",
			src:  "package p\nimport \"lanes\"\nvar x lanes.Varying[uint32, 2]\nfunc f() { _ = lanes.Varying[uint16, 2](x) }",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				as, ok := stmt.(*ast.AssignStmt)
				if !ok {
					t.Fatalf("expected *ast.AssignStmt, got %T", stmt)
				}
				call, ok := as.Rhs[0].(*ast.CallExpr)
				if !ok {
					t.Fatalf("expected *ast.CallExpr, got %T", as.Rhs[0])
				}
				ile, ok := call.Fun.(*ast.IndexListExpr)
				if !ok {
					t.Fatalf("expected *ast.IndexListExpr, got %T", call.Fun)
				}
				if len(ile.Indices) != 2 {
					t.Fatalf("expected 2 indices, got %d", len(ile.Indices))
				}
				if lit, ok := ile.Indices[1].(*ast.BasicLit); !ok || lit.Value != "2" {
					t.Errorf("expected second index = 2, got %v", ile.Indices[1])
				}
			},
		},
		{
			name: "constrained type switch case",
			src:  "package p\nimport \"lanes\"\nfunc f(v any) { switch v.(type) { case lanes.Varying[int, 4]: } }",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				ts, ok := stmt.(*ast.TypeSwitchStmt)
				if !ok {
					t.Fatalf("expected *ast.TypeSwitchStmt, got %T", stmt)
				}
				body := ts.Body.List
				if len(body) == 0 {
					t.Fatal("expected at least one case clause")
				}
				cc, ok := body[0].(*ast.CaseClause)
				if !ok {
					t.Fatalf("expected *ast.CaseClause, got %T", body[0])
				}
				if len(cc.List) != 1 {
					t.Fatalf("expected 1 case type, got %d", len(cc.List))
				}
				ile, ok := cc.List[0].(*ast.IndexListExpr)
				if !ok {
					t.Fatalf("expected *ast.IndexListExpr, got %T", cc.List[0])
				}
				if len(ile.Indices) != 2 {
					t.Fatalf("expected 2 indices, got %d", len(ile.Indices))
				}
				if lit, ok := ile.Indices[1].(*ast.BasicLit); !ok || lit.Value != "4" {
					t.Errorf("expected second index = 4, got %v", ile.Indices[1])
				}
			},
		},
		{
			name: "constrained conversion assign",
			src:  "package p\nimport \"lanes\"\nvar x int\nfunc f() { _ = lanes.Varying[int, 4](x) }",
			spmd: true,
			check: func(t *testing.T, f *ast.File) {
				stmt := firstFuncStmt(f)
				as, ok := stmt.(*ast.AssignStmt)
				if !ok {
					t.Fatalf("expected *ast.AssignStmt, got %T", stmt)
				}
				call, ok := as.Rhs[0].(*ast.CallExpr)
				if !ok {
					t.Fatalf("expected *ast.CallExpr, got %T", as.Rhs[0])
				}
				ile, ok := call.Fun.(*ast.IndexListExpr)
				if !ok {
					t.Fatalf("expected *ast.IndexListExpr, got %T", call.Fun)
				}
				if len(ile.Indices) != 2 {
					t.Fatalf("expected 2 indices, got %d", len(ile.Indices))
				}
				if lit, ok := ile.Indices[1].(*ast.BasicLit); !ok || lit.Value != "4" {
					t.Errorf("expected second index = 4, got %v", ile.Indices[1])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.spmd {
				revert := setGOEXPERIMENT("spmd")
				defer revert()
			} else {
				revert := setGOEXPERIMENT("")
				defer revert()
			}

			fset := token.NewFileSet()
			f, err := ParseFile(fset, "", tt.src, 0)
			if err != nil {
				t.Fatalf("ParseFile failed: %v", err)
			}
			tt.check(t, f)
		})
	}
}
