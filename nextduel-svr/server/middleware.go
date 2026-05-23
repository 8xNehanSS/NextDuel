package server

import (
	"log"
	"net/http"
)

func (s *Server) AuthMiddleware(next Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("REQUEST START [%s] %s", r.Method, r.URL.Path)
		ctx := NewUnauthenticatedContext(w, r, s.GetDbConnection(), s.config, s.GetGameManager())

		token, err := r.Cookie("NEXT-DUEL")
		if err != nil || token.Value == "" {
			ctx.Redirect("/login", http.StatusTemporaryRedirect)
			return
		}

		userID, valid := validateSession(token.Value)
		if !valid {
			ctx.Redirect("/login", http.StatusTemporaryRedirect)
			return
		}

		ctx.SetAuthenticated(userID)
		next(ctx)
		log.Printf("REQUEST END [%s] %s", r.Method, r.URL.Path)
	})
}

func (s *Server) PublicMiddleware(next Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewUnauthenticatedContext(w, r, s.GetDbConnection(), s.config, s.GetGameManager())
		next(ctx)
	})
}

func (s *Server) WsMiddleware(next Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("WS START [%s] %s", r.Method, r.URL.Path)
		ctx := NewUnauthenticatedContext(w, r, s.GetDbConnection(), s.config, s.GetGameManager())

		token := ctx.GetQueryParam("token")
		if token == "" {
			return
		}

		// fake decoded user (replace with JWT later)
		userID := token
		ctx.SetAuthenticated(userID)
		next(ctx)
		log.Printf("WS END [%s] %s", r.Method, r.URL.Path)
	})
}

func (s *Server) CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Vary", "Origin")

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")

		next.ServeHTTP(w, r)
	})
}
