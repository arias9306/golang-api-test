package controllers

import (
	"net/http"

	"github.com/arias9306/golang-api-test/entity"
	"github.com/arias9306/golang-api-test/services"
	"github.com/gin-gonic/gin"
)

type MoviesController interface {
	GetAll(c *gin.Context)
	GetByID(c *gin.Context)
	UpdateMovieRating(c *gin.Context)
	UpdateMovieGenre(c *gin.Context)
}

type moviesController struct {
	moviesService services.MoviesService
}

type ratingBody struct {
	Rating float32 `json:"rating" binding:"required,min=0,max=10"`
}

type genresBody struct {
	NewGenres    []string `json:"newGenres"`
	RemoveGenres []string `json:"removeGenres"`
}

func New(movieService services.MoviesService) MoviesController {
	return &moviesController{
		moviesService: movieService,
	}
}

func (c *moviesController) GetAll(cxt *gin.Context) {
	var movies []entity.Movie = make([]entity.Movie, 0)
	var err error
	var movie entity.Movie
	title := cxt.DefaultQuery("title", "")
	id := cxt.DefaultQuery("id", "")
	releasedYear := cxt.DefaultQuery("released", "")
	genre := cxt.DefaultQuery("genre", "")
	rating := cxt.DefaultQuery("rating", "")
	ratingOrder := cxt.DefaultQuery("order", "higher")

	if id != "" {
		movie, err = c.moviesService.GetById(id)
		if err != nil {
			cxt.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		movies = append(movies, movie)
	}

	if title != "" {
		movies, err = c.moviesService.FindByStringField("title", title)
		if err != nil {
			cxt.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

	}

	if releasedYear != "" {
		releasedMovies, err := c.moviesService.GetByReleasedYears(releasedYear)
		if err != nil {
			cxt.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		movies = append(movies, releasedMovies...)
	}

	if genre != "" {
		movies, err = c.moviesService.FindByStringField("genres", genre)
		if err != nil {
			cxt.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
	}

	if rating != "" {
		moviesByRating, err := c.moviesService.FindByRating(rating, ratingOrder)
		if err != nil {
			cxt.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		movies = append(movies, moviesByRating...)
	}

	if len(movies) == 0 {
		movies, err = c.moviesService.GetAll()
		if err != nil {
			cxt.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
	}

	cxt.JSON(200, movies)
}

func (c *moviesController) GetByID(cxt *gin.Context) {
	movie, err := c.moviesService.GetById(cxt.Param("id"))
	if err != nil {
		cxt.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cxt.JSON(200, movie)
}

func (c *moviesController) UpdateMovieRating(cxt *gin.Context) {
	body := new(ratingBody)
	err := cxt.Bind(body)
	if err != nil {
		cxt.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movie, e := c.moviesService.UpdateRating(cxt.Param("id"), body.Rating)

	if e != nil {
		cxt.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}

	cxt.JSON(200, movie)
}

func (c *moviesController) UpdateMovieGenre(cxt *gin.Context) {
	body := new(genresBody)
	err := cxt.Bind(body)
	if err != nil {
		cxt.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	movie, e := c.moviesService.UpdateGenre(cxt.Param("id"), body.NewGenres, body.RemoveGenres)

	if e != nil {
		cxt.JSON(http.StatusBadRequest, gin.H{"error": e.Error()})
		return
	}

	cxt.JSON(200, movie)
}
