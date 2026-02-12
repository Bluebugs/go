// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import . "internal/types/errors"

func (x *operand) handleSPMDAssignability(V, T Type, cause *string) (handled, assignable bool, code Code) {
	return false, false, 0
}

func (x *operand) convertibleToSPMD(check *Checker, T Type, cause *string) bool {
	return false
}
