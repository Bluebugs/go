// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "internal/buildcfg"

func (u *unifier) handleSPMDUnification(x, y Type, mode unifyMode) (handled bool, unified bool) {
	if !buildcfg.Experiment.SPMD {
		return false, false
	}

	xSPMD, xIsSPMD := x.(*SPMDType)
	ySPMD, yIsSPMD := y.(*SPMDType)

	if !xIsSPMD && !yIsSPMD {
		return false, false
	}

	if xIsSPMD && yIsSPMD {
		if xSPMD.qualifier != ySPMD.qualifier {
			return true, false
		}
		if xSPMD.qualifier == VaryingQualifier && xSPMD.constraint != ySPMD.constraint {
			// Two different specific constraints (e.g. 4 vs 8) are incompatible.
			if xSPMD.constraint > 0 && ySPMD.constraint > 0 {
				return true, false
			}
			// Allow when one side is unconstrained (-1) or universal (0).
			// This enables generic builtins to accept constrained arguments,
			// e.g. FromConstrained[T](data Varying[T]) accepting Varying[T, 8].
		}
		return true, u.nify(xSPMD.elem, ySPMD.elem, mode, nil)
	}

	var spmdType *SPMDType
	var regularType Type
	var spmdIsX bool

	if xIsSPMD {
		spmdType = xSPMD
		regularType = y
		spmdIsX = true
	} else {
		spmdType = ySPMD
		regularType = x
		spmdIsX = false
	}

	elemType := spmdType.elem

	if spmdIsX {
		return true, u.nify(elemType, regularType, mode, nil)
	}
	return true, u.nify(regularType, elemType, mode, nil)
}
