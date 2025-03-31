package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gopkg.in/fsnotify.v1"
)

const (
	reposApiURL = "https://api.github.com/users/TeoMatosevic/repos?type=all"
	orgsApiURL  = "https://api.github.com/users/TeoMatosevic/orgs?type=all"
	Language    = iota
	Readme      = iota
	Org         = iota
)

func main() {
	initDatabase()

	repos := Repositories{}
	orgs := Organizations{}
	urls := Urls{}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if strings.HasSuffix(event.Name, "app_offline.htm") {
					fmt.Println("Exiting...")
					os.Exit(0)
				}
			}
		}
	}()

	currentDir, err := os.Getwd()
	if err := watcher.Add(currentDir); err != nil {
		fmt.Println("Error:", err)
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET"},
		AllowHeaders: []string{"Origin"},
	}))

	router.GET("/api/v1/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"repositories":  toRepositories(repos.read()),
			"organizations": orgs.read(),
		})
	})

	router.POST("/api/v1/repos", func(c *gin.Context) {
		u, err := getRepositories(&repos, &urls, &orgs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"urls": u.u,
			"type": u.t,
		})
	})

	router.POST("/api/v1/urls", func(c *gin.Context) {
		u := sendRequests(&repos, &urls, &orgs)
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
