// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "internal/buildcfg"

func (w *typeWriter) handleSPMDTypeString(typ Type) bool {
	if !buildcfg.Experiment.SPMD {
		return false
	}

	if t, ok := typ.(*SPMDType); ok {
		switch t.qualifier {
		case UniformQualifier:
			w.typ(t.elem)
		case VaryingQualifier:
			w.string("lanes.Varying[")
			w.typ(t.elem)
			w.byte(']')
		}
		return true
	}
	return false
}
