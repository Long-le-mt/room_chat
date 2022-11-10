package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/ws", func(c *gin.Context) {
		ServeWs(c.Writer, c.Request)
	})

	router.Run("localhost:8080")
}
