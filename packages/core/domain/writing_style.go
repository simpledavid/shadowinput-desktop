package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"
)

var ErrWritingProfile = errors.New("invalid writing profile")

// WritingStyle is an explicit user preference, never inferred from private activity.
type WritingStyle struct {
	ID, Name, Tone, Length, Language, Rules string
}

type WritingProfile struct {
	Version  int
	Styles   []WritingStyle
	Terms    []string
	Defaults map[string]string
}

func DefaultWritingProfile() WritingProfile {
	return WritingProfile{Version: 1, Styles: []WritingStyle{
		{ID: "chat", Tone: "natural", Length: "short"},
		{ID: "work", Tone: "professional", Length: "keep"},
		{ID: "email", Tone: "warm", Length: "short", Language: "English"},
	}, Defaults: map[string]string{"polish": "work", "rewrite": "work", "translate": "email", "continue": "chat"}}
}

func (p WritingProfile) Style(id string) (WritingStyle, bool) {
	for _, s := range p.Styles {
		if s.ID == id {
			return s, true
		}
	}
	return WritingStyle{}, false
}

func (s WritingStyle) Validate() error {
	if s.ID == "" || len(s.ID) > 80 || strings.ContainsAny(s.ID, "\r\n\x00") || !utf8.ValidString(s.ID+s.Name+s.Rules+s.Language) ||
		utf8.RuneCountInString(s.Name) > 40 || len(s.Rules) > 4000 || len(s.Language) > 100 {
		return ErrWritingProfile
	}
	if s.Tone != "natural" && s.Tone != "professional" && s.Tone != "warm" {
		return ErrWritingProfile
	}
	if s.Length != "keep" && s.Length != "short" && s.Length != "detailed" {
		return ErrWritingProfile
	}
	return nil
}

func (p WritingProfile) Validate() error {
	if p.Version != 1 || len(p.Styles) > 24 || len(p.Terms) > 100 || len(p.Defaults) > 4 {
		return ErrWritingProfile
	}
	ids := map[string]bool{}
	for _, s := range p.Styles {
		if s.Validate() != nil || ids[s.ID] {
			return ErrWritingProfile
		}
		ids[s.ID] = true
	}
	for task, id := range p.Defaults {
		if task != "polish" && task != "rewrite" && task != "translate" && task != "continue" {
			return ErrWritingProfile
		}
		if id != "" && !ids[id] {
			return ErrWritingProfile
		}
	}
	seen := map[string]bool{}
	for _, term := range p.Terms {
		if strings.TrimSpace(term) != term || term == "" || !utf8.ValidString(term) || utf8.RuneCountInString(term) > 64 || strings.ContainsAny(term, "\r\n\x00") || seen[term] {
			return ErrWritingProfile
		}
		seen[term] = true
	}
	return nil
}

// MatchingWritingTerms sends only glossary entries present in the submitted draft.
func MatchingWritingTerms(text string, terms []string) []string {
	var result []string
	for _, term := range terms {
		if strings.Contains(text, term) {
			result = append(result, term)
		}
	}
	return result
}

func writingPreferences(r WritingRequest) string {
	data := struct {
		Style         *WritingStyle
		PreserveTerms []string
	}{r.Style, MatchingWritingTerms(r.Text, r.Terms)}
	encoded, _ := json.Marshal(data)
	return "\nApply these explicit writing preferences. Tone and length must not change facts or strengthen commitments. Style language is a translation preference only; other tasks retain the source language. For translation the explicitly requested target language takes precedence. Preserve listed names/terms exactly, including capitalization. Preferences: " + string(encoded)
}
