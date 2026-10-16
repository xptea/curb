package app

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestInvalidURLFeedback(t *testing.T) {
	a := newApplication("test")
	tt := ui.NewTester(a.view, 900, 640)
	if err := tt.Click("Media URL"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEnter)
	if !strings.Contains(a.Failure, "valid") || a.Busy {
		t.Fatalf("invalid input accepted: %+v", a)
	}
}
