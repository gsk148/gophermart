package handlers

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/gsk148/gophermart/internal/auth"
	"github.com/gsk148/gophermart/internal/models"
	"github.com/gsk148/gophermart/internal/storage"
	"github.com/gsk148/gophermart/internal/utils"
)

type Handler struct {
	Router echo.Echo
	Store  storage.Storage
}

func NewHandler(Router echo.Echo, DB storage.Storage) {
	h := &Handler{
		Router,
		DB,
	}
	h.init()
}

func (h *Handler) init() {
	h.Router.POST("/api/users/login", h.login)
	h.Router.POST("/api/users/register", h.register)
}

func (h *Handler) login(c echo.Context) error {
	userAPI := &models.User{}

	if err := c.Bind(userAPI); err != nil {
		return c.String(http.StatusBadRequest, "Not valid request")
	}

	if err := c.Validate(userAPI); err != nil {
		return c.String(http.StatusBadRequest, "Validation error")
	}

	userDB, err := h.Store.GetUserByLogin(userAPI.Login)

	if err != nil && err != storage.ErrNoDBResult {
		return c.String(http.StatusInternalServerError, "Unexpected error")
	}

	if !utils.CheckHashAndPassword(userDB.Password, userAPI.Password) {
		return c.String(http.StatusUnauthorized, "Unauthorized user")
	}

	err = auth.GenerateCookie(c, userDB.ID)
	if err != nil {
		return c.String(http.StatusUnauthorized, "Failed to generate cookie")
	}

	return c.String(http.StatusOK, "OK")
}

func (h *Handler) register(c echo.Context) error {
	userAPI := &models.User{}

	if err := c.Bind(userAPI); err != nil {
		return c.String(http.StatusBadRequest, "Not valid request")
	}

	if err := c.Validate(userAPI); err != nil {
		return c.String(http.StatusBadRequest, "Validation error")
	}

	userDB, err := h.Store.GetUserByLogin(userAPI.Login)
	switch {
	case err == nil && userDB.Login != "":
		return c.String(http.StatusConflict, "Login exists")
	case err == storage.ErrNoDBResult:
		cryptedPsw, err := utils.HashString(userAPI.Password)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Internal server error")
		}

		userID, err := h.Store.Register(userAPI.Login, cryptedPsw)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Internal server error")
		}

		err = auth.GenerateCookie(c, userID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Internal server error")
		}
	default:
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	return nil
}
