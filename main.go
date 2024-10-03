package main

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	reposApiURL = "https://api.github.com/users/TeoMatosevic/repos?type=all"
	Language    = iota
	Readme      = iota
)

func main() {
	var repos Repositories
	var urls Urls
	scheduler(&repos, &urls)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET"},
		AllowHeaders: []string{"Origin"},
	}))

	router.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"repositories": repos,
		})
	})

	port := os.Getenv("HTTP_PLATFORM_PORT")

    if port == "" {
        port = "8080"
    }

	router.Run("127.0.0.1:" + port)
}
