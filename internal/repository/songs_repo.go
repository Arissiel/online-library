package repository

import (
	"database/sql"
	"online-library/internal/models"
)

type SongRepository interface {
	GetSongLyricsByID(songID int) (string, string, error) //возвращаем и название песни для удобства пользователя
	GetFilteredSongs(group string, title string, page int, limit int) ([]models.Song, error)
	AddSong(group, song, releaseDate, text, link string) (int, error)
	UpdateSong(songID int, group, title, releaseDate, text, link string) error
	DeleteSong(songID int) error
}

type PostgresSongRepository struct {
	db *sql.DB
}

func NewPostgresSongRepository(db *sql.DB) *PostgresSongRepository {
	return &PostgresSongRepository{db: db}
}
