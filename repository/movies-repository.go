package repository

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/arias9306/golang-api-test/entity"
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
)

func init() {
	db, err = memdb.NewMemDB(shema)
	if err != nil {
		panic(err)
	}

	txn2 := db.Txn(true)

	movies := []*entity.Movie{
		{Title: "movie 1", ReleasedYear: 2001, Rating: 4.4, ID: 1, Genres: []string{"fiction", "comedy"}},
		{Title: "movie 2", ReleasedYear: 2002, Rating: 1.1, ID: 2, Genres: []string{"fiction", "drama"}},
		{Title: "movie 3", ReleasedYear: 2004, Rating: 2.4, ID: 3, Genres: []string{"action", "comedy"}},
		{Title: "movie 4", ReleasedYear: 2005, Rating: 5.2, ID: 4, Genres: []string{"comedy"}},
	}
	for _, p := range movies {
		if err := txn2.Insert("movie", p); err != nil {
			panic(err)
		}
	}

	txn2.Commit()
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
		movie := obj.(*entity.Movie)
		movies = append(movies, *movie)
	}

	return movies, nil
}

func (c *moviesRepository) GetAll() ([]entity.Movie, error) {
	var movies []entity.Movie
	txn := db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("movie", "id")
	if err != nil {
		return []entity.Movie{}, errors.New(err.Error())
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		movie := obj.(*entity.Movie)
		movies = append(movies, *movie)
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

	return *it.(*entity.Movie), nil
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
			obj, ok := raw.(*entity.Movie)
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
		movie := obj.(*entity.Movie)
		movies = append(movies, *movie)
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
		movie := obj.(*entity.Movie)
		movies = append(movies, *movie)
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
	movie := it.(*entity.Movie)
	movie.Rating = rating

	e := txn.Insert("movie", movie)
	if e != nil {
		return entity.Movie{}, errors.New(e.Error())
	}

	txn.Commit()

	return *movie, nil
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

	movie := it.(*entity.Movie)
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

	return *movie, nil

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
