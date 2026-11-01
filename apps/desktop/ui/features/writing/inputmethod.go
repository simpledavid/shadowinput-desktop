package writing

import (
	"arion.local/desktop/ui/shared/copy"
	"arion.local/desktop/ui/shared/kit"
	"arion.local/desktop/ui/shared/theme"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo/ui"
)

func inputSettings(c *ui.Context, s *state.State, a state.Actions, t theme.Tokens, m copy.Catalog) {
	ui.Scroll(c.Key("input-settings-scroll")).Grow(1).MinHeight(0).Children(func() {
		ui.Column(c).Gap(18).Children(func() {
			ui.Text(c, m.Text(copy.ShadowInputSettings)).FontSize(theme.Heading).FontWeight(650)
			ui.Text(c, m.Text(copy.ShadowInputSteps)).Wrap().TextColor(t.Muted)
			ui.Row(c).Gap(10).Children(func() {
				kit.Button(c.Key("install-ime"), t, kit.Primary, m.Text(copy.ShadowInstallIME), !s.IMEInstalling).OnClick(a.InstallIME)
				ui.Text(c, s.IMEStatus).Wrap().Grow(1).FontSize(theme.Caption)
			})
			ui.Text(c, m.Text(copy.ShadowCandidateKeys)).Wrap()
			provider(c, s, a, t, m)
			if s.Failure != "" {
				recovery(c, s, a, t, m)
			}
			if s.Busy {
				kit.Button(c.Key("ime-cancel"), t, kit.Ghost, m.Text(copy.WriteCancel), true).OnClick(a.Cancel)
			}
			if s.Provider == state.ProviderChatGPT && len(s.Models) > 0 {
				if kit.Select(c.Key("ime-model"), &s.Model, s.Models, func(id string) string { return s.ModelNames[id] }, nil).Label(m.Text(copy.WriteModel)).Changed() {
					a.Changed()
					a.Preferences()
				}
			}
			stylePicker(c.Key("ime-style"), s, a, m)
			detail := copy.ShadowAILocal
			if s.Provider == state.ProviderChatGPT {
				detail = copy.ShadowAICloud
			}
			ui.Text(c, m.Text(detail)).Wrap().FontSize(theme.Caption).TextColor(t.Muted)
			key := copy.ShadowAIEnable
			if s.AICandidates {
				key = copy.ShadowAIDisable
			}
			kit.Button(c.Key("ai-candidates"), t, kit.Outline, m.Text(key), true).OnClick(a.ToggleCandidates)
			ui.Text(c, m.Text(copy.ShadowAINote)).Wrap().FontSize(theme.Caption).TextColor(t.Muted)
			kit.Button(c.Key("open-writing"), t, kit.Ghost, m.Text(copy.ShadowWritingTool), true).OnClick(a.OpenEditor)
		})
	})
}
