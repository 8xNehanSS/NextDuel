package routes

import (
	"net/http"
	"nextduel-svr/routes/health"
	"nextduel-svr/routes/login"
	"nextduel-svr/routes/register"
	"nextduel-svr/server"
)

func RegisterRoutes(s *server.Server) {
	mux := s.Router()

	// HTML File server
	fileServer := http.FileServer(http.Dir("./web/dist"))
	mux.Handle("/", fileServer)

	// api routes
	apiMux := http.NewServeMux()
	registerPublicRoutes(s, apiMux)
	registerPrivateRoutes(s, apiMux)
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// websocket route
	mux.HandleFunc("/ws", WSHandler)
}

func registerPrivateRoutes(s *server.Server, mux *http.ServeMux) {
	mux.Handle("GET /health", s.AuthMiddleware(health.HealthHandler))
}

func registerPublicRoutes(s *server.Server, mux *http.ServeMux) {
	mux.Handle("POST /login", s.PublicMiddleware(login.LoginHandler))
	mux.Handle("POST /register", s.PublicMiddleware(register.RegisterUser))
}
