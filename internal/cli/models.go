package cli

import (
	"yokai/internal/anime"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
)

// Model represents the application state
type Model struct {
	list            list.Model
	help            help.Model
	keys            KeyMap
	spinner         spinner.Model
	textInput       textinput.Model
	loading         bool
	quitting        bool
	animes          []anime.LatestEpisode
	servers         []anime.Server
	selectedEpisode *anime.LatestEpisode
	searchResults   []anime.Anime
	currentPage     int
	totalPages      int
	currentAnime    *anime.Anime
	currentEpisodes *anime.Episode
	err             error
	activeView      string
	previousView    string
	mainMenuItems   []list.Item
	searchMode      bool
}

// MenuItem represents an item in any menu list
type MenuItem struct {
	title       string
	description string
}

func (i MenuItem) Title() string       { return i.title }
func (i MenuItem) Description() string { return i.description }
func (i MenuItem) FilterValue() string { return i.title }

// NewMenuItem creates a new menu item
func NewMenuItem(title, description string) MenuItem {
	return MenuItem{
		title:       title,
		description: description,
	}
}
