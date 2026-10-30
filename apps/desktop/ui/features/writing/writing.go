package writing

import (
	"arion.local/desktop/ui/shared/copy"
	"arion.local/desktop/ui/shared/icons"
	"arion.local/desktop/ui/shared/kit"
	"arion.local/desktop/ui/shared/theme"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo/ui"
	"strings"
)

var taskKeys = map[string]copy.Key{"polish": copy.WritePolish, "translate": copy.WriteTranslate, "rewrite": copy.WriteRewrite, "continue": copy.WriteContinue}

func Editor(c *ui.Context, s *state.State, a state.Actions) {
	t := theme.Apply(c)
	m := copy.For(s.Locale)
	canRun := !s.Busy && s.Connected && s.Model != "" && strings.TrimSpace(s.Source) != ""
	if c.Shortcut(ui.Cmd, ui.KeyEnter) || c.Shortcut(ui.Ctrl, ui.KeyEnter) {
		if canRun {
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
	ui.Column(c).Fill().Padding(24).Gap(16).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(12).Children(func() {
			ui.Column(c).Grow(1).Gap(4).Children(func() {
				ui.Text(c, state.Name).FontSize(theme.Display).FontWeight(650)
				ui.Text(c, m.Text(copy.WriteTagline)).FontSize(theme.Body).TextColor(t.Muted)
			})
			if kit.Select(c.Key("locale"), &s.Locale, []copy.Locale{copy.English, copy.SimplifiedChinese}, func(l copy.Locale) string {
				if l == copy.English {
					return "English"
				}
				return "简体中文"
			}, nil).Width(120).Label(m.Text(copy.WriteUILanguage)).Changed() {
				a.Preferences()
			}
		})
		kit.Card(c, t).Padding(12, 16).Children(func() {
			ui.Row(c).AlignItems(ui.Center).Gap(10).Children(func() {
				ui.Icon(c, icons.OpenAI).FontSize(20)
				ui.Column(c).Grow(1).Gap(4).Children(func() {
					if s.Connected {
						ui.Text(c, s.Email).FontWeight(500)
						ui.Text(c, m.Text(copy.WritePlan)).FontSize(theme.Caption).TextColor(t.Muted)
					} else {
						ui.Text(c, m.Text(copy.WriteAccount))
						ui.Text(c, m.Text(copy.WriteAccountDetail)).FontSize(theme.Caption).TextColor(t.Muted)
					}
				})
				if s.Connected {
					kit.Button(c.Key("usage"), t, kit.Ghost, m.Text(copy.WriteUsage), true).OnClick(a.ManageUsage)
					kit.Button(c.Key("logout"), t, kit.Ghost, m.Text(copy.WriteSignOut), !s.Busy).OnClick(a.Logout)
				} else {
					kit.Button(c.Key("login"), t, kit.Outline, m.Text(copy.WriteSignIn), !s.Busy).OnClick(a.SignIn)
				}
			})
		})
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			for _, task := range []string{"polish", "translate", "rewrite", "continue"} {
				variant := kit.Ghost
				if s.Task == task {
					variant = kit.Outline
				}
				kit.Button(c.Key(task), t, variant, m.Text(taskKeys[task]), !s.Busy).Height(40).OnClick(func() { s.Task = task; a.Changed() })
			}
			ui.Spacer(c)
			kit.Button(c.Key("import"), t, kit.Ghost, m.Text(copy.WriteImport), !s.Busy).OnClick(a.Import)
		})
		ui.Column(c).Gap(8).Children(func() {
			ui.Text(c, m.Text(copy.WriteDraft)).FontSize(theme.Caption).TextColor(t.Muted)
			if ui.TextArea(c.Key("draft"), &s.Source).Height(140).Label(m.Text(copy.WriteDraft)).Placeholder(m.Text(copy.WriteDraftHint)).Disabled(s.Busy).Changed() {
				a.Changed()
			}
			if s.Task == "translate" {
				if ui.TextInput(c.Key("language"), &s.Language).Label(m.Text(copy.WriteTargetLanguage)).Placeholder(m.Text(copy.WriteTargetLanguage)).Disabled(s.Busy).Changed() {
					a.Changed()
				}
			}
			if ui.TextInput(c.Key("instruction"), &s.Instruction).Placeholder(m.Text(copy.WriteInstruction)).Label(m.Text(copy.WriteInstruction)).Disabled(s.Busy).Changed() {
				a.Changed()
			}
		})
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			if len(s.Models) > 0 {
				if kit.Select(c.Key("model"), &s.Model, s.Models, func(id string) string { return s.ModelNames[id] }, nil).MaxWidth(310).Label(m.Text(copy.WriteModel)).Disabled(s.Busy).Changed() {
					a.Changed()
					a.Preferences()
				}
			} else {
				ui.Text(c, m.Text(copy.WriteModelEmpty)).FontSize(theme.Caption).TextColor(t.Muted)
			}
			if s.Connected {
				kit.IconButton(c.Key("refresh"), t, icons.RotateCw, m.Text(copy.WriteRefresh)).Disabled(s.Busy).OnClick(a.RefreshModels)
			}
			ui.Spacer(c)
			if s.Busy {
				kit.Button(c.Key("cancel"), t, kit.Outline, m.Text(copy.WriteCancel), true).OnClick(a.Cancel)
			} else {
				kit.Button(c.Key("generate"), t, kit.Primary, m.Text(copy.WriteGenerate), canRun).Height(40).OnClick(a.Generate)
			}
		})
		ui.Column(c).Grow(1).MinHeight(100).Gap(8).Children(func() {
			ui.Row(c).Children(func() {
				ui.Text(c, m.Text(copy.WritePreview)).FontSize(theme.Caption).TextColor(t.Muted)
				ui.Spacer(c)
				if s.Busy {
					ui.Text(c, m.Text(copy.WriteWorking)).FontSize(theme.Caption).TextColor(t.Muted)
				}
			})
			ui.TextArea(c.Key("result"), &s.Output).Grow(1).MinHeight(80).ReadOnly(true).Label(m.Text(copy.WritePreview)).Placeholder(m.Text(copy.WritePreviewHint))
		})
		if s.Failure != "" {
			ui.Text(c, m.Text(s.Failure)).Wrap().FontSize(theme.Body).TextColor(t.Danger)
		} else if s.Notice != "" {
			ui.Text(c, m.Text(s.Notice)).Wrap().FontSize(theme.Caption).TextColor(t.Muted)
		}
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			kit.Button(c.Key("attribution"), t, kit.Ghost, m.Text(copy.WriteAttribution), true).OnClick(a.Source)
			ui.Spacer(c)
			if s.AccessRequired {
				kit.Button(c.Key("access"), t, kit.Ghost, m.Text(copy.WriteAccess), !s.Busy).OnClick(a.RequestAccess)
			}
			kit.Button(c.Key("copy"), t, kit.Outline, m.Text(copy.WriteCopy), s.Completed && !s.Busy).OnClick(a.Copy)
			kit.Button(c.Key("paste"), t, kit.Primary, m.Text(copy.WritePaste), s.Completed && !s.Busy && s.HasTarget).OnClick(a.Paste)
		})
		ui.Text(c, m.Text(copy.WritePrivacy)).FontSize(theme.Caption).TextColor(t.Muted).Wrap()
	})
}
