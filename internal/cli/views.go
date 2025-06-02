package cli

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

// View renders the current view
func (m Model) View() string {
	if m.quitting {
		return "Thanks for using Okarun! 👋\n"
	}

	var content string
	if m.err != nil {
		content = HighlightStyle.Render(fmt.Sprintf("Error: %v", m.err))
	} else if m.loading {
		content = fmt.Sprintf("%s Loading...", m.spinner.View())
	} else if m.searchMode {
		content = fmt.Sprintf("Search Anime:\n\n%s\n\n(Enter to search, Esc to cancel)", m.textInput.View())
	} else {
		content = m.list.View()
	}

	return DocStyle.Render(content)
}

// InitializeList creates and configures a new list
func InitializeList(items []list.Item) list.Model {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(TitleStyle.GetForeground())
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(HighlightStyle.GetForeground())

	l := list.New(items, delegate, 30, 10)
	l.Title = "🌸 Okarun CLI"
	l.Styles.Title = TitleStyle
	l.SetShowHelp(true)
	l.KeyMap.Quit = key.NewBinding()

	return l
}

// GetMainMenuItems returns the default main menu items
func GetMainMenuItems() []list.Item {
	return []list.Item{
		NewMenuItem("Recent Updates", "See recently updated anime"),
		NewMenuItem("Search Anime", "Search for anime titles"),
		NewMenuItem("Exit", "Exit the application"),
	}
}
