package tui

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/feildrix/solami/internal/chains"
	"github.com/feildrix/solami/internal/config"
	"github.com/feildrix/solami/internal/networks"
	"github.com/feildrix/solami/internal/storage"
	"github.com/feildrix/solami/internal/wallet"
)

type screen int

const (
	screenOnboarding screen = iota
	screenCreateShowMnemonic
	screenCreatePassword
	screenImportMnemonic
	screenImportPassword
	screenUnlock
	screenDashboard
	screenSendTo
	screenSendAmount
	screenSendConfirm
	screenSendPassword
	screenReceive
	screenNetworks
	screenSettings
)

type balanceMsg struct {
	balance chains.Balance
	err     error
}

type feeMsg struct {
	fee chains.FeeEstimate
	err error
}

type sendMsg struct {
	result chains.SendResult
	err    error
}

type tickMsg time.Time

// Model is the Bubble Tea application model.
type Model struct {
	ctx              context.Context
	paths            storage.Paths
	cfg              config.Config
	store            wallet.Store
	hasWallet        bool
	plain            *wallet.PlainWallet
	screen           screen
	cursor           int
	input            textinput.Model
	mnemonic         string
	errorText        string
	statusText       string
	balance          string
	loading          bool
	sendTo           string
	sendAmount       string
	sendFee          chains.FeeEstimate
	updatedAt        time.Time
	lastBalanceFetch time.Time
}

// NewModel creates a TUI model.
func NewModel(ctx context.Context, paths storage.Paths, cfg config.Config, store wallet.Store) (Model, error) {
	exists, err := store.Exists()
	if err != nil {
		return Model{}, err
	}
	input := textinput.New()
	input.CharLimit = 256
	input.Width = 64
	input.Prompt = "> "
	initial := screenOnboarding
	if exists {
		initial = screenUnlock
		input.Placeholder = "Password"
		input.EchoMode = textinput.EchoPassword
		input.Focus()
	}
	return Model{
		ctx:       ctx,
		paths:     paths,
		cfg:       cfg,
		store:     store,
		hasWallet: exists,
		screen:    initial,
		input:     input,
		updatedAt: time.Now(),
	}, nil
}

