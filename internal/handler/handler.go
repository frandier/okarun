package handler

import "yokai/internal/anime"

type Handler struct {
	scrapper anime.Jkanime
}

func NewHandler(scrapper anime.Jkanime) *Handler {
	return &Handler{
		scrapper: scrapper,
	}
}
