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
