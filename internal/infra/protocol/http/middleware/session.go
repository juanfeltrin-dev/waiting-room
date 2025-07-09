package middleware

import (
	"net/http"
	"waitingroom/internal/infra/jwt"
	"waitingroom/internal/schemas/request"
)

func ValidateSessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := request.Header{
			Authorization: r.Header.Get("Authorization"),
		}
		if h.Authorization == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		_, err := jwt.VerifyToken(h.GetToken())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
