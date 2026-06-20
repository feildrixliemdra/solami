package app

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/feildrix/solami/internal/config"
	"github.com/feildrix/solami/internal/storage"
	"github.com/feildrix/solami/internal/tui"
	"github.com/feildrix/solami/internal/wallet"
)

// Run starts the Solami TUI.
func Run(ctx context.Context) error {
	paths, err := storage.DefaultPaths()
	if err != nil {
		return err
	}
	if err := paths.Ensure(); err != nil {
		return err
	}
	cfg, err := config.Load(paths.Config)
	if err != nil {
		return err
	}
	if err := config.Save(paths.Config, cfg); err != nil {
		return err
	}
	store := wallet.NewStore(paths.Wallet)
	model, err := tui.NewModel(ctx, paths, cfg, store)
	if err != nil {
		return err
	}
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run tui: %w", err)
	}
	return nil
}
