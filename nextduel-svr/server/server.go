package server

import (
	"log"
	"net/http"
)

type Server struct {
	Mux *http.ServeMux
}

func NewServer() *Server {
	s := &Server{
		Mux: http.NewServeMux(),
	}
	return s
}

func (s *Server) Start(port string) error {
	log.Println("Server running on", port)
	http.HandleFunc("/ws", s.wsHandler)
	err := http.ListenAndServe(port, s.Mux)
	if err != nil {
		return err
	}
	return nil
}
