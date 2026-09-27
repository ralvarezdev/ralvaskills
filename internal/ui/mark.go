package ui

import "github.com/ralvarezdev/termkit"

var (
	// SuccessMark prints a green ✓ line.
	SuccessMark = termkit.SuccessMark

	// WarnMark prints an amber ⚠ line.
	WarnMark = termkit.WarnMark

	// ErrorMark prints a red ✗ line.
	ErrorMark = termkit.ErrorMark

	// Arrow prints a muted → line.
	Arrow = termkit.ArrowMark

	// ReLink prints a warn-colored "re-link" badge for skills that need to be re-linked.
	ReLink = ReLinkStyle.Render("re-link")
)
