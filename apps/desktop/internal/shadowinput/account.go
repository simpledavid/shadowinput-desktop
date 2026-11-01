package shadowinput

import (
	"arion.local/desktop/infrastructure/chatgptplan"
	state "arion.local/desktop/ui/shared/writing"
	"time"
)

func (a *app) switchAccount() {
	if a.s.Busy || a.client == nil {
		return
	}
	_, job := a.start(30 * time.Second)
	go func() {
		err := a.client.SwitchAccount()
		a.finish(job, err, func() { a.updateAccount(); a.s.Models = nil; a.s.Model = ""; a.changed(); a.signIn() })
	}()
}

func (a *app) initAccount() {
	if a.s.Provider != state.ProviderChatGPT || a.s.Busy || a.client != nil {
		return
	}
	_, job := a.start(30 * time.Second)
	go func() {
		client, err := chatgptplan.New(chatgptplan.Vault{Directory: a.directory})
		a.finish(job, err, func() {
			a.client = client
			a.updateAccount()
			if a.s.Connected {
				a.models()
			}
		})
	}()
}
