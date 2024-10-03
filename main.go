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
	Language    = iota
	Readme      = iota
)

func main() {
	var repos Repositories
	var urls Urls
	scheduler(&repos, &urls)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error:", err)
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
			"repositories": repos,
		})
	})

	port := os.Getenv("HTTP_PLATFORM_PORT")

	if port == "" {
		port = "8080"
	}

	router.Run("127.0.0.1:" + port)
}
