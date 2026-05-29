package server

import (
	"log"
	"net/http"
	"nextduel-svr/gamepkg"
	"nextduel-svr/utils"

	"gorm.io/gorm"
)

type Server struct {
	mux      *http.ServeMux
	database *gorm.DB
	config   utils.Config

	// game functionality
	gameManager *gamepkg.GameManager
}

func NewServer() *Server {
	config := utils.LoadConfig()

	s := &Server{
		mux:    http.NewServeMux(),
		config: config,
	}
	return s
}

func (s *Server) GetDbConnection() *gorm.DB {
	return s.database
}

func (s *Server) SetDatabase(db *gorm.DB) {
	s.database = db
}

func (s *Server) GetGameManager() *gamepkg.GameManager {
	return s.gameManager
}

func (s *Server) SetGameManager(game *gamepkg.GameManager) {
	s.gameManager = game
}

func (s *Server) Router() *http.ServeMux {
	return s.mux
}

func (s *Server) Start() error {
	log.Println("Server running on", s.config.Port)
	router := s.CorsMiddleware(s.Router())
	err := http.ListenAndServe(":"+s.config.Port, router)
	if err != nil {
		return err
	}
	return nil
}
