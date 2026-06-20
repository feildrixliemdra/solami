package tui

import (
	"fmt"
	"strings"

	"github.com/feildrix/solami/internal/networks"
)

const asciiLogo = "   _____       __                 _\n" +
	"  / ___/____  / /___ _____ ___  _(_)\n" +
	"  \\__ \\/ __ \\/ / __ `/ __ `__ \\/ / /\n" +
	" ___/ / /_/ / / /_/ / / / / / / / /\n" +
	"/____/\\____/_/\\__,_/_/ /_/ /_/_/_/"

// View renders the TUI.
func (m Model) View() string {
	var body string
	switch m.screen {
	case screenOnboarding:
		body = titleStyle.Render(asciiLogo) + "\n\n" + m.renderMenu("Welcome to Solami", []string{"Create New Wallet", "Import Existing Wallet", "Exit"})
	case screenCreateShowMnemonic:
		body = m.renderCreateMnemonic()
	case screenUnlock:
		body = titleStyle.Render(asciiLogo) + "\n\n" + m.renderInput()
	case screenCreatePassword, screenImportPassword, screenImportMnemonic, screenSendTo, screenSendAmount, screenSendPassword:
		body = m.renderInput()
	case screenDashboard:
		body = m.renderDashboard()
	case screenSendConfirm:
		body = m.renderSendConfirm()
	case screenReceive:
		body = m.renderReceive()
	case screenNetworks:
		items := make([]string, 0, len(networks.Defaults()))
		for _, network := range networks.Defaults() {
			label := network.Name
			if network.ID == m.cfg.ActiveNetworkID {
				label += " (active)"
			}
			if !network.SendEnabled {
				label += " - send disabled"
			}
			items = append(items, label)
		}
		body = m.renderMenu("Switch Network", items)
	case screenSettings:
		body = m.renderMenu("Settings", []string{"Export recovery phrase", "Export private key", "Back"})
	default:
		body = "Unknown screen"
	}
	footerText := "esc back  |  ctrl+c quit"
	if m.screen == screenDashboard {
		footerText = "esc back  |  r refresh  |  ctrl+c quit"
	}
	footer := "\n\n" + mutedStyle.Render(footerText)
	if m.errorText != "" {
		footer += "\n" + errorStyle.Render(m.errorText)
	}
	if m.statusText != "" {
		footer += "\n" + statusStyle.Render(m.statusText)
	}
	if m.loading {
		footer += "\n" + mutedStyle.Render("Working...")
	}
	return boxStyle.Render(body + footer)
}

func (m Model) renderMenu(title string, items []string) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")
	for i, item := range items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		b.WriteString(cursor)
		b.WriteString(item)
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) renderCreateMnemonic() string {
	return titleStyle.Render("Recovery Phrase") + "\n\n" +
		"Write this phrase down and store it offline. It will not be shown again during normal use.\n\n" +
		m.mnemonic + "\n\n" +
		mutedStyle.Render("Press enter after you have saved it.")
}

func (m Model) renderInput() string {
	title := map[screen]string{
		screenCreatePassword: "Create Wallet Password",
		screenImportMnemonic: "Import Existing Mnemonic",
		screenImportPassword: "Create Wallet Password",
		screenUnlock:         "Unlock Wallet",
		screenSendTo:         "Recipient Address",
		screenSendAmount:     "Amount",
		screenSendPassword:   "Confirm Password",
	}[m.screen]
	if title == "" {
		title = "Input"
	}
	return titleStyle.Render(title) + "\n\n" + m.input.View()
}

func (m Model) renderDashboard() string {
	network := m.cfg.ActiveNetwork()
	address := ""
	if m.plain != nil {
		if network.Chain == networks.ChainEthereum {
			address = m.plain.Account.EthereumAddress
		} else {
			address = m.plain.Account.SolanaAddress
		}
	}
	sendStatus := "enabled"
	if !network.SendEnabled {
		sendStatus = "disabled in v1"
	}
	header := fmt.Sprintf(
		"%s\n\nActive Network: %s\nAddress: %s\nBalance: %s\nSending: %s\n\n",
		titleStyle.Render(asciiLogo),
		network.Name,
		shortAddress(address),
		defaultText(m.balance, "loading..."),
		sendStatus,
	)
	menu := m.renderMenu("", []string{"Send", "Receive", "Switch Network", "Settings", "Lock Wallet", "Exit"})
	return header + strings.TrimSpace(menu)
}

func (m Model) renderReceive() string {
	network := m.cfg.ActiveNetwork()
	address := ""
	if m.plain != nil {
		if network.Chain == networks.ChainEthereum {
			address = m.plain.Account.EthereumAddress
		} else {
			address = m.plain.Account.SolanaAddress
		}
	}
	return titleStyle.Render("Receive") + "\n\n" +
		"Network: " + network.Name + "\n\n" +
		address + "\n\n" +
		"Only send " + network.Symbol + " on " + network.Name + " to this address.\n\n" +
		mutedStyle.Render("Press enter to return.")
}

func (m Model) renderSendConfirm() string {
	network := m.cfg.ActiveNetwork()
	items := []string{"Confirm", "Cancel"}
	return titleStyle.Render("Confirm Transaction") + "\n\n" +
		"Network: " + network.Name + "\n" +
		"To: " + m.sendTo + "\n" +
		"Amount: " + m.sendAmount + " " + network.Symbol + "\n" +
		"Estimated Fee: " + m.sendFee.Formatted + " " + m.sendFee.Symbol + "\n\n" +
		strings.TrimSpace(m.renderMenu("", items))
}

func shortAddress(address string) string {
	if len(address) <= 16 {
		return address
	}
	return address[:8] + "..." + address[len(address)-6:]
}

func defaultText(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
