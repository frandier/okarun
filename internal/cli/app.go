package cli

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// NewModel initializes a new model with default values
func NewModel() Model {
	mainMenuItems := GetMainMenuItems()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = LoadingStyle

	ti := textinput.New()
	ti.Placeholder = "Enter anime name..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return Model{
		list:          InitializeList(mainMenuItems),
		help:          help.New(),
		keys:          DefaultKeyMap(),
		spinner:       s,
		textInput:     ti,
		activeView:    "main",
		mainMenuItems: mainMenuItems,
	}
}

// Update handles all application updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle pagination in episodes view
		if m.activeView == "episodes" && !m.loading {
			switch {
			case key.Matches(msg, m.keys.Left):
				return m, NavigateToPrevPage(&m)
			case key.Matches(msg, m.keys.Right):
				return m, NavigateToNextPage(&m)
			}
		}

		// Handle back navigation first
		if key.Matches(msg, m.keys.Back) && m.activeView != "main" {
			return m, NavigateBack(&m)
		}

		// Check if we're in search mode
		if m.searchMode {
			switch msg.Type {
			case tea.KeyEnter:
				query := m.textInput.Value()
				if query != "" {
					m.loading = true
					m.searchMode = false
					return m, SearchAnime(query)
				}
			case tea.KeyEsc:
				m.searchMode = false
				return m, NavigateToMain(&m)
			default:
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
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
					case "Search Anime":
						return m, NavigateToSearch(&m)
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
				case "search":
					if !m.loading {
						idx := m.list.Index()
						if idx < len(m.searchResults) {
							selectedAnime := m.searchResults[idx]
							return m, NavigateToEpisodes(&m, selectedAnime)
						}
					}
				case "episodes":
					if !m.loading {
						idx := m.list.Index()
						items := m.list.Items()
						if idx < len(items) {
							episode := m.currentEpisodes.Episodes[idx]
							return m, NavigateToServerSelect(&m, &episode)
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

	case SearchAnimeMsg:
		return m, UpdateSearchResults(&m, msg)

	case FetchEpisodesMsg:
		return m, UpdateEpisodesList(&m, msg)

	case FetchNextPageMsg:
		return m, UpdatePageList(&m, msg)

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
