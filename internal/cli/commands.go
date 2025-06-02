package cli

import (
	"fmt"
	"os/exec"
	"yokai/internal/anime"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// FetchLatestMsg represents a message containing fetched episodes
type FetchLatestMsg struct {
	Episodes []anime.LatestEpisode
	Err      error
}

// FetchLatest fetches the latest episodes
func FetchLatest() tea.Msg {
	client := &anime.Jkanime{}
	episodes, err := client.GetLatestEpisodes()
	return FetchLatestMsg{Episodes: episodes, Err: err}
}

// NavigateToMain returns the model to the main menu state
func NavigateToMain(m *Model) tea.Cmd {
	m.activeView = "main"
	m.loading = false
	m.err = nil
	m.list.SetItems(m.mainMenuItems)
	m.list.Title = "🌸 Okarun CLI"
	m.list.Select(0) // Reset cursor to first item
	return nil
}

// NavigateToRecent prepares the model for the recent updates view
func NavigateToRecent(m *Model) tea.Cmd {
	m.loading = true
	m.previousView = "main"
	m.activeView = "recent"
	m.list.Title = "🌸 Recent Updates (Press ESC to go back)"
	return tea.Batch(
		m.spinner.Tick,
		FetchLatest,
	)
}

// UpdateRecentList updates the list with recent episodes
func UpdateRecentList(m *Model, msg FetchLatestMsg) tea.Cmd {
	m.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return nil
	}

	m.animes = msg.Episodes
	items := make([]list.Item, len(msg.Episodes))
	for i, ep := range msg.Episodes {
		items[i] = NewMenuItem(
			ep.Title,
			"Episode "+ep.Episode,
		)
	}
	m.list.SetItems(items)
	m.list.Select(0) // Reset cursor to first item
	return nil
}

// FetchServersMsg represents a message containing fetched servers
type FetchServersMsg struct {
	Servers []anime.Server
	Episode *anime.LatestEpisode
	Err     error
}

// FetchServers fetches the available servers for an episode
func FetchServers(ep anime.LatestEpisode) tea.Cmd {
	return func() tea.Msg {
		client := &anime.Jkanime{}
		servers, err := client.GetServers(ep.Slug, ep.Episode)
		return FetchServersMsg{
			Servers: servers,
			Episode: &ep,
			Err:     err,
		}
	}
}

// NavigateToServerSelect prepares the model for the server selection view
func NavigateToServerSelect(m *Model, ep *anime.LatestEpisode) tea.Cmd {
	m.loading = true
	m.previousView = m.activeView // Store the previous view before changing
	m.activeView = "servers"
	m.selectedEpisode = ep
	m.list.Title = fmt.Sprintf("🌸 %s - Episode %s (Press ESC to go back)", ep.Title, ep.Episode)
	return tea.Batch(
		m.spinner.Tick,
		FetchServers(*ep),
	)
}

// UpdateServerList updates the list with available servers
func UpdateServerList(m *Model, msg FetchServersMsg) tea.Cmd {
	m.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return nil
	}

	m.servers = msg.Servers
	items := make([]list.Item, len(msg.Servers))
	for i, server := range msg.Servers {
		items[i] = NewMenuItem(
			fmt.Sprintf("Server %d", i+1),
			server.Server,
		)
	}
	m.list.SetItems(items)
	m.list.Select(0) // Reset cursor to first item
	return nil
}

// NavigateToServers prepares the model for the servers view
func NavigateToServers(m *Model, servers []anime.Server) tea.Cmd {
	m.activeView = "servers"
	m.servers = servers

	// Convert servers to list items
	items := make([]list.Item, len(servers))
	for i, server := range servers {
		items[i] = NewMenuItem(server.Server, "Select to play episode")
	}

	m.list.Title = "🌸 Select Server (Press ESC to go back)"
	m.list.SetItems(items)
	return nil
}

// NavigateBack handles going back to the previous view
func NavigateBack(m *Model) tea.Cmd {
	switch m.activeView {
	case "episodes":
		return NavigateBackToSearch(m)
	case "servers":
		if m.previousView == "recent" {
			m.activeView = "recent"
			items := make([]list.Item, len(m.animes))
			for i, anime := range m.animes {
				items[i] = NewMenuItem(anime.Title, fmt.Sprintf("Episode %s", anime.Episode))
			}
			m.list.SetItems(items)
			m.list.Title = "🌸 Recent Updates (Press ESC to go back)"
			return nil
		} else {
			return RestoreEpisodeList(m)
		}
	default:
		return NavigateToMain(m)
	}
}

