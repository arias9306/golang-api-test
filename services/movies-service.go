package services

import (
	"errors"
	"strconv"
	"strings"

	"github.com/arias9306/golang-api-test/entity"
	"github.com/arias9306/golang-api-test/repository"
)

type MoviesService interface {
	FindByStringField(field, value string) ([]entity.Movie, error)
	GetAll() ([]entity.Movie, error)
	GetById(ID string) (entity.Movie, error)
	GetByReleasedYears(releasedYears string) ([]entity.Movie, error)
	FindByRating(value string, order string) ([]entity.Movie, error)
	UpdateRating(ID string, rating float32) (entity.Movie, error)
	UpdateGenre(ID string, newGenres, removeGenres []string) (entity.Movie, error)
}

type moviesService struct {
	movieRepository repository.MoviesRepository
}

func New() MoviesService {
	movieRepository := repository.New()
	return &moviesService{
		movieRepository: movieRepository,
	}
}

func (c *moviesService) FindByStringField(field, value string) ([]entity.Movie, error) {
	return c.movieRepository.FindByStringField(field, value)
}

func (c *moviesService) GetAll() ([]entity.Movie, error) {
	return c.movieRepository.GetAll()
}

func (c *moviesService) GetById(ID string) (entity.Movie, error) {

	id, err := strconv.Atoi(ID)
	if err != nil {
		return entity.Movie{}, errors.New(err.Error())
	}

	movie, e := c.movieRepository.GetById(id)

	if e != nil {
		return entity.Movie{}, errors.New(e.Error())
	}

	return movie, nil
}

func (c *moviesService) GetByReleasedYears(releasedYears string) ([]entity.Movie, error) {
	var years []int
	twoDates := strings.Contains(releasedYears, "-")

	if !twoDates && len(releasedYears) != 4 {
		return []entity.Movie{}, errors.New("the released year param is not valid")
	}

	var sliceOfYears []string
	if twoDates {
		sliceOfYears = strings.Split(releasedYears, "-")
	} else {
		sliceOfYears = []string{releasedYears, releasedYears}
	}

	for _, year := range sliceOfYears {
		validYear, err := strconv.Atoi(year)
		if err != nil {
			return []entity.Movie{}, errors.New("the range year is not valid")
		}
		years = append(years, validYear)
	}

	len := len(years);

	if  len != 2 {
		return []entity.Movie{}, errors.New("the range year is not valid")
	}

	if len == 2 {
		year1, year2 := years[0], years[1]
		if year1 > year2 {
			return []entity.Movie{}, errors.New("the range year is not valid")
		}
	}


	movies, e := c.movieRepository.FindByReleasedRange(years)
	if e != nil {
		return []entity.Movie{}, errors.New(e.Error())
	}
	return movies, nil
}

func (c *moviesService) FindByRating(value string, order string) ([]entity.Movie, error) {
	ID, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return []entity.Movie{}, errors.New(err.Error())
	}
	movies, e := c.movieRepository.FindByRating(float32(ID), order)
	if e != nil {
		return []entity.Movie{}, errors.New(e.Error())
	}
	return movies, nil
}

func (c *moviesService) UpdateRating(ID string, rating float32) (entity.Movie, error) {
	id, err := strconv.Atoi(ID)
	if err != nil {
		return entity.Movie{}, errors.New(err.Error())
	}

	return c.movieRepository.UpdateRating(int(id), rating)
}

func (c *moviesService) UpdateGenre(ID string, newGenres, removeGenres []string) (entity.Movie, error) {
	id, err := strconv.Atoi(ID)
	if err != nil {
		return entity.Movie{}, errors.New(err.Error())
	}

	return c.movieRepository.UpdateGenres(id, newGenres, removeGenres)
}
