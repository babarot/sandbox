package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	listWidth             = 30
	listHeight            = 60
	maxDialogContentWidth = 24 // listWidth(30) - ボーダー(2) - パディング(2) - 余裕(2)
	minDialogContentWidth = 10 // 最小のダイアログ幅
)

type item struct {
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type screen int

const (
	listScreen screen = iota
	confirmScreen
)

type model struct {
	list   list.Model
	screen screen
	width  int
	height int
	styles styles
}

type styles struct {
	dialog lipgloss.Style
}

func initStyles() styles {
	s := styles{}
	s.dialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	return s
}

func isFileNameTooLong(filename string) bool {
	return len(filename) > maxDialogContentWidth
}

func getDialogWidth(content string) int {
	// 最長の行の長さを取得
	lines := strings.Split(content, "\n")
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	// 最小幅と最大幅の範囲内に収める
	width := maxLen + 4 // ボーダー(2) + パディング(2)
	if width < minDialogContentWidth {
		width = minDialogContentWidth
	}
	if width > listWidth {
		width = listWidth
	}
	return width
}

func (m model) Init() tea.Cmd {
	return getFiles
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch m.screen {
		case listScreen:
			switch msg.String() {
			case "d", "D":
				m.screen = confirmScreen
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		case confirmScreen:
			switch msg.String() {
			case "y", "Y":
				selected := m.list.SelectedItem().(item)
				// ファイル名が長すぎる場合は削除をスキップ
				if isFileNameTooLong(selected.title) {
					m.screen = listScreen
					return m, nil
				}
				err := os.Remove(selected.title)
				if err != nil {
					fmt.Printf("Error deleting file: %v\n", err)
				}
				m.screen = listScreen
				return m, getFiles
			case "n", "N", "esc":
				m.screen = listScreen
				return m, nil
			}
		}
	case []list.Item:
		m.list.SetItems(msg)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	baseView := m.list.View()

	switch m.screen {
	case confirmScreen:
		selected := m.list.SelectedItem().(item)
		if isFileNameTooLong(selected.title) {
			content := lipgloss.JoinVertical(lipgloss.Center,
				"File name too long",
				"for confirmation dialog",
				"",
				"(press esc)",
			)
			width := getDialogWidth(content)
			dialog := m.styles.dialog.
				Width(width).
				AlignHorizontal(lipgloss.Left).
				Render(content)

			lines := strings.Split(baseView, "\n")
			dialogLines := strings.Split(dialog, "\n")
			selectedIndex := m.list.Index()
			insertPosition := selectedIndex + 3
			if insertPosition+len(dialogLines) > len(lines) {
				insertPosition = len(lines) - len(dialogLines)
			}
			for i, dialogLine := range dialogLines {
				paddedLine := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(dialogLine)
				if insertPosition+i < len(lines) {
					lines[insertPosition+i] = paddedLine
				}
			}
			return strings.Join(lines, "\n")
		}

		content := lipgloss.JoinVertical(lipgloss.Center,
			"Delete "+selected.title+"?",
			"",
			"(y/n)",
		)
		width := getDialogWidth(content)
		dialog := m.styles.dialog.
			Width(width).
			Render(content)

		lines := strings.Split(baseView, "\n")
		dialogLines := strings.Split(dialog, "\n")
		selectedIndex := m.list.Index()
		insertPosition := selectedIndex + 2
		if insertPosition+len(dialogLines) > len(lines) {
			insertPosition = len(lines) - len(dialogLines)
		}
		for i, dialogLine := range dialogLines {
			paddedLine := lipgloss.NewStyle().Width(listWidth).Align(lipgloss.Left).PaddingLeft(2).Render(dialogLine)
			if insertPosition+i < len(lines) {
				lines[insertPosition+i] = paddedLine
			}
		}
		return strings.Join(lines, "\n")

	default:
		return baseView
	}
}

func getFiles() tea.Msg {
	var items []list.Item
	files, err := os.ReadDir(".")
	if err != nil {
		return items
	}

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		desc := fmt.Sprintf("Size: %d bytes", info.Size())
		if info.IsDir() {
			desc = "Directory"
		}
		items = append(items, item{
			title: file.Name(),
			desc:  desc,
		})
	}
	return items
}

func main() {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true

	l := list.New([]list.Item{}, delegate, listWidth, listHeight)
	l.Title = "File Manager"
	l.SetShowHelp(false)
	l.Styles.Title = l.Styles.Title.MarginLeft(2)
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "delete"),
			),
		}
	}

	m := model{
		list:   l,
		screen: listScreen,
		styles: initStyles(),
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
