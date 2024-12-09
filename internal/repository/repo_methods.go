package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"online-library/internal/logger"
	"online-library/internal/models"
)

func (r *PostgresSongRepository) GetFilteredSongs(group string, title string, page int, limit int) ([]models.Song, error) {

	logger.Log.Debugf("GetFilteredSongs called with group: %s, title: %s, page: %d, limit: %d", group, title, page, limit)

	offset := (page - 1) * limit

	// Формируем SQL запрос с фильтрами
	query := `
		SELECT song_id, group_name, song, release_date, lyrics, link
		FROM songs
		WHERE (group_name ILIKE $1 OR $1 IS NULL)
		  AND (song ILIKE $2 OR $2 IS NULL)
		ORDER BY song
		LIMIT $3 OFFSET $4
	`

	// Подготовка аргументов для запроса
	args := []interface{}{group, title, limit, offset}

	logger.Log.Debugf("Executing query: %s with args: %v", query, args)

	// Выполнение запроса
	rows, err := r.db.Query(query, args...)
	if err != nil {
		logger.Log.Errorf("Error executing query: %v", err)
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var songs []models.Song
	for rows.Next() {
		var song models.Song
		if err := rows.Scan(&song.SongID, &song.Group, &song.Song, &song.ReleaseDate, &song.Lyrics, &song.Link); err != nil {
			logger.Log.Errorf("Error scanning row: %v", err)
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		songs = append(songs, song)
	}

	// Проверка на ошибки при переборе строк
	if err := rows.Err(); err != nil {
		logger.Log.Errorf("Error iterating rows: %v", err)
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return songs, nil

}

func (r *PostgresSongRepository) GetSongLyricsByID(songID int) (string, string, error) {
	logger.Log.Debugf("GetSongLyricsByID called with songID: %d", songID)

	var song string
	var lyrics sql.NullString // Используем sql.NullString для проверки наличия текста

	query := "SELECT song, lyrics FROM songs WHERE song_id = $1"
	err := r.db.QueryRow(query, songID).Scan(&song, &lyrics)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Warnf("Song with ID %d not found", songID)
			// Если песня не найдена
			return "", "", sql.ErrNoRows
		}
		logger.Log.Errorf("Database error: %v", err)
		// Любая другая ошибка базы данных
		return "", "", fmt.Errorf("database error: %w", err)
	}

	// Если текст песни отсутствует
	if !lyrics.Valid {
		logger.Log.Infof("Song found, but lyrics are missing for song ID %d", songID)
		return song, "", nil // Возвращаем песню, но пустой текст
	}

	logger.Log.Infof("Song found with lyrics for song ID %d", songID)
	return song, lyrics.String, nil
}

func (r *PostgresSongRepository) AddSong(group, song, releaseDate, text, link string) (int, error) {
	logger.Log.Debugf("AddSong called with group: %s, song: %s, releaseDate: %s", group, song, releaseDate)

	query := `INSERT INTO songs (group_name, song, release_date, lyrics, link) 
			  VALUES ($1, $2, $3, $4, $5)
			  RETURNING song_id
			  `
	var songID int
	err := r.db.QueryRow(query, group, song, releaseDate, text, link).Scan(&songID)
	if err != nil {
		logger.Log.Errorf("Failed to insert song: %v", err)
		return 0, fmt.Errorf("failed to insert song: %w", err)
	}
	logger.Log.Infof("Song added successfully with ID %d", songID)
	return songID, nil
}

func (r *PostgresSongRepository) UpdateSong(songID int, group, title, releaseDate, text, link string) error {
	logger.Log.Debugf("UpdateSong called for songID: %d", songID)

	query := `
		UPDATE songs
		SET group_name = $1, song = $2, release_date = $3, lyrics = $4, link = $5
		WHERE song_id = $6
		`
	_, err := r.db.Exec(query, group, title, releaseDate, text, link, songID)
	if err != nil {
		logger.Log.Errorf("Failed to update song ID %d: %v", songID, err)
		return fmt.Errorf("failed to update song: %w", err)
	}

	logger.Log.Infof("Song with ID %d updated successfully", songID)
	return nil
}

func (r *PostgresSongRepository) DeleteSong(songID int) error {
	logger.Log.Debugf("DeleteSong called for songID: %d", songID)

	query := `
		DELETE FROM songs
		WHERE song_id = $1
	`
	_, err := r.db.Exec(query, songID)
	if err != nil {
		logger.Log.Errorf("Failed to delete song ID %d: %v", songID, err)
		return fmt.Errorf("failed to delete song: %w", err)
	}
	return nil
}
