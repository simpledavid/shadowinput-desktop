package writing

import (
	"arion.local/desktop/ui/shared/copy"
	"arion.local/desktop/ui/shared/kit"
	"arion.local/desktop/ui/shared/theme"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo/ui"
)

var issueKeys = map[string]copy.Key{"numbers": copy.WriteRiskNumbers, "dates": copy.WriteRiskDates, "terms": copy.WriteRiskTerms, "commitments": copy.WriteRiskCommitments}

func result(c *ui.Context, s *state.State, t theme.Tokens, m copy.Catalog) {
	ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
		ui.Text(c, m.Text(copy.WritePreview)).FontSize(theme.Caption).TextColor(t.Muted).Grow(1)
		if s.Completed {
			label := copy.WriteShowDiff
			if s.ShowDiff {
				label = copy.WriteShowResult
			}
			kit.Button(c.Key("diff"), t, kit.Ghost, m.Text(label), true).OnClick(func() { s.ShowDiff = !s.ShowDiff })
		}
	})
	if s.Completed && s.ShowDiff {
		ui.Scroll(c.Key("diff-scroll")).Height(150).Children(func() {
			spans := make([]ui.Span, 0, len(s.Review.Edits))
			for _, e := range s.Review.Edits {
				span := ui.Span{Text: e.Text}
				if e.Kind == "remove" {
					span.Color = t.Danger
					span.Strikethrough = true
				}
				if e.Kind == "add" {
					span.Color = t.Success
					span.Underline = true
				}
				spans = append(spans, span)
			}
			ui.RichText(c, spans...).Wrap().FontSize(theme.Body).Selectable()
		})
		ui.Text(c, m.Text(copy.WriteDiffLegend)).FontSize(theme.Caption).TextColor(t.Muted)
		if s.Review.Coarse {
			ui.Text(c, m.Text(copy.WriteDiffCoarse)).Wrap().FontSize(theme.Caption).TextColor(t.Muted)
		}
	} else {
		ui.TextArea(c.Key("result"), &s.Output).Height(150).ReadOnly(true).Label(m.Text(copy.WritePreview)).Placeholder(m.Text(copy.WritePreviewHint))
	}
	if s.Completed && len(s.Review.Issues) > 0 {
		kit.Card(c, t).Padding(12).Children(func() {
			ui.Column(c).Gap(8).Children(func() {
				ui.Text(c, m.Text(copy.WriteRiskTitle)).FontWeight(600)
				for _, issue := range s.Review.Issues {
					ui.Text(c, m.Text(issueKeys[issue.Kind])).Wrap().TextColor(t.Warning).FontSize(theme.Caption)
					if issue.Before != "" || issue.After != "" {
						ui.Text(c, issue.Before+" → "+issue.After).Wrap().FontSize(theme.Caption)
					}
				}
				ui.Text(c, m.Text(copy.WriteRiskHint)).Wrap().FontSize(theme.Caption).TextColor(t.Muted)
				ui.Checkbox(c.Key("reviewed"), &s.Reviewed, m.Text(copy.WriteReviewed))
			})
		})
	}
}