// Init initializes the TUI.
func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Update handles TUI events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		var cmd tea.Cmd
		now := time.Time(msg)
		if m.shouldAutoLock(now) {
			m.lock()
		} else if m.plain != nil && m.screen == screenDashboard && !m.loading && now.Sub(m.lastBalanceFetch) >= 15*time.Second {
			m.lastBalanceFetch = now
			cmd = m.fetchBalanceCmd()
		}
		return m, tea.Batch(tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) }), cmd)
	case balanceMsg:
		m.loading = false
		m.lastBalanceFetch = time.Now()
		if msg.err != nil {
			m.balance = "unavailable"
			m.errorText = userFacingError(msg.err)
			return m, nil
		}
		m.balance = fmt.Sprintf("%s %s", msg.balance.Formatted, msg.balance.Symbol)
		m.errorText = ""
		return m, nil
	case feeMsg:
		m.loading = false
		if msg.err != nil {
			m.errorText = userFacingError(msg.err)
			m.screen = screenSendAmount
			m.prepareInput("Amount", false)
			m.input.SetValue(m.sendAmount)
			return m, nil
		}
		m.sendFee = msg.fee
		m.screen = screenSendConfirm
		return m, nil
	case sendMsg:
		m.loading = false
		if msg.err != nil {
			m.errorText = userFacingError(msg.err)
			m.screen = screenDashboard
			return m, m.fetchBalanceCmd()
		}
		m.statusText = "Transaction sent: " + msg.result.Hash
		if msg.result.ExplorerURL != "" {
			m.statusText += "\n" + msg.result.ExplorerURL
		}
		m.errorText = ""
		m.screen = screenDashboard
		return m, m.fetchBalanceCmd()
	case tea.KeyMsg:
		m.updatedAt = time.Now()
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m.handleBack()
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenOnboarding:
		return m.updateMenu(msg, []string{"Create New Wallet", "Import Existing Wallet", "Exit"}, func(index int) (Model, tea.Cmd) {
			switch index {
			case 0:
				mnemonic, err := wallet.GenerateMnemonic()
				if err != nil {
					m.errorText = err.Error()
					return m, nil
				}
				m.mnemonic = mnemonic
				m.screen = screenCreateShowMnemonic
				m.errorText = ""
				return m, nil
			case 1:
				m.screen = screenImportMnemonic
				m.prepareInput("Mnemonic", false)
				return m, nil
			default:
				return m, tea.Quit
			}
		})
	case screenCreateShowMnemonic:
		if msg.String() == "enter" {
			m.screen = screenCreatePassword
			m.prepareInput("Create password", true)
		}
		return m, nil
	case screenCreatePassword:
		return m.updateTextInput(msg, func(value string) (Model, tea.Cmd) {
			plain, err := m.store.Import(m.mnemonic, value, false)
			if err != nil {
				m.errorText = err.Error()
				return m, nil
			}
			m.plain = &plain
			m.hasWallet = true
			m.screen = screenDashboard
			m.statusText = "Wallet created"
			m.input.Blur()
			return m, m.fetchBalanceCmd()
		})
	case screenImportMnemonic:
		return m.updateTextInput(msg, func(value string) (Model, tea.Cmd) {
			if err := wallet.ValidateMnemonic(strings.TrimSpace(value)); err != nil {
				m.errorText = err.Error()
				return m, nil
			}
			m.mnemonic = strings.TrimSpace(value)
			m.screen = screenImportPassword
			m.prepareInput("Create password", true)
			return m, nil
		})
	case screenImportPassword:
		return m.updateTextInput(msg, func(value string) (Model, tea.Cmd) {
			plain, err := m.store.Import(m.mnemonic, value, false)
			if err != nil {
				m.errorText = err.Error()
				return m, nil
			}
			m.plain = &plain
			m.hasWallet = true
			m.screen = screenDashboard
			m.statusText = "Wallet imported"
			m.input.Blur()
			return m, m.fetchBalanceCmd()
		})
	case screenUnlock:
		return m.updateTextInput(msg, func(value string) (Model, tea.Cmd) {
			plain, err := m.store.Unlock(value)
			if err != nil {
				if errors.Is(err, wallet.ErrInvalidPassword) {
					m.errorText = "Invalid password"
				} else {
					m.errorText = err.Error()
				}
				return m, nil
			}
			m.plain = &plain
			m.screen = screenDashboard
			m.input.Blur()
			m.statusText = "Wallet unlocked"
			return m, m.fetchBalanceCmd()
		})
	case screenDashboard:
		if msg.String() == "r" || msg.String() == "R" {
			m.loading = true
			m.errorText = ""
			m.lastBalanceFetch = time.Now()
			return m, m.fetchBalanceCmd()
		}
		return m.updateMenu(msg, []string{"Send", "Receive", "Switch Network", "Settings", "Lock Wallet", "Exit"}, m.dashboardSelect)
	case screenSendTo:
		return m.updateTextInput(msg, func(value string) (Model, tea.Cmd) {
			m.sendTo = strings.TrimSpace(value)
			m.screen = screenSendAmount
			m.prepareInput("Amount", false)
			return m, nil
		})
	case screenSendAmount:
		return m.updateTextInput(msg, func(value string) (Model, tea.Cmd) {
			m.sendAmount = strings.TrimSpace(value)
			m.loading = true
			m.errorText = ""
			return m, m.estimateFeeCmd()
		})
	case screenSendConfirm:
		return m.updateMenu(msg, []string{"Confirm", "Cancel"}, func(index int) (Model, tea.Cmd) {
			if index == 0 {
				m.screen = screenSendPassword
				m.prepareInput("Password", true)
				return m, nil
			}
			m.screen = screenDashboard
			m.statusText = "Transaction cancelled"
			return m, nil
		})
	case screenSendPassword:
		return m.updateTextInput(msg, func(value string) (Model, tea.Cmd) {
			plain, err := m.store.Unlock(value)
			if err != nil {
				m.errorText = "Password confirmation failed"
				return m, nil
			}
			m.loading = true
			m.input.Blur()
			return m, m.sendCmd(plain)
		})
	case screenReceive:
		if msg.String() == "enter" {
			m.screen = screenDashboard
		}
		return m, nil
	case screenNetworks:
		items := make([]string, 0, len(networks.Defaults()))
		for _, network := range networks.Defaults() {
			label := network.Name
			if !network.SendEnabled {
				label += " (send disabled)"
			}
			items = append(items, label)
		}
		return m.updateMenu(msg, items, func(index int) (Model, tea.Cmd) {
			network := networks.Defaults()[index]
			m.cfg.ActiveNetworkID = network.ID
			if err := config.Save(m.paths.Config, m.cfg); err != nil {
				m.errorText = err.Error()
			} else {
				m.statusText = "Switched to " + network.Name
				m.errorText = ""
			}
			m.screen = screenDashboard
			m.cursor = 0
			return m, m.fetchBalanceCmd()
		})
	case screenSettings:
		return m.updateMenu(msg, []string{"Export recovery phrase", "Export private key", "Back"}, func(index int) (Model, tea.Cmd) {
			switch index {
			case 0, 1:
				m.statusText = "Sensitive export requires password confirmation and is not enabled in this v1 build."
			}
			m.screen = screenDashboard
			return m, nil
		})
	}
	return m, nil
}

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func userFacingError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if strings.Contains(message, "<html") || strings.Contains(message, "<!DOCTYPE") {
		message = htmlTagPattern.ReplaceAllString(message, " ")
		message = strings.Join(strings.Fields(message), " ")
		if len(message) > 140 {
			message = message[:140] + "..."
		}
		return "RPC endpoint returned HTML instead of JSON-RPC: " + message
	}
	if len(message) > 180 {
		message = message[:180] + "..."
	}
	return message
}

