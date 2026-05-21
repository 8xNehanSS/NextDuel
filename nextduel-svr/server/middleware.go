package server

import (
	"net/http"
)

func (s *Server) AuthMiddleware(next Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewUnauthenticatedContext(w, r, s.GetDbConnection(), s.config)

		token, err := r.Cookie("NEXT-DUEL")
		if err != nil || token.Value == "" {
			ctx.Redirect("/login", http.StatusTemporaryRedirect)
			return
		}

		// fake decoded user (replace with JWT later)
		userID := "nehan"
		ctx.SetAuthenticated(userID)
		next(ctx)
	})
}

func (s *Server) PublicMiddleware(next Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewUnauthenticatedContext(w, r, s.GetDbConnection(), s.config)
		next(ctx)
	})
}
