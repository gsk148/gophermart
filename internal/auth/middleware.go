package auth

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func Authorization(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := "topSecret"

		cookie, err := r.Cookie(string(cookieName))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		token := cookie.Value

		claims := jwtCustomClaims{}
		parsedTokenInfo, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !parsedTokenInfo.Valid {
			w.WriteHeader(http.StatusUnauthorized)
		}

		ctx := context.WithValue(r.Context(), userID, claims.UserID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