// NavigateBackToSearch returns to the search results view
func NavigateBackToSearch(m *Model) tea.Cmd {
	m.activeView = "search"
	m.loading = false
	m.err = nil

	// Restore the search results list
	items := make([]list.Item, len(m.searchResults))
	for i, anime := range m.searchResults {
		items[i] = NewMenuItem(anime.Title, anime.Synopsis)
	}
	m.list.SetItems(items)
	m.list.Title = "🌸 Search Results (Press ESC to go back)"
	return nil
}

// PlayEpisodeMsg represents a message containing the streaming URL
type PlayEpisodeMsg struct {
	StreamingURL string
	Err          error
}

// PlayEpisode starts playback of the selected episode
func PlayEpisode(server anime.Server) tea.Cmd {
	return func() tea.Msg {
		client := &anime.Jkanime{}
		streamingURL, err := client.GetStreaming(server.Server, server.Remote)
		return PlayEpisodeMsg{StreamingURL: streamingURL, Err: err}
	}
}

// HandlePlayback handles the MPV playback of the streaming URL
func HandlePlayback(m *Model, msg PlayEpisodeMsg) tea.Cmd {
	if msg.Err != nil {
		m.err = msg.Err
		return nil
	}

	// Run MPV in a separate goroutine to not block the UI
	go func() {
		exec.Command("mpv", msg.StreamingURL).Run()
	}()

	// Just reset the loading state and stay in the current view
	m.loading = false
	return nil
}

// SearchAnimeMsg represents the result of an anime search
type SearchAnimeMsg struct {
	Results []anime.Anime
	Err     error
}

// NavigateToSearch prepares the model for the search view
func NavigateToSearch(m *Model) tea.Cmd {
	m.activeView = "search"
	m.searchMode = true
	m.textInput.Focus()
	m.textInput.SetValue("")
	m.list.Title = "🌸 Search Anime (Press ESC to go back)"
	return textinput.Blink
}

// SearchAnime performs the anime search
func SearchAnime(query string) tea.Cmd {
	return func() tea.Msg {
		client := &anime.Jkanime{}
		results, err := client.GetSearch(query, 1) // Start with page 1
		return SearchAnimeMsg{Results: results, Err: err}
	}
}

// UpdateSearchResults updates the list with search results
func UpdateSearchResults(m *Model, msg SearchAnimeMsg) tea.Cmd {
	m.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return nil
	}

	m.searchResults = msg.Results
	items := make([]list.Item, len(msg.Results))
	for i, anime := range msg.Results {
		items[i] = NewMenuItem(anime.Title, anime.Synopsis)
	}
	m.list.SetItems(items)
	m.list.Select(0) // Reset cursor to first item
	m.searchMode = false
	return nil
}

// FetchEpisodesMsg represents a message containing fetched episodes for an anime
type FetchEpisodesMsg struct {
	Episodes   *anime.Episode
	AnimeTitle string
	Err        error
}

// FetchEpisodes fetches episodes for a specific anime
func FetchEpisodes(slug string, title string) tea.Cmd {
	return func() tea.Msg {
		client := &anime.Jkanime{}
		episodes, err := client.GetEpisodes(slug, 1) // Start with page 1
		return FetchEpisodesMsg{
			Episodes:   episodes,
			AnimeTitle: title,
			Err:        err,
		}
	}
}

// NavigateToEpisodes prepares the model for the episodes view
func NavigateToEpisodes(m *Model, searchedAnime anime.Anime) tea.Cmd {
	m.loading = true
	m.previousView = m.activeView // Store the previous view before changing
	m.activeView = "episodes"
	m.currentAnime = &searchedAnime
	m.currentPage = 1
	m.list.Title = fmt.Sprintf("🌸 %s - Episodes (Press ESC to go back)", searchedAnime.Title)
	return tea.Batch(
		m.spinner.Tick,
		FetchEpisodes(searchedAnime.Slug, searchedAnime.Title),
	)
}

