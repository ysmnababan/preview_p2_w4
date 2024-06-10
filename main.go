package main

import (
	"pagi/config"
	"pagi/handler"
	"pagi/helper"
	"pagi/models"

	"github.com/labstack/echo/v4"
)

func main() {
	db := config.Connect()

	db.AutoMigrate(&models.Player{}, &models.User{})
	handler := &handler.Repo{DB: db}

	e := echo.New()
	e.POST("/auth/login", handler.Login)
	e.POST("/auth/register", handler.Register)
	e.POST("/auth/refresh", handler.RefreshToken)
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			helper.Logging(c).Info("Api Request")
			return next(c)
		}
	})

	// create new group
	player := e.Group("")
	player.Use(helper.Auth)
	{
		player.GET("/players", handler.GetPlayer)
		player.POST("/players", handler.CreatePlayer)
		player.PUT("/players/:id", handler.UpdatePlayer)
	}

	e.Logger.Fatal(e.Start(":8080"))
}
