package shadowinput

import (
	"context"
	"path/filepath"
	"runtime"
	"time"

	"arion.local/core/domain"
	"arion.local/desktop/infrastructure/provision"
	"arion.local/desktop/infrastructure/rewriting"
	"arion.local/desktop/ui/shared/copy"
	state "arion.local/desktop/ui/shared/writing"
	"arion.local/models"
	"github.com/egoist/mygo"
)

func (a *app) initLocal() {
	catalog, err := models.Load()
	resources, pathErr := mygo.App.Path(mygo.PathResources)
	if err != nil || pathErr != nil {
		a.s.Local = state.LocalStatus{State: "unsupported", Failure: copy.WriteLocalRuntimeError}
		return
	}
	a.local = rewriting.New(catalog, provision.NewStore(filepath.Join(a.directory, "models")), rewriting.Locate(resources), 4, runtime.GOOS+"-"+runtime.GOARCH)
	a.local.OnStatus(func(rewriting.Status) {
		// Coalesce download progress, and read the latest snapshot on the UI thread.
		if a.localStatusQueued.CompareAndSwap(false, true) {
			mygo.RunOnMain(func() {
				a.localStatusQueued.Store(false)
				if !a.quitting {
					a.applyLocalStatus()
					a.invalidate()
				}
			})
		}
	})
	a.applyLocalStatus()
	done := make(chan struct{})
	a.localDone = done
	go func() { defer close(done); a.local.Refresh() }()
}

func (a *app) applyLocalStatus() {
	if a.local == nil {
		return
	}
	s := a.local.Status()
	a.s.Local.State, a.s.Local.Phase, a.s.Local.Name = string(s.State), string(s.Phase), s.Name
	a.s.Local.Done, a.s.Local.Total, a.s.Local.SizeBytes = s.Done, s.Total, s.SizeBytes
	a.s.Local.Installed = s.Installed
	a.s.Local.Failure = ""
	if s.State != rewriting.Failed {
		return
	}
	switch s.Failure {
	case domain.CodeArtifactUnverified:
		a.s.Local.Failure = copy.WriteLocalCorrupt
		a.s.Local.Installed = false
	case domain.CodeProvisioningFailed:
		a.s.Local.Failure = copy.WriteLocalDownloadError
	case "", domain.CodeCancelled:
	default:
		a.s.Local.Failure = copy.WriteLocalRuntimeError
	}
}

// startLocal serializes replacement after cancelling superseded operations.
// Merely selecting local or opening the app never authorizes a download.
func (a *app) startLocal(download, enable bool) {
	if a.local == nil || a.s.Busy {
		return
	}
	if enable && a.s.Provider != state.ProviderLocal {
		return
	}
	if a.panel != nil && enable {
		a.openEditor()
	}
	if a.localCancel != nil {
		a.localCancel()
	}
	a.localJob++
	job := a.localJob
	ctx, cancel := context.WithCancel(context.Background())
	previous, done := a.localDone, make(chan struct{})
	a.localCancel, a.localDone = cancel, done
	a.s.Local.Busy = true
	a.changed()
	a.invalidate()
	go func() {
		defer close(done)
		defer cancel()
		if previous != nil {
			<-previous
		}
		if ctx.Err() == nil {
			a.local.Disable()
			if enable && ctx.Err() == nil {
				if download {
					_ = a.local.Download(ctx)
				} else {
					_ = a.local.Enable(ctx)
				}
			}
		}
		mygo.RunOnMain(func() {
			if a.quitting || job != a.localJob {
				return
			}
			a.localCancel = nil
			a.s.Local.Busy = false
			a.applyLocalStatus()
			a.invalidate()
		})
	}()
}

func (a *app) providerChanged() {
	a.stop()
	a.changed()
	a.savePreferences()
	if a.s.Provider == state.ProviderChatGPT {
		a.startLocal(false, false)
		if a.client == nil {
			a.initAccount()
		} else {
			a.updateAccount()
			if a.s.Connected {
				a.models()
			}
		}
	}
}

func (a *app) closeLocal() {
	if a.localCancel != nil {
		a.localCancel()
	}
	if a.localDone != nil {
		select {
		case <-a.localDone:
		case <-time.After(6 * time.Second):
		}
	}
	if a.local != nil {
		a.local.Disable()
	}
}
