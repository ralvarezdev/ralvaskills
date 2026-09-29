package ui

var (
	// SuccessMark prints a green ✓ line.
	SuccessMark = Theme.SuccessMark()

	// WarnMark prints an amber ⚠ line.
	WarnMark = Theme.WarnMark()

	// ErrorMark prints a red ✗ line.
	ErrorMark = Theme.ErrorMark()

	// Arrow prints a muted → line.
	Arrow = Theme.ArrowMark()

	// ReLink prints a warn-colored "re-link" badge for skills that need to be re-linked.
	ReLink = ReLinkStyle.Render("re-link")
)
