package cli

import (
	"fmt"
	"os/exec"
	"yokai/internal/anime"

	"github.com/charmbracelet/bubbles/list"
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
	case "servers":
		m.activeView = "recent"
		m.loading = false
		m.err = nil
		// Restore the episodes list
		items := make([]list.Item, len(m.animes))
		for i, ep := range m.animes {
			items[i] = NewMenuItem(
				ep.Title,
				fmt.Sprintf("Episode %s", ep.Episode),
			)
		}
		m.list.SetItems(items)
		m.list.Title = "🌸 Recent Updates (Press ESC to go back)"
		m.list.Select(0) // Reset cursor to first item
		return nil
	default:
		return NavigateToMain(m)
	}
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
