package writing

import (
	feature "arion.local/desktop/ui/features/writing"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo/ui"
)

func View(s *state.State, a state.Actions) func(*ui.Context) {
	return func(c *ui.Context) { feature.Editor(c, s, a) }
}
