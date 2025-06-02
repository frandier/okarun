package cli

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// NewModel initializes a new model with default values
func NewModel() Model {
	mainMenuItems := GetMainMenuItems()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = LoadingStyle

	return Model{
		list:          InitializeList(mainMenuItems),
		help:          help.New(),
		keys:          DefaultKeyMap(),
		spinner:       s,
		activeView:    "main",
		mainMenuItems: mainMenuItems,
	}
}

// Update handles all application updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle back navigation first
		if key.Matches(msg, m.keys.Back) && m.activeView != "main" {
			return m, NavigateBack(&m)
		}

		// Then check for quit
		if key.Matches(msg, m.keys.Quit) {
			m.quitting = true
			return m, tea.Quit
		}

		if key.Matches(msg, m.keys.Help) {
			return m, nil
		}

		// Handle menu selections
		if key.Matches(msg, m.keys.Enter) {
			if m.list.SelectedItem() != nil {
				switch m.activeView {
				case "main":
					switch m.list.SelectedItem().(MenuItem).title {
					case "Recent Updates":
						if !m.loading {
							return m, NavigateToRecent(&m)
						}
					case "Exit":
						m.quitting = true
						return m, tea.Quit
					}
				case "recent":
					if !m.loading {
						// Find the selected episode
						idx := m.list.Index()
						if idx < len(m.animes) {
							return m, NavigateToServerSelect(&m, &m.animes[idx])
						}
					}
				case "servers":
					if !m.loading && m.selectedEpisode != nil {
						idx := m.list.Index()
						if idx < len(m.servers) {
							server := m.servers[idx]
							m.loading = true
							return m, PlayEpisode(server)
						}
					}
				}
			}
		}

	case FetchLatestMsg:
		return m, UpdateRecentList(&m, msg)

	case FetchServersMsg:
		return m, UpdateServerList(&m, msg)

	case PlayEpisodeMsg:
		return m, HandlePlayback(&m, msg)

	case tea.WindowSizeMsg:
		h, v := DocStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick)
}
