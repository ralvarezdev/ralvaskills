package main

import (
	"reflect"
	"testing"

	"github.com/ralvarezdev/termkit/session"

	"github.com/ralvarezdev/ralvaskills/v3/internal/ui"
)

// TestListPinUnpinRowActions pins the row actions on the installed-skills list:
// "p" runs `pin <name>`, "n" runs `pin <name> --remove`.
func TestListPinUnpinRowActions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		key  string
		want []string
	}{
		{"p", []string{"go-architect"}},
		{"n", []string{"go-architect", "--remove"}},
	} {
		target, args, ok := ui.SessionRowAction(session.RowActionEvent{
			Source: listCmd,
			ID:     "go-architect",
			Key:    tc.key,
		})
		if !ok || target != pinCmd || !reflect.DeepEqual(args, tc.want) {
			t.Errorf("key %q: target=%v args=%v ok=%v, want pinCmd %v", tc.key, target, args, ok, tc.want)
		}
	}
}
