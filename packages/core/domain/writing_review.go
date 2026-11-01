package domain

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
)

type TextEdit struct{ Kind, Text string }
type WritingIssue struct{ Kind, Before, After string }
type WritingReview struct {
	Edits  []TextEdit
	Issues []WritingIssue
	Coarse bool
}

// ReviewWriting is local, advisory comparison, not a semantic correctness verdict.
func ReviewWriting(source, result string, terms []string) WritingReview {
	r := WritingReview{}
	r.Edits, r.Coarse = writingDiff(source, result)
	compare := func(kind string, a, b []string) {
		slices.Sort(a)
		slices.Sort(b)
		if !slices.Equal(a, b) {
			r.Issues = append(r.Issues, WritingIssue{kind, reviewExcerpt(strings.Join(a, ", ")), reviewExcerpt(strings.Join(b, ", "))})
		}
	}
	compare("numbers", writingNumbers.FindAllString(source, -1), writingNumbers.FindAllString(result, -1))
	compare("dates", writingDates.FindAllString(source, -1), writingDates.FindAllString(result, -1))
	for _, term := range terms {
		if strings.Count(source, term) > 0 && strings.Count(source, term) != strings.Count(result, term) {
			r.Issues = append(r.Issues, WritingIssue{"terms", term, ""})
		}
	}
	if (uncertain(source) || writingHedges.MatchString(source)) && !(uncertain(result) || writingHedges.MatchString(result)) ||
		!writingPromises.MatchString(source) && writingPromises.MatchString(result) || negations(result) < negations(source) {
		r.Issues = append(r.Issues, WritingIssue{Kind: "commitments"})
	}
	return r
}

var writingNumbers = regexp.MustCompile(`[0-9]+(?:[.,][0-9]+)*(?:%|％)?`)
var writingDates = regexp.MustCompile(`(?i)(?:[0-9]{4}[-/年][0-9]{1,2}[-/月][0-9]{1,2}日?|[0-9]{1,2}月[0-9]{1,2}[日号]?|(?:这|下|本|上)?(?:周|星期)[一二三四五六日天]|今天|明天|后天|下周|下个月|\b(?:monday|tuesday|wednesday|thursday|friday|saturday|sunday|today|tomorrow|next week|next month)\b)`)
var writingHedges = regexp.MustCompile(`(?i)预计|争取|尽量|计划|打算|\b(?:expect|estimate|aim|intend|plan)\b`)
var writingPromises = regexp.MustCompile(`(?i)保证|承诺|一定会|肯定会|\b(?:guarantee|promise|definitely|certainly)\b`)

func reviewExcerpt(text string) string {
	r := []rune(text)
	if len(r) > 100 {
		return string(r[:100]) + "…"
	}
	return text
}

// Keep Latin words together and CJK characters individually comparable.
func writingTokens(text string) []string {
	var result []string
	var word strings.Builder
	flush := func() {
		if word.Len() > 0 {
			result = append(result, word.String())
			word.Reset()
		}
	}
	for _, r := range text {
		if (unicode.IsLetter(r) && !unicode.Is(unicode.Han, r)) || unicode.IsNumber(r) || unicode.IsMark(r) {
			word.WriteRune(r)
		} else {
			flush()
			result = append(result, string(r))
		}
	}
	flush()
	return result
}

func writingDiff(before, after string) ([]TextEdit, bool) {
	a, b := writingTokens(before), writingTokens(after)
	var edits []TextEdit
	var run strings.Builder
	runKind := ""
	flush := func() {
		if run.Len() > 0 {
			edits = append(edits, TextEdit{runKind, run.String()})
			run.Reset()
		}
	}
	appendEdit := func(kind, text string) {
		if text == "" {
			return
		}
		if runKind != kind {
			flush()
			runKind = kind
		}
		run.WriteString(text)
	}
	start := 0
	for start < len(a) && start < len(b) && a[start] == b[start] {
		start++
	}
	appendEdit("equal", strings.Join(a[:start], ""))
	a, b = a[start:], b[start:]
	end := 0
	for end < len(a) && end < len(b) && a[len(a)-end-1] == b[len(b)-end-1] {
		end++
	}
	suffix := strings.Join(a[len(a)-end:], "")
	a, b = a[:len(a)-end], b[:len(b)-end]
	coarse := (len(a)+1)*(len(b)+1) > 1_000_000
	if coarse {
		appendEdit("remove", strings.Join(a, ""))
		appendEdit("add", strings.Join(b, ""))
	} else {
		width := len(b) + 1
		dp := make([]uint32, (len(a)+1)*width)
		for i := len(a) - 1; i >= 0; i-- {
			for j := len(b) - 1; j >= 0; j-- {
				if a[i] == b[j] {
					dp[i*width+j] = dp[(i+1)*width+j+1] + 1
				} else {
					dp[i*width+j] = max(dp[(i+1)*width+j], dp[i*width+j+1])
				}
			}
		}
		for i, j := 0, 0; i < len(a) || j < len(b); {
			if i < len(a) && j < len(b) && a[i] == b[j] {
				appendEdit("equal", a[i])
				i++
				j++
			} else if i < len(a) && (j == len(b) || dp[(i+1)*width+j] >= dp[i*width+j+1]) {
				appendEdit("remove", a[i])
				i++
			} else {
				appendEdit("add", b[j])
				j++
			}
		}
	}
	appendEdit("equal", suffix)
	flush()
	return edits, coarse
}
