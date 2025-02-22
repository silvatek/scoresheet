package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Add handlers for requests only made by malicious bots
func AddBotHandlers(e *echo.Echo) {
	e.GET("/*.php", BotHandler)
	e.GET("/*/*.php", BotHandler)
	e.GET("/wp-admin*", BotHandler)
	e.GET("/wp*", BotHandler)
	e.GET("/.env*", BotHandler)
	e.GET("/.git*", BotHandler)
	e.GET("/phpmyadmin*", BotHandler)
}

// Return a minimal "not found" response
func BotHandler(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}
