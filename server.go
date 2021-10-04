package main

import (
	"os"

	"github.com/arias9306/golang-api-test/controllers"
	"github.com/arias9306/golang-api-test/services"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

var (
	movieService    services.MoviesService       = services.New()
	movieController controllers.MoviesController = controllers.New(movieService)
)

func main() {

	server := gin.Default()

	server.GET("/echo", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Up and Running",
		})
	})

	movies := server.Group("/movies")
	{
		movies.GET("", movieController.GetAll)
		movies.GET("/:id", movieController.GetByID)
		movies.PUT("/:id/rating", movieController.UpdateMovieRating)
		movies.PUT("/:id/genre", movieController.UpdateMovieGenre)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	server.Run(":"+port)
}
