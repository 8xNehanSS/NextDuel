package routes

import (
	"nextduel-svr/routes/health"
	"nextduel-svr/server"
)

func RegisterRoutes(s *server.Server) {

	// Health
	s.Mux.HandleFunc("/health", health.HealthHandler)

	// // WebSocket
	// s.Mux.HandleFunc("/ws", handlers.WSHandler)

	// // Game APIs
	// s.Mux.HandleFunc("/api/join", handlers.JoinQueueHandler)
}
