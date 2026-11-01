package writing

import (
	"arion.local/desktop/ui/shared/copy"
	"arion.local/desktop/ui/shared/kit"
	"arion.local/desktop/ui/shared/theme"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo/ui"
)

var builtinStyles = map[string]copy.Key{"chat": copy.WriteStyleChat, "work": copy.WriteStyleWork, "email": copy.WriteStyleEmail}
var toneKeys = map[string]copy.Key{"natural": copy.WriteToneNatural, "professional": copy.WriteToneProfessional, "warm": copy.WriteToneWarm}
var lengthKeys = map[string]copy.Key{"keep": copy.WriteLengthKeep, "short": copy.WriteLengthShort, "detailed": copy.WriteLengthDetailed}

func styleIDs(s *state.State) []string {
	ids := []string{""}
	for _, style := range s.Profile.Styles {
		ids = append(ids, style.ID)
	}
	return ids
}
func styleName(s *state.State, m copy.Catalog, id string) string {
	if style, ok := s.Profile.Style(id); ok {
		if style.Name != "" {
			return style.Name
		}
		if key, ok := builtinStyles[id]; ok {
			return m.Text(key)
		}
	}
	return m.Text(copy.WriteStyleNone)
}
func stylePicker(c *ui.Context, s *state.State, a state.Actions, m copy.Catalog) ui.Element {
	e := kit.Select(c, &s.StyleID, styleIDs(s), func(id string) string { return styleName(s, m, id) }, nil).Label(m.Text(copy.WriteStyle)).Disabled(s.Busy || !s.ProfileReady)
	if e.Changed() {
		a.StyleChanged()
	}
	return e
}
func profileFailure(c *ui.Context, s *state.State, a state.Actions, t theme.Tokens, m copy.Catalog) {
	if s.ProfileFailure != "" {
		ui.Text(c, m.Text(s.ProfileFailure)).Wrap().TextColor(t.Danger)
		if !s.ProfileReady {
			kit.Button(c.Key("retry-profile"), t, kit.Outline, m.Text(copy.WriteRetryProfile), true).OnClick(a.RetryProfile)
		}
	}
}
func styleSettings(c *ui.Context, s *state.State, a state.Actions, t theme.Tokens, m copy.Catalog) {
	ui.Text(c, m.Text(copy.WriteStylePrivacy)).Wrap().FontSize(theme.Caption).TextColor(t.Muted)
	profileFailure(c, s, a, t, m)
	if !s.ProfileReady {
		return
	}
	ui.Scroll(c.Key("settings-scroll")).Grow(1).MinHeight(0).Children(func() {
		ui.Column(c).Gap(12).Children(func() {
			ui.Row(c).Gap(8).Children(func() {
				stylePicker(c.Key("edit-style"), s, a, m).Grow(1)
				kit.Button(c.Key("new-style"), t, kit.Outline, m.Text(copy.WriteNewStyle), len(s.Profile.Styles) < 24).OnClick(a.NewStyle)
				kit.Button(c.Key("delete-style"), t, kit.Ghost, m.Text(copy.WriteDeleteStyle), s.StyleID != "").OnClick(a.DeleteStyle)
			})
			if s.StyleID != "" {
				ui.TextInput(c.Key("style-name"), &s.StyleDraft.Name).Label(m.Text(copy.WriteStyleName)).Placeholder(styleName(s, m, s.StyleID))
				ui.Row(c).Gap(10).Children(func() {
					ui.Column(c).Grow(1).Gap(5).Children(func() {
						ui.Text(c, m.Text(copy.WriteTone)).FontSize(theme.Caption)
						kit.Select(c.Key("tone"), &s.StyleDraft.Tone, []string{"natural", "professional", "warm"}, func(id string) string { return m.Text(toneKeys[id]) }, nil).Label(m.Text(copy.WriteTone)).FillWidth()
					})
					ui.Column(c).Grow(1).Gap(5).Children(func() {
						ui.Text(c, m.Text(copy.WriteLength)).FontSize(theme.Caption)
						kit.Select(c.Key("length"), &s.StyleDraft.Length, []string{"keep", "short", "detailed"}, func(id string) string { return m.Text(lengthKeys[id]) }, nil).Label(m.Text(copy.WriteLength)).FillWidth()
					})
				})
				ui.TextInput(c.Key("style-language"), &s.StyleDraft.Language).Label(m.Text(copy.WriteStyleLanguage)).Placeholder(m.Text(copy.WriteStyleLanguage))
				ui.TextArea(c.Key("style-rules"), &s.StyleDraft.Rules).Height(100).Label(m.Text(copy.WriteStyleRules)).Placeholder(m.Text(copy.WriteStyleRules))
				kit.Button(c.Key("save-style"), t, kit.Primary, m.Text(copy.WriteSaveStyle), true).OnClick(a.SaveStyle)
			}
			ui.Text(c, m.Text(copy.WriteDefaults)).FontWeight(600)
			for _, task := range []string{"polish", "translate", "rewrite", "continue"} {
				ui.Row(c).AlignItems(ui.Center).Gap(12).Children(func() {
					ui.Text(c, m.Text(taskKeys[task])).Width(90)
					id := s.Profile.Defaults[task]
					if kit.Select(c.Key("default-"+task), &id, styleIDs(s), func(id string) string { return styleName(s, m, id) }, nil).Label(m.Text(taskKeys[task])).Grow(1).Changed() {
						a.SetDefault(task, id)
					}
				})
			}
			ui.Text(c, m.Text(copy.WriteTerms)).FontWeight(600)
			ui.Text(c, m.Text(copy.WriteTermsHint)).Wrap().FontSize(theme.Caption).TextColor(t.Muted)
			ui.TextArea(c.Key("terms"), &s.TermsDraft).Height(110).Label(m.Text(copy.WriteTerms))
			kit.Button(c.Key("save-terms"), t, kit.Primary, m.Text(copy.WriteSaveTerms), true).OnClick(a.SaveTerms)
		})
	})
	if s.Notice != "" {
		ui.Text(c, m.Text(s.Notice)).Wrap().FontSize(theme.Caption).TextColor(t.Muted)
	}
}
