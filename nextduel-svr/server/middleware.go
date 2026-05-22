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

		// fake decoded user (replace with JWT later)
		userID := "nehan"
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
