package entity

type Movie struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	ReleasedYear int      `json:"releasedYear"`
	Rating       float32  `json:"rating"`
	Genres       []string `json:"genres"`
}