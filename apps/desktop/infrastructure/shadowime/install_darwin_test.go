//go:build darwin

package shadowime

import (
	"strings"
	"testing"
)

func TestObservedStatusRequiresPersistedParentAndMode(t *testing.T) {
	for _, s := range []observation{
		{Complete: true, OwnCount: 2, ParentFound: true},
		{Complete: true, OwnCount: 2, ParentFound: true, Available: true},
		{Complete: true, Selected: true},
		{ParentFound: true, ParentEnabled: true, Available: true},
	} {
		got := observedStatus(s)
		if strings.HasPrefix(got, "已启用") || strings.HasPrefix(got, "已选中") {
			t.Fatalf("claimed usable for %+v: %s", s, got)
		}
	}
	s := observation{Complete: true, OwnCount: 2, ParentFound: true, ParentEnabled: true, Available: true}
	if !strings.HasPrefix(observedStatus(s), "已启用") {
		t.Fatal("enabled source not reported")
	}
	s.Selected = true
	if !strings.HasPrefix(observedStatus(s), "已选中") {
		t.Fatal("selected source not reported")
	}
}
