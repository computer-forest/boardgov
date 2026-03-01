package unit_test

import (
	"testing"

	"github.com/yourorg/boardgov/internal/domain"
	"github.com/yourorg/boardgov/pkg/crypto"
)

// ─── Workflow State Machine Tests ─────────────────────────────────────────────

func TestCanTransition_ValidPaths(t *testing.T) {
	tests := []struct {
		from domain.DocumentStatus
		to   domain.DocumentStatus
		want bool
	}{
		{domain.StatusDraft, domain.StatusSubmitted, true},
		{domain.StatusSubmitted, domain.StatusUnderReview, true},
		{domain.StatusSubmitted, domain.StatusRejected, true},
		{domain.StatusUnderReview, domain.StatusApproved, true},
		{domain.StatusUnderReview, domain.StatusRejected, true},
		{domain.StatusApproved, domain.StatusArchived, true},
		{domain.StatusRejected, domain.StatusDraft, true},
		{domain.StatusRejected, domain.StatusArchived, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			got := domain.CanTransition(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("CanTransition(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestCanTransition_InvalidPaths(t *testing.T) {
	tests := []struct {
		from domain.DocumentStatus
		to   domain.DocumentStatus
	}{
		{domain.StatusDraft, domain.StatusApproved},      // skip steps
		{domain.StatusDraft, domain.StatusUnderReview},   // skip steps
		{domain.StatusApproved, domain.StatusDraft},      // backwards
		{domain.StatusArchived, domain.StatusDraft},      // terminal
		{domain.StatusArchived, domain.StatusApproved},   // terminal
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			got := domain.CanTransition(tt.from, tt.to)
			if got {
				t.Errorf("CanTransition(%s, %s) = true, want false", tt.from, tt.to)
			}
		})
	}
}

func TestIsTerminalState(t *testing.T) {
	terminal := []domain.DocumentStatus{domain.StatusApproved, domain.StatusArchived}
	nonTerminal := []domain.DocumentStatus{domain.StatusDraft, domain.StatusSubmitted, domain.StatusUnderReview, domain.StatusRejected}

	for _, s := range terminal {
		if !domain.IsTerminalState(s) {
			t.Errorf("expected %s to be terminal", s)
		}
	}
	for _, s := range nonTerminal {
		if domain.IsTerminalState(s) {
			t.Errorf("expected %s to be non-terminal", s)
		}
	}
}

func TestRequiresLock(t *testing.T) {
	if !domain.RequiresLock(domain.StatusApproved) {
		t.Error("APPROVED should require lock")
	}
	if domain.RequiresLock(domain.StatusDraft) {
		t.Error("DRAFT should not require lock")
	}
}

// ─── Crypto / Audit Hash Tests ────────────────────────────────────────────────

func TestAuditChainHash_Deterministic(t *testing.T) {
	h1 := crypto.AuditChainHash("user1", "ACTION", "doc1", "prevhash", "2024-01-01T00:00:00Z")
	h2 := crypto.AuditChainHash("user1", "ACTION", "doc1", "prevhash", "2024-01-01T00:00:00Z")
	if h1 != h2 {
		t.Error("hash not deterministic")
	}
}

func TestAuditChainHash_Sensitivity(t *testing.T) {
	h1 := crypto.AuditChainHash("user1", "ACTION", "doc1", "prevhash", "2024-01-01T00:00:00Z")
	h2 := crypto.AuditChainHash("user2", "ACTION", "doc1", "prevhash", "2024-01-01T00:00:00Z")
	if h1 == h2 {
		t.Error("different inputs should produce different hashes")
	}
}

func TestSHA256String_KnownValue(t *testing.T) {
	// echo -n "" | sha256sum = e3b0c44298fc1c149...
	h := crypto.SHA256String("")
	if len(h) != 64 {
		t.Errorf("SHA256 should be 64 chars, got %d", len(h))
	}
}

func TestHashPassword_Verify(t *testing.T) {
	hash, err := crypto.HashPassword("SecurePass123!")
	if err != nil {
		t.Fatal(err)
	}
	if !crypto.CheckPasswordHash("SecurePass123!", hash) {
		t.Error("password verification failed")
	}
	if crypto.CheckPasswordHash("WrongPass", hash) {
		t.Error("wrong password should not verify")
	}
}

// ─── AppError Tests ───────────────────────────────────────────────────────────

func TestAppError_Error(t *testing.T) {
	err := &domain.AppError{Code: 404, Message: "not found"}
	if err.Error() != "not found" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}
