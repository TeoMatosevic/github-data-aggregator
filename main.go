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
	initDatabase()

	repos := Repositories{}
	urls := Urls{}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET"},
		AllowHeaders: []string{"Origin"},
	}))

	router.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"repositories": toRepositories(repos.read()),
		})
	})

	router.POST("/api/v1/repos", func(c *gin.Context) {
		u, err := getRepositories(&repos, &urls)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"urls": u,
		})
	})

	router.POST("/api/v1/urls", func(c *gin.Context) {
		u := sendRequests(&repos, &urls)
		c.JSON(http.StatusOK, gin.H{
			"urls": u,
		})
	})

	port := os.Getenv("HTTP_PLATFORM_PORT")

	if port == "" {
		port = "8080"
	}

	if os.Getenv("ENVIRONMENT") == "local" {
		router.Run("127.0.0.1:" + port)
	} else {
		router.Run(":" + port)
	}
}
