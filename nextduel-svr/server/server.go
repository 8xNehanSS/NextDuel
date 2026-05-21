package server

import (
	"log"
	"net/http"
	"nextduel-svr/game"
	"nextduel-svr/utils"

	"gorm.io/gorm"
)

type Server struct {
	mux      *http.ServeMux
	database *gorm.DB
	config   utils.Config

	// game functionality
	GameManager *game.GameManager
}

func NewServer() *Server {
	config := utils.LoadConfig()
	gameManager := game.NewGameManager()

	s := &Server{
		mux:         http.NewServeMux(),
		config:      config,
		GameManager: gameManager,
	}
	return s
}

func (s *Server) GetDbConnection() *gorm.DB {
	return s.database
}

func (s *Server) SetDatabase(db *gorm.DB) {
	s.database = db
}

func (s *Server) Router() *http.ServeMux {
	return s.mux
}

func (s *Server) Start() error {
	log.Println("Server running on", s.config.Port)
	err := http.ListenAndServe(":"+s.config.Port, s.Router())
	if err != nil {
		return err
	}
	return nil
}
