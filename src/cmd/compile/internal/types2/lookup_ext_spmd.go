// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import "internal/buildcfg"

// spmdUnwrapVaryingPointer returns the pointed-to type when t is Varying[*S],
// i.e., an SPMDType wrapping a Pointer. Returns (nil, false) otherwise.
//
// Used by the selector path to extend field-or-method lookup to
// varying-pointer receivers: for lookup purposes, Varying[*S].field behaves
// like (*S).field, while the returned value is later wrapped back into
// Varying[fieldT] by spmdWrapFieldType.
func spmdUnwrapVaryingPointer(t Type) (Type, bool) {
	if !buildcfg.Experiment.SPMD {
		return nil, false
	}
	sv, ok := t.(*SPMDType)
	if !ok {
		return nil, false
	}
	ptr, ok := sv.Elem().(*Pointer)
	if !ok {
		return nil, false
	}
	return ptr.Elem(), true
}
