package server

import (
	"log"
	"net/http"
	"nextduel-svr/utils"

	"gorm.io/gorm"
)

type Server struct {
	mux      *http.ServeMux
	database *gorm.DB
	config   utils.Config
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
