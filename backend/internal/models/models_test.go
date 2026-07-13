package models

import "testing"

func TestValidStatus(t *testing.T) {
	for _, s := range []string{StatusDraft, StatusApproved, StatusPaid} {
		if !ValidStatus(s) {
			t.Errorf("expected %q to be a valid status", s)
		}
	}
	for _, s := range []string{"", "void", "pending"} {
		if ValidStatus(s) {
			t.Errorf("expected %q to be an invalid status", s)
		}
	}
}

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		// Happy path forward.
		{StatusDraft, StatusApproved, true},
		{StatusApproved, StatusPaid, true},
		// Reopen an approved payout.
		{StatusApproved, StatusDraft, true},
		// Illegal skips.
		{StatusDraft, StatusPaid, false},
		// Paid is terminal.
		{StatusPaid, StatusApproved, false},
		{StatusPaid, StatusDraft, false},
		{StatusPaid, StatusPaid, false},
		// No-op / backward moves that aren't allowed.
		{StatusDraft, StatusDraft, false},
		{StatusApproved, StatusApproved, false},
		// Unknown statuses.
		{StatusDraft, "void", false},
		{"void", StatusApproved, false},
	}
	for _, c := range cases {
		if got := CanTransition(c.from, c.to); got != c.want {
			t.Errorf("CanTransition(%q, %q) = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}
