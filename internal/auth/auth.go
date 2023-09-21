package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type (
	cookie string
	user   string
)

const (
	cookieName cookie = "authToken"
	userID     user   = "userID"
)

type jwtCustomClaims struct {
	UserID uint
	jwt.RegisteredClaims
}

func GenerateCookie(w http.ResponseWriter, userID uint) error {
	secret := "topSecret"

	expirationTime := &jwt.NumericDate{Time: time.Now().Add(time.Hour)}
	claims := jwtCustomClaims{}
	claims.UserID = userID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return err
	}
	cookie := new(http.Cookie)
	cookie.Name = string(cookieName)
	cookie.Value = tokenString
	cookie.Expires = expirationTime.Time
	http.SetCookie(w, cookie)
	return nil
}

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

func GetUserIDFromToken(w http.ResponseWriter, r *http.Request) int {
	userID, ok := r.Context().Value(userID).(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return 0
	}
	return userID
}
