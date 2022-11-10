package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	hub := newHub()
	go hub.Run()

	router.GET("/ws", func(c *gin.Context) {
		ServeWs(hub, c.Writer, c.Request)
	})

	router.Run("localhost:8080")
}
