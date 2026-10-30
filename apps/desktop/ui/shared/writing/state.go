package writing

import "arion.local/desktop/ui/shared/copy"

const Name = "shadowinput"
const Shortcut = "CmdOrCtrl+Shift+J"

type State struct {
	AccessRequired                                            bool
	Source, Output, Task, Instruction, Language, Model, Email string
	Models                                                    []string
	ModelNames                                                map[string]string
	Locale                                                    copy.Locale
	Busy, Connected, Completed, HasTarget                     bool
	Failure, Notice                                           copy.Key
}

func NewState() State {
	return State{Task: "polish", Language: "English", Locale: copy.English, ModelNames: map[string]string{}}
}

// Actions are supplied by the composition root. Views never access credentials.
type Actions struct {
	RequestAccess                                                                                                         func()
	SignIn, Logout, RefreshModels, Generate, Cancel, Import, Copy, Paste, Hide, ManageUsage, Source, Changed, Preferences func()
}
