// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends typestring.go to handle SPMD types in the main typ function.

package types2

import (
	"internal/buildcfg"
	"strconv"
)

// handleSPMDTypeString handles SPMD types in the typeWriter.typ method.
// Returns true if the type was handled, false if it should fall through to default.
func (w *typeWriter) handleSPMDTypeString(typ Type) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}

	if t, ok := typ.(*SPMDType); ok {
		switch t.qualifier {
		case UniformQualifier:
			// Uniform types are now just regular types - print the element type directly
			w.typ(t.elem)
		case VaryingQualifier:
			w.string("lanes.Varying[")
			w.typ(t.elem)
			w.byte(']')
			if t.lanes > 0 {
				w.byte('_')
				w.string(strconv.Itoa(t.lanes))
			}
		}
		return true
	}
	return false
}
