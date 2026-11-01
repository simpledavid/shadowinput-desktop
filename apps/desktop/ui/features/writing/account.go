package writing

import (
	"arion.local/desktop/ui/shared/copy"
	"arion.local/desktop/ui/shared/icons"
	"arion.local/desktop/ui/shared/kit"
	"arion.local/desktop/ui/shared/theme"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo/ui"
)

func account(c *ui.Context, s *state.State, a state.Actions, t theme.Tokens, m copy.Catalog) {
	kit.Card(c, t).Padding(12).Children(func() {
		ui.Column(c).Gap(8).Children(func() {
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, icons.OpenAI).FontSize(20)
				ui.Column(c).Grow(1).Gap(3).Children(func() {
					label := m.Text(copy.WriteAccount)
					if s.Connected {
						label = s.Email
					}
					ui.Text(c, label).FontWeight(550).Wrap()
					label = m.Text(copy.WriteAccountDetail)
					if s.Connected {
						label = m.Text(copy.WritePlan)
					}
					ui.Text(c, label).FontSize(theme.Caption).TextColor(t.Muted).Wrap()
				})
				if s.Connected {
					kit.Button(c.Key("usage"), t, kit.Ghost, m.Text(copy.WriteUsage), true).OnClick(a.ManageUsage)
				} else {
					kit.Button(c.Key("login"), t, kit.Primary, m.Text(copy.WriteSignIn), !s.Busy).OnClick(a.SignIn)
				}
			})
			ui.Text(c, m.Text(copy.WriteQuotaDetail)).FontSize(theme.Caption).TextColor(t.Muted).Wrap()
			if s.Connected {
				ui.Row(c).Gap(8).Children(func() {
					kit.Button(c.Key("refresh"), t, kit.Ghost, m.Text(copy.WriteRefresh), !s.Busy).OnClick(a.RefreshModels)
					kit.Button(c.Key("switch-account"), t, kit.Ghost, m.Text(copy.WriteSwitchAccount), !s.Busy).OnClick(a.SwitchAccount)
					kit.Button(c.Key("logout"), t, kit.Ghost, m.Text(copy.WriteSignOut), !s.Busy).OnClick(a.Logout)
				})
			}
			if s.AccessRequired {
				kit.Button(c.Key("access"), t, kit.Outline, m.Text(copy.WriteAccess), !s.Busy).OnClick(a.RequestAccess)
			}
		})
	})
}

func recovery(c *ui.Context, s *state.State, a state.Actions, t theme.Tokens, m copy.Catalog) {
	ui.Column(c).Gap(6).Children(func() {
		ui.Text(c, m.Text(s.Failure)).Wrap().FontSize(theme.Caption).TextColor(t.Danger)
		switch s.Failure {
		case copy.WriteAuthError, copy.WritePermissionError:
			kit.Button(c.Key("reauthorize"), t, kit.Outline, m.Text(copy.WriteReauthorize), !s.Busy).OnClick(a.SignIn)
		case copy.WriteLimitError:
			kit.Button(c.Key("limit-usage"), t, kit.Outline, m.Text(copy.WriteUsage), true).OnClick(a.ManageUsage)
		case copy.WriteModelError:
			kit.Button(c.Key("reload-models"), t, kit.Outline, m.Text(copy.WriteRefresh), !s.Busy).OnClick(a.RefreshModels)
		}
	})
}
