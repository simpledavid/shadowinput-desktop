package domain

import (
	"strings"
	"testing"
)

func TestWritingDiffReconstructsBothDrafts(t *testing.T) {
	for _, pair := range [][2]string{{"", "new"}, {"old", ""}, {"你好，Alice 👩🏽‍💻\nFriday", "你好，Alice 👩🏽‍💻\nMonday"}, {"预计下周答复", "预计下周给你答复"}, {strings.Repeat("甲", 3000), strings.Repeat("乙", 3000)}} {
		r := ReviewWriting(pair[0], pair[1], nil)
		var before, after strings.Builder
		for _, e := range r.Edits {
			if e.Kind != "add" {
				before.WriteString(e.Text)
			}
			if e.Kind != "remove" {
				after.WriteString(e.Text)
			}
		}
		if before.String() != pair[0] || after.String() != pair[1] {
			t.Fatal("diff lost or invented text")
		}
	}
	if !ReviewWriting(strings.Repeat("甲", 3000), strings.Repeat("乙", 3000), nil).Coarse {
		t.Fatal("unbounded detailed diff")
	}
}

func TestWritingReviewFlagsMeaningfulChanges(t *testing.T) {
	r := ReviewWriting("预计周五给 Alice 答复，费用 1200 元。", "保证周一给 Alicia 答复，费用 2000 元。", []string{"Alice"})
	kinds := map[string]bool{}
	for _, issue := range r.Issues {
		kinds[issue.Kind] = true
	}
	for _, kind := range []string{"numbers", "dates", "terms", "commitments"} {
		if !kinds[kind] {
			t.Errorf("missing %s", kind)
		}
	}
	if len(ReviewWriting("预计周五给 Alice 答复，费用 1200 元。", "费用 1200 元，预计周五给 Alice 答复。", []string{"Alice"}).Issues) != 0 {
		t.Fatal("unchanged facts were flagged")
	}
	if len(ReviewWriting("We may deliver Friday.", "We guarantee delivery Friday.", nil).Issues) == 0 {
		t.Fatal("lost hedge was missed")
	}
	if len(ReviewWriting("We cannot deliver.", "We can deliver.", nil).Issues) == 0 {
		t.Fatal("lost negation was missed")
	}
}

func TestWritingProfilesValidateReferencesAndLimitSharedTerms(t *testing.T) {
	p := DefaultWritingProfile()
	if p.Validate() != nil {
		t.Fatal("invalid defaults")
	}
	p.Defaults["polish"] = "missing"
	if p.Validate() == nil {
		t.Fatal("dangling default accepted")
	}
	p = DefaultWritingProfile()
	p.Terms = []string{"Alice", "Alice"}
	if p.Validate() == nil {
		t.Fatal("duplicate terms accepted")
	}
	s := DefaultWritingProfile().Styles[0]
	r := WritingRequest{Task: "polish", Text: "Hello Alice", Model: "test", Style: &s, Terms: []string{"Alice", "Private Project"}}
	if r.Validate() != nil {
		t.Fatal("valid request rejected")
	}
	prompt := WritingInstructions(r)
	if !strings.Contains(prompt, "Alice") || strings.Contains(prompt, "Private Project") {
		t.Fatal("unrelated glossary entry shared")
	}
}
