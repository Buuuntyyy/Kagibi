// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package authprovider

import (
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const testTOTPSecret = "JBSWY3DPEHPK3PXP"

func codeForStep(t *testing.T, step int64) string {
	t.Helper()
	opts := totp.ValidateOpts{Period: 30, Skew: 0, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1}
	code, err := totp.GenerateCodeCustom(testTOTPSecret, time.Unix(step*30, 0), opts)
	if err != nil {
		t.Fatalf("GenerateCodeCustom: %v", err)
	}
	return code
}

// The current, previous and next time-steps must all validate (±1 period window),
// and matchTOTPStep must report the exact step that matched.
func TestMatchTOTPStep_WindowAndStep(t *testing.T) {
	nowStep := time.Now().Unix() / 30
	for _, delta := range []int64{-1, 0, 1} {
		want := nowStep + delta
		got, valid := matchTOTPStep(codeForStep(t, want), testTOTPSecret)
		if !valid {
			t.Fatalf("code for step %+d should validate", delta)
		}
		// Allow a one-step drift if the clock ticked over a 30s boundary mid-test.
		if got != want && got != want+1 && got != want-1 {
			t.Fatalf("delta %+d: got step %d, want ~%d", delta, got, want)
		}
	}
}

// A code from outside the ±1 window must not validate.
func TestMatchTOTPStep_OutsideWindow(t *testing.T) {
	nowStep := time.Now().Unix() / 30
	if _, valid := matchTOTPStep(codeForStep(t, nowStep+5), testTOTPSecret); valid {
		t.Fatal("a code 5 steps in the future must not validate in the current window")
	}
}

// Replay semantics: once a step is recorded as last-used, an earlier-or-equal step's
// code must be treated as a replay. This mirrors the check in ValidateTOTPCode.
func TestMatchTOTPStep_ReplayIsOlderStep(t *testing.T) {
	nowStep := time.Now().Unix() / 30

	curStep, ok := matchTOTPStep(codeForStep(t, nowStep), testTOTPSecret)
	if !ok {
		t.Fatal("current code should validate")
	}
	prevStep, ok := matchTOTPStep(codeForStep(t, nowStep-1), testTOTPSecret)
	if !ok {
		t.Fatal("previous code should validate")
	}
	// After consuming the current step, the previous window's code is a replay.
	if !(prevStep <= curStep) {
		t.Fatalf("previous step %d should be <= current step %d (replay)", prevStep, curStep)
	}
}
