package repository

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/arias9306/golang-api-test/entity"
	"github.com/eefret/gomdb"
	"github.com/hashicorp/go-memdb"
)

var (
	db    *memdb.MemDB
	err   error
	shema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"movie": {
				Name: "movie",
				Indexes: map[string]*memdb.IndexSchema{
					"title": {
						Name:    "title",
						Indexer: &memdb.StringFieldIndex{Field: "Title"},
					},
					"releasedYear": {
						Name:    "releasedYear",
						Indexer: &memdb.IntFieldIndex{Field: "ReleasedYear"},
					},
					"rating": {
						Name:    "rating",
						Indexer: &memdb.StringFieldIndex{Field: "Rating"},
					},
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.IntFieldIndex{Field: "ID"},
					},
					"genres": {
						Name:    "genres",
						Indexer: &memdb.StringSliceFieldIndex{Field: "Genres"},
					},
				},
			},
		},
	}
	id int
)

const YOUR_API_KEY = "bd48bfa6"

func init() {
	db, err = memdb.NewMemDB(shema)
	if err != nil {
		panic(err)
	}
	id = 1
}

type MoviesRepository interface {
	FindByStringField(field, value string) ([]entity.Movie, error)
	GetAll() ([]entity.Movie, error)
	GetById(ID int) (entity.Movie, error)
	FindByReleasedRange(value []int) ([]entity.Movie, error)
	FindByRating(value float32, order string) ([]entity.Movie, error)
	UpdateRating(ID int, rating float32) (entity.Movie, error)
	UpdateGenres(ID int, newGenres, removeGenres []string) (entity.Movie, error)
}

type moviesRepository struct{}

func New() MoviesRepository {
	return &moviesRepository{}
}