func (m Model) updateMenu(msg tea.KeyMsg, items []string, selectFn func(int) (Model, tea.Cmd)) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
		}
	case "enter":
		return selectFn(m.cursor)
	}
	return m, nil
}

func (m Model) updateTextInput(msg tea.KeyMsg, submit func(string) (Model, tea.Cmd)) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		value := strings.TrimSpace(m.input.Value())
		if value == "" {
			m.errorText = "Value cannot be empty"
			return m, nil
		}
		return submit(value)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) dashboardSelect(index int) (Model, tea.Cmd) {
	network := m.cfg.ActiveNetwork()
	switch index {
	case 0:
		if !network.SendEnabled {
			m.errorText = "Sending is disabled for " + network.Name + " in v1"
			return m, nil
		}
		m.screen = screenSendTo
		m.prepareInput("Recipient address", false)
		m.sendTo = ""
		m.sendAmount = ""
		return m, nil
	case 1:
		m.screen = screenReceive
		return m, nil
	case 2:
		m.screen = screenNetworks
		m.cursor = 0
		return m, nil
	case 3:
		m.screen = screenSettings
		m.cursor = 0
		return m, nil
	case 4:
		m.lock()
		return m, nil
	default:
		return m, tea.Quit
	}
}

func (m Model) handleBack() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenDashboard:
		return m, nil
	case screenUnlock, screenOnboarding:
		return m, tea.Quit
	default:
		m.screen = screenDashboard
		m.input.Blur()
		m.errorText = ""
		return m, nil
	}
}

func (m *Model) prepareInput(placeholder string, secret bool) {
	m.input = textinput.New()
	m.input.Placeholder = placeholder
	m.input.CharLimit = 512
	m.input.Width = 72
	m.input.Prompt = "> "
	if secret {
		m.input.EchoMode = textinput.EchoPassword
	}
	m.input.Focus()
	m.errorText = ""
}

func (m *Model) lock() {
	m.plain = nil
	m.screen = screenUnlock
	m.prepareInput("Password", true)
	m.statusText = "Wallet locked"
	m.balance = ""
	m.cursor = 0
}

func (m Model) shouldAutoLock(now time.Time) bool {
	if m.plain == nil || m.cfg.AutoLockMinutes <= 0 {
		return false
	}
	return now.Sub(m.updatedAt) >= time.Duration(m.cfg.AutoLockMinutes)*time.Minute
}

func (m Model) fetchBalanceCmd() tea.Cmd {
	if m.plain == nil {
		return nil
	}
	network := m.cfg.ActiveNetwork()
	account := m.plain.Account
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 15*time.Second)
		defer cancel()
		client, err := chains.NewClient(network)
		if err != nil {
			return balanceMsg{err: err}
		}
		balance, err := client.Balance(ctx, account)
		return balanceMsg{balance: balance, err: err}
	}
}

func (m Model) estimateFeeCmd() tea.Cmd {
	network := m.cfg.ActiveNetwork()
	account := m.plain.Account
	to := m.sendTo
	amount := m.sendAmount
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 15*time.Second)
		defer cancel()
		client, err := chains.NewClient(network)
		if err != nil {
			return feeMsg{err: err}
		}
		fee, err := client.EstimateFee(ctx, account, to, amount)
		return feeMsg{fee: fee, err: err}
	}
}

func (m Model) sendCmd(plain wallet.PlainWallet) tea.Cmd {
	network := m.cfg.ActiveNetwork()
	to := m.sendTo
	amount := m.sendAmount
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 45*time.Second)
		defer cancel()
		client, err := chains.NewClient(network)
		if err != nil {
			return sendMsg{err: err}
		}
		result, err := client.Send(ctx, plain, to, amount)
		return sendMsg{result: result, err: err}
	}
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	boxStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(78)
	mutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)
