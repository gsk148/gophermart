package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"

	"github.com/gsk148/gophermart/internal/config"
	"github.com/gsk148/gophermart/internal/handlers"
	"github.com/gsk148/gophermart/internal/storage"
	customValidator "github.com/gsk148/gophermart/internal/validator"
)

func main() {
	cfg := config.MustLoad()

	db, err := storage.InitStorage(
		fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
			cfg.Host,
			cfg.Port,
			cfg.Username,
			cfg.Name,
			cfg.Password,
			cfg.Sslmode))
	if err != nil {
		log.Error("failed to init storage")
		os.Exit(1)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	//add validator
	e.Validator = customValidator.NewValidator(validator.New())

	//add handler
	handlers.NewHandler(*e, *db)
	if err := e.Start(":8080"); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
