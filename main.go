package main

import (
	"net/http"

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
	router.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"repositories": repos,
		})
	})

	router.Run("127.0.0.1:8080")
}
