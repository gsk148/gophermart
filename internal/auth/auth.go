package auth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo"
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

func GenerateCookie(c echo.Context, userID uint) error {
	secret := "topSecret"

	expirationTime := &jwt.NumericDate{Time: time.Now().Add(time.Hour)}
	claims := &jwtCustomClaims{
		userID,
		jwt.RegisteredClaims{
			ExpiresAt: expirationTime,
		},
	}

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

func Authorization() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			secret := "topSecret"

			cookie, err := c.Cookie(string(cookieName))
			if err != nil {
				return c.String(http.StatusUnauthorized, "")
			}
			token := cookie.Value

			claims := jwtCustomClaims{}
			parsedTokenInfo, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !parsedTokenInfo.Valid {
				return c.String(http.StatusUnauthorized, "")
			}

			return nil
		}
	}
}
