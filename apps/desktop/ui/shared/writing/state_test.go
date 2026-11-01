package writing

import (
	"arion.local/core/domain"
	"testing"
)

func TestReviewGatesDeliveryAndCannotSurviveInvalidation(t *testing.T) {
	s := NewState()
	s.Source, s.Output, s.Completed = "Maybe 12", "Definitely 20", true
	s.Review = domain.ReviewWriting(s.Source, s.Output, nil)
	if len(s.Review.Issues) == 0 || s.CanDeliver() {
		t.Fatal("unreviewed changes can be delivered")
	}
	s.Reviewed = true
	if !s.CanDeliver() {
		t.Fatal("reviewed result cannot be delivered")
	}
	s.Busy = true
	if s.CanDeliver() {
		t.Fatal("busy result can be delivered")
	}
	s.Busy = false
	s.InvalidateResult()
	if s.CanDeliver() || s.Reviewed || s.Output != "" || len(s.Review.Issues) != 0 || s.Source != "Maybe 12" {
		t.Fatal("invalidation lost the draft or retained stale approval")
	}
}
