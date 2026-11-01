package writing

import (
	"arion.local/desktop/ui/shared/copy"
	"arion.local/desktop/ui/shared/kit"
	"arion.local/desktop/ui/shared/theme"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo/ui"
)

func provider(c *ui.Context, s *state.State, a state.Actions, t theme.Tokens, m copy.Catalog) {
	ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
		if kit.Select(c.Key("provider"), &s.Provider, []string{state.ProviderLocal, state.ProviderChatGPT}, func(id string) string {
			if id == state.ProviderLocal {
				return m.Text(copy.WriteProviderLocal)
			}
			return m.Text(copy.WriteProviderChatGPT)
		}, nil).Width(180).Label(m.Text(copy.WriteProvider)).Disabled(s.Busy).Changed() {
			a.ProviderChanged()
		}
		ui.Spacer(c)
		if s.AccessRequired {
			kit.Button(c.Key("access"), t, kit.Ghost, m.Text(copy.WriteAccess), !s.Busy).OnClick(a.RequestAccess)
		}
	})
	if s.Provider == state.ProviderLocal {
		localModel(c, s, a, t, m)
		return
	}
	if !s.Compact {
		account(c, s, a, t, m)
		return
	}
	ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
		label := s.Email
		if !s.Connected {
			label = m.Text(copy.WriteAccount)
		}
		ui.Text(c, label).FontSize(theme.Caption).TextColor(t.Muted).Grow(1).Wrap()
		kit.Button(c.Key("account"), t, kit.Ghost, m.Text(copy.WriteAccountSettings), true).OnClick(a.OpenEditor)
	})
}

func localModel(c *ui.Context, s *state.State, a state.Actions, t theme.Tokens, m copy.Catalog) {
	label := copy.WriteLocalOff
	switch s.Local.State {
	case "ready":
		label = copy.WriteLocalReady
	case "loading":
		label = copy.WriteLocalLoading
	case "unsupported":
		label = copy.WriteLocalUnsupported
	case "provisioning":
		label = copy.WriteLocalChecking
		switch s.Local.Phase {
		case "downloading":
			label = copy.WriteLocalDownloading
		case "verifying":
			label = copy.WriteLocalVerifying
		}
	}
	if s.Local.Busy && s.Local.State == "off" {
		label = copy.WriteLocalChecking
	}
	if s.Compact {
		ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
			ui.Text(c, m.Text(label)).Grow(1).Wrap().FontSize(theme.Caption).TextColor(t.Muted)
			kit.Button(c.Key("local-settings"), t, kit.Ghost, m.Text(copy.WriteExpand), !s.Busy).OnClick(a.OpenEditor)
		})
		return
	}
	kit.Card(c, t).Padding(12).Children(func() {
		ui.Column(c).Gap(6).Children(func() {
			ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
				ui.Column(c).Grow(1).Gap(3).Children(func() {
					name := s.Local.Name
					if name == "" {
						name = m.Text(copy.WriteProviderLocal)
					}
					ui.Text(c, name).FontWeight(550).Wrap()
					ui.Text(c, m.Text(label)).FontSize(theme.Caption).TextColor(t.Muted).Wrap()
				})
				switch {
				case s.Local.Busy:
					kit.Button(c.Key("local-cancel"), t, kit.Outline, m.Text(copy.WriteLocalCancel), true).OnClick(a.CancelLocal)
				case s.Local.State == "ready":
					kit.Button(c.Key("local-stop"), t, kit.Ghost, m.Text(copy.WriteLocalDisable), !s.Busy).OnClick(a.DisableLocal)
				case s.Local.State == "unsupported":
				case s.Local.Installed:
					kit.Button(c.Key("local-enable"), t, kit.Primary, m.Text(copy.WriteLocalEnable), !s.Busy).OnClick(a.EnableLocal)
				default:
					kit.Button(c.Key("local-download"), t, kit.Primary, m.Text(copy.WriteLocalDownload), !s.Busy).OnClick(a.DownloadLocal)
				}
			})
			if s.Local.Busy && s.Local.Total > 0 {
				fraction := float64(s.Local.Done) / float64(s.Local.Total)
				ui.Progress(c.Key("local-progress"), fraction).Label(m.Text(label))
				ui.Text(c, m.Progress(s.Local.Done, s.Local.Total)).FontSize(theme.Caption).TextColor(t.Muted)
			} else if !s.Local.Installed && s.Local.SizeBytes > 0 {
				ui.Text(c, m.DownloadSize(s.Local.SizeBytes)).FontSize(theme.Caption).TextColor(t.Muted)
			}
			if s.Local.Failure != "" {
				ui.Text(c, m.Text(s.Local.Failure)).Wrap().FontSize(theme.Caption).TextColor(t.Danger)
			}
			ui.Text(c, m.Text(copy.WriteLocalPrivacy)).Wrap().FontSize(theme.Caption).TextColor(t.Muted)
		})
	})
}