// UpdateEpisodesList updates the list with episodes
func UpdateEpisodesList(m *Model, msg FetchEpisodesMsg) tea.Cmd {
	m.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return nil
	}

	m.currentEpisodes = msg.Episodes
	m.totalPages = msg.Episodes.TotalPages
	episodes := msg.Episodes.Episodes
	items := make([]list.Item, len(episodes))
	for i, episode := range episodes {
		items[i] = NewMenuItem(
			fmt.Sprintf("Episode %s", episode.Episode),
			fmt.Sprintf("Page %d/%d - Watch episode %s of %s",
				m.currentPage, m.totalPages,
				episode.Episode, msg.AnimeTitle),
		)
	}
	m.list.SetItems(items)
	m.list.Title = fmt.Sprintf("🌸 %s - Page %d/%d (← → to navigate, ESC to go back)",
		msg.AnimeTitle, m.currentPage, m.totalPages)
	m.list.Select(0) // Reset cursor to first item
	return nil
}

// FetchNextPageMsg represents a message containing the next page of episodes
type FetchNextPageMsg struct {
	Episodes   *anime.Episode
	AnimeTitle string
	Page       int
	Err        error
}

// NavigateToNextPage fetches the next page of episodes
func NavigateToNextPage(m *Model) tea.Cmd {
	if m.currentPage >= m.totalPages {
		return nil
	}

	m.loading = true
	nextPage := m.currentPage + 1

	return func() tea.Msg {
		client := &anime.Jkanime{}
		episodes, err := client.GetEpisodes(m.currentAnime.Slug, nextPage)
		return FetchNextPageMsg{
			Episodes:   episodes,
			AnimeTitle: m.currentAnime.Title,
			Page:       nextPage,
			Err:        err,
		}
	}
}

// NavigateToPrevPage fetches the previous page of episodes
func NavigateToPrevPage(m *Model) tea.Cmd {
	if m.currentPage <= 1 {
		return nil
	}

	m.loading = true
	prevPage := m.currentPage - 1

	return func() tea.Msg {
		client := &anime.Jkanime{}
		episodes, err := client.GetEpisodes(m.currentAnime.Slug, prevPage)
		return FetchNextPageMsg{
			Episodes:   episodes,
			AnimeTitle: m.currentAnime.Title,
			Page:       prevPage,
			Err:        err,
		}
	}
}

// UpdatePageList updates the list with the current page of episodes
func UpdatePageList(m *Model, msg FetchNextPageMsg) tea.Cmd {
	m.loading = false
	if msg.Err != nil {
		m.err = msg.Err
		return nil
	}

	m.currentPage = msg.Page
	m.currentEpisodes = msg.Episodes
	episodes := msg.Episodes.Episodes
	items := make([]list.Item, len(episodes))
	for i, episode := range episodes {
		items[i] = NewMenuItem(
			fmt.Sprintf("Episode %s", episode.Episode),
			fmt.Sprintf("Page %d/%d - Watch episode %s of %s",
				m.currentPage, m.totalPages,
				episode.Episode, msg.AnimeTitle),
		)
	}
	m.list.SetItems(items)
	m.list.Title = fmt.Sprintf("🌸 %s - Page %d/%d (← → to navigate, ESC to go back)",
		msg.AnimeTitle, m.currentPage, m.totalPages)
	m.list.Select(0)
	return nil
}

// RestoreEpisodeList recreates the episode list UI when returning from servers view
func RestoreEpisodeList(m *Model) tea.Cmd {
	m.activeView = "episodes"
	if m.currentEpisodes == nil || len(m.currentEpisodes.Episodes) == 0 {
		m.err = fmt.Errorf("no episodes loaded")
		return nil
	}

	// Recreate the episode list items
	episodes := m.currentEpisodes.Episodes
	items := make([]list.Item, len(episodes))
	for i, episode := range episodes {
		items[i] = NewMenuItem(
			fmt.Sprintf("Episode %s", episode.Episode),
			fmt.Sprintf("Page %d/%d - Watch episode %s of %s",
				m.currentPage, m.totalPages,
				episode.Episode, m.currentAnime.Title),
		)
	}
	m.list.SetItems(items)
	m.list.Title = fmt.Sprintf("🌸 %s - Page %d/%d (← → to navigate, ESC to go back)",
		m.currentAnime.Title, m.currentPage, m.totalPages)

	return nil
}
