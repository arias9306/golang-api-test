package entity

type Movie struct {
	Title        string   `json:"title"`
	ReleasedYear int      `json:"releasedYear"`
	Rating       float32  `json:"rating"`
	ID           int      `json:"id"`
	Genres       []string `json:"genres"`
}