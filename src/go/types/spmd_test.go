// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

import (
	"go/types"
	"testing"
)

// TestSPMDTypeLanesField verifies the new Lanes() accessor returns the
// width-fixed value when set, 0 for abstract types.
func TestSPMDTypeLanesField(t *testing.T) {
	abstract := types.NewVarying(types.Typ[types.Int])
	if abstract.Lanes() != 0 {
		t.Fatalf("abstract Varying[int]: Lanes() = %d, want 0", abstract.Lanes())
	}
	fixed := types.NewVaryingWithLanes(types.Typ[types.Int], 4)
	if fixed.Lanes() != 4 {
		t.Fatalf("fixed Varying[int]_4: Lanes() = %d, want 4", fixed.Lanes())
	}
}

// TestSPMDTypeIdenticalLanes verifies type identity respects the lane count.
// Two width-fixed types with the same lane count are identical; with
// different lane counts (or one abstract / one fixed) they are not.
func TestSPMDTypeIdenticalLanes(t *testing.T) {
	a := types.NewVaryingWithLanes(types.Typ[types.Int], 4)
	b := types.NewVaryingWithLanes(types.Typ[types.Int], 4)
	c := types.NewVaryingWithLanes(types.Typ[types.Int], 8)
	abstract := types.NewVarying(types.Typ[types.Int])

	if !types.Identical(a, b) {
		t.Error("Identical(Varying[int]_4, Varying[int]_4) = false; want true")
	}
	if types.Identical(a, c) {
		t.Error("Identical(Varying[int]_4, Varying[int]_8) = true; want false")
	}
	if types.Identical(a, abstract) {
		t.Error("Identical(Varying[int]_4, Varying[int]_0) = true; want false (strict)")
	}
}

// TestSPMDTypeStringLanes verifies the printed format includes the width.
func TestSPMDTypeStringLanes(t *testing.T) {
	abstract := types.NewVarying(types.Typ[types.Int])
	if got := abstract.String(); got != "lanes.Varying[int]" {
		t.Errorf("abstract String() = %q; want %q", got, "lanes.Varying[int]")
	}
	fixed := types.NewVaryingWithLanes(types.Typ[types.Int], 4)
	if got := fixed.String(); got != "lanes.Varying[int]_4" {
		t.Errorf("fixed String() = %q; want %q", got, "lanes.Varying[int]_4")
	}
}
