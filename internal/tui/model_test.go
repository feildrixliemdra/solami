package tui

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/feildrix/solami/internal/config"
	"github.com/feildrix/solami/internal/networks"
	"github.com/feildrix/solami/internal/storage"
	"github.com/feildrix/solami/internal/wallet"
)

func TestNewModelRoutesFirstLaunchAndExistingWallet(t *testing.T) {
	paths := storage.NewPaths(t.TempDir())
	cfg := config.Default()
	store := wallet.NewStore(paths.Wallet)
	model, err := NewModel(context.Background(), paths, cfg, store)
	if err != nil {
		t.Fatalf("NewModel failed: %v", err)
	}
	if model.screen != screenOnboarding {
		t.Fatalf("expected onboarding, got %v", model.screen)
	}
	if _, err := store.Import("test test test test test test test test test test test junk", "password", false); err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	model, err = NewModel(context.Background(), paths, cfg, store)
	if err != nil {
		t.Fatalf("NewModel failed: %v", err)
	}
	if model.screen != screenUnlock {
		t.Fatalf("expected unlock, got %v", model.screen)
	}
}

func TestDashboardMainnetSendDisabled(t *testing.T) {
	model := testUnlockedModel(t)
	model.cfg.ActiveNetworkID = networks.EthereumMainnet
	updated, cmd := model.dashboardSelect(0)
	if cmd != nil {
		t.Fatalf("expected no command")
	}
	if updated.screen != screenDashboard {
		t.Fatalf("expected dashboard, got %v", updated.screen)
	}
	if updated.errorText == "" {
		t.Fatalf("expected send-disabled error")
	}
}

func TestDashboardSendRoutesOnTestnet(t *testing.T) {
	model := testUnlockedModel(t)
	model.cfg.ActiveNetworkID = networks.EthereumSepolia
	updated, _ := model.dashboardSelect(0)
	if updated.screen != screenSendTo {
		t.Fatalf("expected send recipient screen, got %v", updated.screen)
	}
}

func TestNetworkSwitchUpdatesConfig(t *testing.T) {
	model := testUnlockedModel(t)
	model.screen = screenNetworks
	model.cursor = 1
	updatedModel, _ := model.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	updated := updatedModel.(Model)
	if updated.cfg.ActiveNetworkID != networks.SolanaDevnet {
		t.Fatalf("expected solana devnet, got %s", updated.cfg.ActiveNetworkID)
	}
	loaded, err := config.Load(updated.paths.Config)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.ActiveNetworkID != networks.SolanaDevnet {
		t.Fatalf("saved config not updated")
	}
}

func TestUserFacingErrorCompactsHTML(t *testing.T) {
	err := errors.New(`fetch ethereum balance: 404 Not Found: <!DOCTYPE HTML PUBLIC "-//IETF//DTD HTML 2.0//EN"><html><head><title>404 Not Found</title></head><body><h1>Not Found</h1></body></html>`)
	got := userFacingError(err)
	if strings.Contains(got, "<html>") || strings.Contains(got, "<!DOCTYPE") {
		t.Fatalf("expected html to be stripped, got %q", got)
	}
	if !strings.Contains(got, "RPC endpoint returned HTML instead of JSON-RPC") {
		t.Fatalf("unexpected message: %q", got)
	}
}

func TestUserFacingErrorTruncatesLongMessages(t *testing.T) {
	got := userFacingError(errors.New(strings.Repeat("x", 300)))
	if len(got) > 183 {
		t.Fatalf("expected truncated message, got length %d", len(got))
	}
}

func testUnlockedModel(t *testing.T) Model {
	t.Helper()
	base := t.TempDir()
	paths := storage.NewPaths(base)
	cfg := config.Default()
	cfgPath := filepath.Join(base, "config.json")
	paths.Config = cfgPath
	if err := config.Save(paths.Config, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	store := wallet.NewStore(paths.Wallet)
	plain, err := store.Import("test test test test test test test test test test test junk", "password", false)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	model, err := NewModel(context.Background(), paths, cfg, store)
	if err != nil {
		t.Fatalf("NewModel failed: %v", err)
	}
	model.screen = screenDashboard
	model.plain = &plain
	return model
}
