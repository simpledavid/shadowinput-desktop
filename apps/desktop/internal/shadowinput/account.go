package shadowinput

import "time"

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
