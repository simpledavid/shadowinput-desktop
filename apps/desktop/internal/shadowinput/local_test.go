package shadowinput

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"arion.local/desktop/infrastructure/llm/engine"
	"arion.local/desktop/infrastructure/localwriting"
	"arion.local/desktop/ui/shared/copy"
	state "arion.local/desktop/ui/shared/writing"
)

func TestProviderPreferencesAndLegacyMigration(t *testing.T) {
	a := &app{s: state.NewState(), directory: t.TempDir()}
	a.loadPreferences()
	if a.s.Provider != state.ProviderLocal {
		t.Fatal("new install not local")
	}
	if err := os.WriteFile(filepath.Join(a.directory, "preferences.json"), []byte(`{"Locale":"en","Model":"saved-cloud-model"}`), 0600); err != nil {
		t.Fatal(err)
	}
	a.loadPreferences()
	if a.s.Provider != state.ProviderChatGPT || a.s.Model != "saved-cloud-model" {
		t.Fatal("legacy preference lost")
	}
	a.s.Provider = state.ProviderLocal
	a.savePreferences()
	b := &app{s: state.NewState(), directory: a.directory}
	b.loadPreferences()
	if b.s.Provider != state.ProviderLocal || b.s.Model != "saved-cloud-model" {
		t.Fatal("provider roundtrip lost cloud choice")
	}
	b.s.Source = "kept draft"
	b.s.Output = "stale"
	b.s.Completed = true
	b.s.Reviewed = true
	b.providerChanged()
	if b.s.Source != "kept draft" || b.s.Completed || b.s.Reviewed || b.client != nil {
		t.Fatal("local selection lost draft, approval or accessed account")
	}
}

func TestLocalFailureMessages(t *testing.T) {
	for err, key := range map[error]copy.Key{engine.ErrInputBudget: copy.WriteLocalInputBudget, engine.ErrOutputBudget: copy.WriteLocalOutputBudget, engine.ErrControlInput: copy.WriteLocalControlInput, localwriting.ErrIncomplete: copy.WriteLocalOutputBudget, context.DeadlineExceeded: copy.WriteTimeoutError} {
		if failureKey(err) != key {
			t.Fatal(err)
		}
	}
}
