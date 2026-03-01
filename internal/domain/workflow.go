package domain

import "fmt"

// State machine transitions
var validTransitions = map[DocumentStatus][]DocumentStatus{
	DocDraft:       {DocSubmitted},
	DocSubmitted:   {DocUnderReview, DocRejected},
	DocUnderReview: {DocApproved, DocRejected},
	DocApproved:    {DocArchived},
	DocRejected:    {DocDraft, DocArchived},
	DocArchived:    {},
}

func CanTransition(from, to DocumentStatus) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func ValidateTransition(from, to DocumentStatus) error {
	if !CanTransition(from, to) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, from, to)
	}
	return nil
}
