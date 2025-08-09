//go:build goexperiment.spmd

// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file extends typestring.go to handle SPMD types in the main typ function.

package types2

import "strconv"

// handleSPMDTypeString handles SPMD types in the typeWriter.typ method.
// Returns true if the type was handled, false if it should fall through to default.
func (w *typeWriter) handleSPMDTypeString(typ Type) bool {
	if t, ok := typ.(*SPMDType); ok {
		switch t.qualifier {
		case UniformQualifier:
			w.string("uniform ")
		case VaryingQualifier:
			w.string("varying")
			if t.IsConstrained() {
				w.byte('[')
				if t.IsUniversalConstrained() {
					// varying[] - universal constraint
				} else {
					// varying[n] - numeric constraint
					w.string(strconv.FormatInt(t.constraint, 10))
				}
				w.byte(']')
			}
			w.byte(' ')
		}
		w.typ(t.elem)
		return true
	}
	return false
}