func (c *moviesRepository) FindByStringField(field, value string) ([]entity.Movie, error) {
	var movies []entity.Movie = make([]entity.Movie, 0)

	txn := db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("movie", field, value)

	if err != nil {
		return []entity.Movie{}, errors.New(err.Error())
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		movie := obj.(entity.Movie)
		movies = append(movies, movie)
	}

	if field == "title" && len(movies) == 0 {
		movie, er := findMovieInOMDBApi(value)
		if er != nil {
			return []entity.Movie{}, errors.New(err.Error())
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func (c *moviesRepository) GetAll() ([]entity.Movie, error) {
	var movies []entity.Movie = make([]entity.Movie, 0)
	txn := db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("movie", "id")
	if err != nil {
		return []entity.Movie{}, errors.New(err.Error())
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		movie := obj.(entity.Movie)
		movies = append(movies, movie)
	}
	return movies, nil
}

func (c *moviesRepository) GetById(ID int) (entity.Movie, error) {

	txn := db.Txn(false)
	defer txn.Abort()

	it, err := txn.First("movie", "id", ID)
	if err != nil {
		return entity.Movie{}, errors.New(err.Error())
	}

	if it == nil {
		return entity.Movie{}, fmt.Errorf("the movie with id: '%s' do not exist", strconv.Itoa(ID))
	}

	return it.(entity.Movie), nil
}

func (c *moviesRepository) FindByReleasedRange(releasedYears []int) ([]entity.Movie, error) {
	var movies []entity.Movie = make([]entity.Movie, 0)
	txn := db.Txn(false)
	defer txn.Abort()

	if len := len(releasedYears); len != 2 {
		return []entity.Movie{}, errors.New("incorrect range")
	}

	it, err := txn.Get("movie", "releasedYear")

	if err != nil {
		return []entity.Movie{}, errors.New(err.Error())
	}

	filterReleasedRange := func(year []int) func(interface{}) bool {
		return func(raw interface{}) bool {
			obj, ok := raw.(entity.Movie)
			if !ok {
				return true
			}
			if obj.ReleasedYear >= year[0] && obj.ReleasedYear <= year[1] {
				return false
			}
			return true
		}
	}

	filter := memdb.NewFilterIterator(it, filterReleasedRange(releasedYears))

	for obj := filter.Next(); obj != nil; obj = filter.Next() {
		movie := obj.(entity.Movie)
		movies = append(movies, movie)
	}

	return movies, nil
}

func (c *moviesRepository) FindByRating(value float32, order string) ([]entity.Movie, error) {
	var movies []entity.Movie = make([]entity.Movie, 0)

	txn := db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("movie", "rating")

	if err != nil {
		return []entity.Movie{}, errors.New(err.Error())
	}

	filterRating := func(i float32, order string) func(interface{}) bool {
		return func(raw interface{}) bool {
			obj, ok := raw.(*entity.Movie)
			if !ok {
				return true
			}
			if order == "higher" {
				if obj.Rating >= i {
					return false
				}
			} else {
				if obj.Rating <= i {
					return false
				}
			}
			return true
		}
	}

	filter := memdb.NewFilterIterator(it, filterRating(value, order))

	for obj := filter.Next(); obj != nil; obj = filter.Next() {
		movie := obj.(entity.Movie)
		movies = append(movies, movie)
	}

	return movies, nil
}

func (c *moviesRepository) UpdateRating(ID int, rating float32) (entity.Movie, error) {
	txn := db.Txn(true)
	defer txn.Abort()

	it, err := txn.First("movie", "id", ID)
	if err != nil {
		return entity.Movie{}, errors.New(err.Error())
	}

	if it == nil {
		return entity.Movie{}, fmt.Errorf("the movie with id: '%s' do not exist", strconv.Itoa(ID))
	}
	movie := it.(entity.Movie)
	movie.Rating = rating

	e := txn.Insert("movie", movie)
	if e != nil {
		return entity.Movie{}, errors.New(e.Error())
	}

	txn.Commit()

	return movie, nil
}

func (c *moviesRepository) UpdateGenres(ID int, newGenres, removeGenres []string) (entity.Movie, error) {

	txn := db.Txn(true)
	defer txn.Abort()

	it, err := txn.First("movie", "id", ID)
	if err != nil {
		return entity.Movie{}, errors.New(err.Error())
	}

	if it == nil {
		return entity.Movie{}, fmt.Errorf("the movie with id: '%s' do not exist", strconv.Itoa(ID))
	}

	movie := it.(entity.Movie)
	movie.Genres = appendGenre(movie.Genres, newGenres)
	sort.Strings(movie.Genres)

	for _, genre := range removeGenres {

		if contains(movie.Genres, genre) {
			index := sort.StringSlice(movie.Genres).Search(genre)
			if index >= 0 {
				movie.Genres = removeIndex(movie.Genres, index)
			}
		}
	}

	e := txn.Insert("movie", movie)
	if e != nil {
		return entity.Movie{}, errors.New(e.Error())
	}

	txn.Commit()

	return movie, nil

}

func appendGenre(a []string, b []string) []string {

	check := make(map[string]int)
	d := append(a, b...)
	res := make([]string, 0)
	for _, val := range d {
		check[val] = 1
	}

	for letter := range check {
		res = append(res, letter)
	}

	return res
}

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

func findMovieInOMDBApi(movieTitle string) (entity.Movie, error) {
	var movies []entity.Movie = make([]entity.Movie, 0)

	api := gomdb.Init(YOUR_API_KEY)
	query := &gomdb.QueryData{Title: movieTitle, SearchType: gomdb.MovieSearch}
	res, err := api.Search(query)
	if err != nil {
		return entity.Movie{}, errors.New(err.Error())
	}

	for _, movie := range res.Search {
		res, err := api.MovieByImdbID(movie.ImdbID)
		if err != nil {
			return entity.Movie{}, errors.New(err.Error())
		}

		rating, err := strconv.ParseFloat(res.ImdbRating, 32)
		if err != nil {
			return entity.Movie{}, errors.New(err.Error())
		}

		var years []string
		if len := len(res.Year); len > 4 {
			years = strings.Split(res.Year, "–")
		} else {
			years = append(years, res.Year)
		}

		year, err := strconv.Atoi(years[0])
		if err != nil {
			return entity.Movie{}, errors.New(err.Error())
		}

		genres := strings.Split(res.Genre, ",")
		movieGenres := make([]string, 0)

		for _, genre := range genres {
			movieGenres = append(movieGenres, strings.TrimSpace(genre))
		}

		movie := &entity.Movie{
			ID:           id,
			Title:        res.Title,
			ReleasedYear: int(year),
			Rating:       float32(rating),
			Genres:       movieGenres,
		}

		id++

		movies = append(movies, *movie)
	}

	txn := db.Txn(true)
	defer txn.Abort()

	for _, movi := range movies {
		err := txn.Insert("movie", movi)
		if err != nil {
			return entity.Movie{}, errors.New(err.Error())
		}
	}

	txn.Commit()
	return movies[0], nil
}
