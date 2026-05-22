package ui

var (
	// SuccessMark prints a green ✓ line.
	SuccessMark = SuccessStyle.Render("✓")
	
	// WarnMark prints an amber ⚠ line.
	WarnMark = WarnStyle.Render("⚠")

	// ErrorMark prints a red ✗ line.
	ErrorMark = ErrorStyle.Render("✗")

	// Arrow prints a muted → line.
	Arrow = ArrowStyle.Render("→")

	// ReLink prints a warn-colored "re-link" badge for skills that need to be re-linked.
	ReLink = ReLinkStyle.Render("re-link")
)
