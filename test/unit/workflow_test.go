package unit_test

import (
	"testing"

	"github.com/yourorg/boardgov/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowTransitions_ValidPaths(t *testing.T) {
	cases := []struct {
		from domain.DocumentStatus
		to   domain.DocumentStatus
		ok   bool
	}{
		{domain.DocDraft, domain.DocSubmitted, true},
		{domain.DocSubmitted, domain.DocUnderReview, true},
		{domain.DocSubmitted, domain.DocRejected, true},
		{domain.DocUnderReview, domain.DocApproved, true},
		{domain.DocUnderReview, domain.DocRejected, true},
		{domain.DocApproved, domain.DocArchived, true},
		{domain.DocRejected, domain.DocDraft, true},
		{domain.DocDraft, domain.DocApproved, false},
		{domain.DocDraft, domain.DocUnderReview, false},
		{domain.DocApproved, domain.DocDraft, false},
		{domain.DocArchived, domain.DocDraft, false},
	}

	for _, tc := range cases {
		t.Run(string(tc.from)+"->"+string(tc.to), func(t *testing.T) {
			result := domain.CanTransition(tc.from, tc.to)
			assert.Equal(t, tc.ok, result)
		})
	}
}

func TestValidateTransition_Error(t *testing.T) {
	err := domain.ValidateTransition(domain.DocDraft, domain.DocApproved)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidTransition)
}

func TestValidateTransition_Success(t *testing.T) {
	err := domain.ValidateTransition(domain.DocDraft, domain.DocSubmitted)
	require.NoError(t, err)
}
