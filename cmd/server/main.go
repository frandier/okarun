package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yokai/internal/anime"
	"yokai/internal/config"
	"yokai/internal/handler"

	"github.com/common-nighthawk/go-figure"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Server struct {
	config *config.Config
	router *mux.Router
	server *http.Server
}

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
		router: mux.NewRouter(),
	}
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logrus.Infof("Started %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		logrus.Infof("Completed %s in %v", r.URL.Path, time.Since(start))
	})
}

func (s *Server) setupRoutes() {
	scraper := &anime.Jkanime{}
	handler := handler.NewHandler(*scraper)

	apiRouter := s.router.PathPrefix("/api").Subrouter()

	apiRouter.HandleFunc("/latest", handler.GetLatestEpisodes).Methods("GET")
	apiRouter.HandleFunc("/anime", handler.GetAnime).Methods("GET")
	apiRouter.HandleFunc("/episodes", handler.GetEpisodes).Methods("GET")
	apiRouter.HandleFunc("/servers", handler.GetServers).Methods("GET")
	apiRouter.HandleFunc("/play", handler.PlayStreaming).Methods("GET")
	apiRouter.HandleFunc("/search", handler.GetSearch).Methods("GET")

	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Welcome to Yokai API!"))
	}).Methods("GET")
}

func (s *Server) Run() error {
	s.setupRoutes()

	s.server = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      s.loggingMiddleware(s.router),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		logrus.Info("Shutting down server...")
		if err := s.server.Shutdown(ctx); err != nil {
			logrus.Error("Server shutdown error:", err)
		}
	}()

	logrus.Infof("Server running on port %s", s.config.Port)
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

func main() {
	goFigure := figure.NewColorFigure("Okarun", "", "Red", true)
	goFigure.Print()

	cfg := config.New()
	server := NewServer(cfg)

	if err := server.Run(); err != nil {
		logrus.Fatal("Error running server:", err)
	}
}
