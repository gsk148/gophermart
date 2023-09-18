package auth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo"
)

type (
	cookie string
)

const (
	cookieName cookie = "authToken"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uint
}

func GenerateCookie(c echo.Context, userID uint) error {
	secret := "topSecret"
	expirationTime := &jwt.NumericDate{Time: time.Now().Add(time.Hour)}
	claims := Claims{}
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
	c.SetCookie(cookie)
	return nil
}
