package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

func (h *Handler) GetLatestEpisodes(w http.ResponseWriter, r *http.Request) {
	latestEpisodes, err := h.scrapper.GetLatestEpisodes()
	if err != nil {
		logrus.Errorf("Error getting latest episodes: %v", err.Error())
		http.Error(w, "Error getting latest episodes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latestEpisodes)
}

func (h *Handler) GetAnime(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Slug is required", http.StatusBadRequest)
		return
	}

	animeDetails, err := h.scrapper.GetAnime(slug)
	if err != nil {
		logrus.Errorf("Error getting anime details: %v", err.Error())
		http.Error(w, "Error getting anime details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(animeDetails)
}

func (h *Handler) GetEpisodes(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Slug is required", http.StatusBadRequest)
		return
	}

	page := r.URL.Query().Get("page")
	if page == "" {
		http.Error(w, "Page is required", http.StatusBadRequest)
		return
	}
	pageNum, err := strconv.Atoi(page)

	if err != nil {
		http.Error(w, "Page must be a number", http.StatusBadRequest)
		return
	}

	episodes, err := h.scrapper.GetEpisodes(slug, pageNum)
	if err != nil {
		logrus.Errorf("Error getting episodes: %v", err.Error())
		http.Error(w, "Error getting episodes", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodes)
}

func (h *Handler) GetServers(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Slug is required", http.StatusBadRequest)
		return
	}

	episode := r.URL.Query().Get("episode")
	if episode == "" {
		http.Error(w, "Episode is required", http.StatusBadRequest)
		return
	}

	if _, err := strconv.Atoi(episode); err != nil {
		http.Error(w, "Episode must be a number", http.StatusBadRequest)
		return
	}

	servers, err := h.scrapper.GetServers(slug, episode)
	if err != nil {
		logrus.Errorf("Error getting servers: %v", err.Error())
		http.Error(w, "Error getting servers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

func (h *Handler) PlayStreaming(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Slug is required", http.StatusBadRequest)
		return
	}

	server := r.URL.Query().Get("server")
	if server == "" {
		http.Error(w, "Server is required", http.StatusBadRequest)
		return
	}

	streamingURL, err := h.scrapper.GetStreaming(server, slug)
	if err != nil {
		logrus.Errorf("Error getting streaming URL: %v", err.Error())
		http.Error(w, "Error getting streaming URL", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, streamingURL, http.StatusFound)
}

func (h *Handler) GetSearch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	page := r.URL.Query().Get("page")
	var pageNum int
	var err error

	if page != "" {
		pageNum, err = strconv.Atoi(page)

		if err != nil {
			http.Error(w, "Page must be a number", http.StatusBadRequest)
			return
		}
	}

	searchResults, err := h.scrapper.GetSearch(name, pageNum)
	if err != nil {
		logrus.Errorf("Error getting search results: %v", err.Error())
		http.Error(w, "Error getting search results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searchResults)
}
