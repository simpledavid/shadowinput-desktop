package writing

import (
	"arion.local/desktop/ui/shared/copy"
	"arion.local/desktop/ui/shared/kit"
	"arion.local/desktop/ui/shared/theme"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo/ui"
	"strings"
)

var taskKeys = map[string]copy.Key{"polish": copy.WritePolish, "translate": copy.WriteTranslate, "rewrite": copy.WriteRewrite, "continue": copy.WriteContinue}

func Editor(c *ui.Context, s *state.State, a state.Actions) {
	t, m := theme.Apply(c), copy.For(s.Locale)
	canRun := !s.Busy && s.Connected && s.ProfileReady && s.Model != "" && strings.TrimSpace(s.Source) != ""
	if c.Shortcut(ui.Cmd, ui.KeyEnter) || c.Shortcut(ui.Ctrl, ui.KeyEnter) {
		if canRun && !s.Settings {
			a.Generate()
		}
	}
	if c.Shortcut(0, ui.KeyEscape) {
		if s.Busy {
			a.Cancel()
		} else {
			a.Hide()
		}
	}
	ui.Column(c).Fill().Padding(18).Gap(12).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(10).Children(func() {
			ui.Text(c, state.Name).FontSize(theme.Heading).FontWeight(650).Grow(1)
			if !s.Compact {
				if kit.Select(c.Key("locale"), &s.Locale, []copy.Locale{copy.English, copy.SimplifiedChinese}, func(l copy.Locale) string {
					if l == copy.English {
						return "English"
					}
					return "简体中文"
				}, nil).Width(115).Label(m.Text(copy.WriteUILanguage)).Changed() {
					a.Preferences()
				}
			}
			if s.Settings {
				kit.Button(c.Key("back"), t, kit.Ghost, m.Text(copy.WriteBack), true).OnClick(a.OpenEditor)
			} else {
				kit.Button(c.Key("styles"), t, kit.Ghost, m.Text(copy.WriteStyles), !s.Busy).OnClick(a.OpenStyles)
			}
			if s.Compact {
				kit.Button(c.Key("expand"), t, kit.Ghost, m.Text(copy.WriteExpand), !s.Busy).OnClick(a.OpenEditor)
			}
		})
		if s.Settings {
			styleSettings(c, s, a, t, m)
			return
		}
		if !s.Compact {
			account(c, s, a, t, m)
		} else {
			ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
				label := s.Email
				if !s.Connected {
					label = m.Text(copy.WriteAccount)
				}
				ui.Text(c, label).FontSize(theme.Caption).TextColor(t.Muted).Grow(1).Wrap()
				kit.Button(c.Key("account"), t, kit.Ghost, m.Text(copy.WriteAccountSettings), true).OnClick(a.OpenEditor)
			})
		}
		ui.Scroll(c.Key("writing-scroll")).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).Gap(12).Children(func() {
				ui.Row(c).Gap(4).Children(func() {
					for _, task := range []string{"polish", "translate", "rewrite", "continue"} {
						variant := kit.Ghost
						if s.Task == task {
							variant = kit.Outline
						}
						kit.Button(c.Key(task), t, variant, m.Text(taskKeys[task]), !s.Busy).Grow(1).OnClick(func() { s.Task = task; a.TaskChanged() })
					}
				})
				ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
					stylePicker(c.Key("style"), s, a, m).Grow(1)
					if len(s.Models) > 0 {
						if kit.Select(c.Key("model"), &s.Model, s.Models, func(id string) string { return s.ModelNames[id] }, nil).Width(210).Label(m.Text(copy.WriteModel)).Disabled(s.Busy).Changed() {
							a.Changed()
							a.Preferences()
						}
					}
				})
				ui.Row(c).Children(func() {
					label := copy.WriteDraft
					if s.HasTarget {
						label = copy.WriteSelection
					}
					ui.Text(c, m.Text(label)).FontSize(theme.Caption).TextColor(t.Muted).Grow(1)
					kit.Button(c.Key("import"), t, kit.Ghost, m.Text(copy.WriteImport), !s.Busy).OnClick(a.Import)
				})
				height := float32(120)
				if s.Compact {
					height = 88
				}
				if ui.TextArea(c.Key("draft"), &s.Source).Height(height).Label(m.Text(copy.WriteDraft)).Placeholder(m.Text(copy.WriteDraftHint)).ReadOnly(s.HasTarget).Disabled(s.Busy).Changed() {
					a.EditSource()
				}
				if s.Task == "translate" {
					if ui.TextInput(c.Key("language"), &s.Language).Label(m.Text(copy.WriteTargetLanguage)).Placeholder(m.Text(copy.WriteTargetLanguage)).Disabled(s.Busy).Changed() {
						a.Changed()
					}
				}
				if ui.TextInput(c.Key("instruction"), &s.Instruction).Placeholder(m.Text(copy.WriteInstruction)).Label(m.Text(copy.WriteInstruction)).Disabled(s.Busy).Changed() {
					a.Changed()
				}
				result(c, s, t, m)
			})
		})
		if s.Completed && len(s.Review.Issues) > 0 && !s.Reviewed {
			ui.Text(c, m.Text(copy.WriteRiskPending)).Wrap().FontSize(theme.Caption).TextColor(t.Warning)
		}
		if s.ProfileFailure != "" {
			ui.Text(c, m.Text(s.ProfileFailure)).Wrap().TextColor(t.Danger)
		}
		if s.Failure != "" {
			recovery(c, s, a, t, m)
		} else if s.Notice != "" {
			ui.Text(c, m.Text(s.Notice)).Wrap().FontSize(theme.Caption).TextColor(t.Muted)
		}
		ui.Row(c).Gap(8).Children(func() {
			if s.Busy {
				kit.Button(c.Key("cancel"), t, kit.Outline, m.Text(copy.WriteCancel), true).OnClick(a.Cancel)
			} else {
				kit.Button(c.Key("generate"), t, kit.Primary, m.Text(copy.WriteGenerate), canRun).OnClick(a.Generate)
			}
			ui.Spacer(c)
			kit.Button(c.Key("copy"), t, kit.Outline, m.Text(copy.WriteCopy), s.CanDeliver()).OnClick(a.Copy)
			kit.Button(c.Key("paste"), t, kit.Primary, m.Text(copy.WriteApply), s.CanDeliver() && s.HasTarget).OnClick(a.Paste)
		})
		ui.Text(c, m.Text(copy.WriteKeyboardHint)).FontSize(theme.Caption).TextColor(t.Muted).Wrap()
	})
}
