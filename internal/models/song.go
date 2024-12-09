package models

// Song представляет сущность песни.
// @Description Модель песни с основными атрибутами.
type Song struct {
	Group       string `json:"group"`
	Song        string `json:"song"`
	SongID      int    `json:"song_id,omitempty"`
	Lyrics      string `json:"lyrics,omitempty"`
	ReleaseDate string `json:"release_date,omitempty"`
	Link        string `json:"link,omitempty"`
}

type SongDetail struct {
	ReleaseDate string `json:"release_date"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}
