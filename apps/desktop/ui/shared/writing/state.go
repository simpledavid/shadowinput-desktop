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
const ProviderLocal = "local"
const ProviderChatGPT = "chatgpt"

const Shortcut = "CmdOrCtrl+Shift+J"

type State struct {
	InputPage, AICandidates, IMEInstalling                    bool
	IMEStatus                                                 string
	AccessRequired                                            bool
	Provider                                                  string
	Local                                                     LocalStatus
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
	return State{Provider: ProviderLocal, Task: "polish", Language: "English", Locale: copy.English, ModelNames: map[string]string{}, Profile: domain.DefaultWritingProfile(), StyleID: "work", ShowDiff: true}
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

// LocalStatus contains display-only state; no runtime or filesystem handles.
type LocalStatus struct {
	State, Phase, Name     string
	Done, Total, SizeBytes int64
	Installed, Busy        bool
	Failure                copy.Key
}

func (s *State) CanGenerate() bool {
	if s.Busy || !s.ProfileReady || s.Settings {
		return false
	}
	if s.Provider == ProviderLocal {
		return s.Local.State == "ready" && !s.Local.Busy
	}
	return s.Provider == ProviderChatGPT && s.Connected && s.Model != ""
}

func (s *State) CanSuggest() bool {
	return s.AICandidates && s.ProfileReady && !s.Busy && !s.Settings && s.Local.State == "ready" && !s.Local.Busy
}

// Actions are supplied by the composition root. Views never access credentials.
type Actions struct {
	InstallIME, ToggleCandidates, OpenInputSettings                                                                                         func()
	ProviderChanged, DownloadLocal, EnableLocal, DisableLocal, CancelLocal                                                                  func()
	RequestAccess                                                                                                                           func()
	SignIn, Logout, RefreshModels, Generate, Cancel, Import, Copy, Paste, Hide, ManageUsage, Source, Changed, Preferences                   func()
	EditSource, TaskChanged, StyleChanged, OpenEditor, OpenStyles, SaveStyle, NewStyle, DeleteStyle, SaveTerms, RetryProfile, SwitchAccount func()
	SetDefault                                                                                                                              func(task, style string)
}
