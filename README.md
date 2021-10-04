# Golang API test

## Getting Started

Api to search for movies, I used a local database in memory (go-memdb), in case of performing a search and no information is found in the local database, the movie is searched in the gombd api and the search results are stored in the local database for future searches.

### Run the app

```console
$ go run server.go
```

### Endpoints

```
GET - /echo - Check if the API is up and running
GET - /movies - Get all movies
GET - /movies/{id} - Get movie by Id
GET - /movies?id={id} - Get movie by Id
GET - /movies?title={title} - Get movie by title
GET - /movies?genre={genre} - Get movie by genre
GET - /movies?released={year1} - Get movie by released year
GET - /movies?released={year1}-{year2} - Get movie by released year
GET - /movies?rating={rating}&order=higher - Get movie by rating higher
GET - /movies?rating=3&order=lower - Get movie by rating lower
```

Import the postman collection `Movies.postman_collection.json`

heroku
