//go:build !goexperiment.spmd

// Package lanes provides cross-lane operations for SPMD programming.
// This file provides stubs when SPMD experiment is disabled.
package lanes

// When SPMD is disabled, all functions are unavailable and will cause compile errors.
// This prevents accidental usage of SPMD code without the experiment flag.

// Note: These functions are intentionally not implemented to cause compile-time errors
// when SPMD code is used without GOEXPERIMENT=spmd.