package tui

import "github.com/charmbracelet/huh"

// newPasswordConfirmationForm builds the reusable sensitive-action password form.
func newPasswordConfirmationForm(password *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(password),
		),
	)
}
