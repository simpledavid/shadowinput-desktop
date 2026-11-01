package writing

import (
	"arion.local/core/domain"
	"arion.local/desktop/ui/shared/copy"
)

// Review values cross into the UI through this shared boundary.
type Review = domain.WritingReview
type Issue = domain.WritingIssue
type Change = domain.TextEdit

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
	Compact, Settings, ShowDiff, Reviewed, ProfileReady       bool
	StyleID, TermsDraft                                       string
	Profile                                                   domain.WritingProfile
	StyleDraft                                                domain.WritingStyle
	Review                                                    domain.WritingReview
	ProfileFailure                                            copy.Key
}

func NewState() State {
	return State{Task: "polish", Language: "English", Locale: copy.English, ModelNames: map[string]string{}, Profile: domain.DefaultWritingProfile(), StyleID: "work", ShowDiff: true}
}

func (s *State) CanDeliver() bool {
	return s.Completed && !s.Busy && (len(s.Review.Issues) == 0 || s.Reviewed)
}

func (s *State) InvalidateResult() {
	s.Completed, s.Reviewed = false, false
	s.Output = ""
	s.Review = domain.WritingReview{}
	s.Failure, s.Notice = "", ""
}

// Actions are supplied by the composition root. Views never access credentials.
type Actions struct {
	RequestAccess                                                                                                                           func()
	SignIn, Logout, RefreshModels, Generate, Cancel, Import, Copy, Paste, Hide, ManageUsage, Source, Changed, Preferences                   func()
	EditSource, TaskChanged, StyleChanged, OpenEditor, OpenStyles, SaveStyle, NewStyle, DeleteStyle, SaveTerms, RetryProfile, SwitchAccount func()
	SetDefault                                                                                                                              func(task, style string)
}
