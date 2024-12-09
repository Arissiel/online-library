package database

import (
	"database/sql"
	"online-library/internal/logger"
)

// RunMigrations выполняет миграции для создания структуры базы данных
func RunMigrations(db *sql.DB) error {
	queries := []string{
		// Создание таблицы
		`
		CREATE TABLE IF NOT EXISTS songs (
			song_id SERIAL PRIMARY KEY,
			group_name VARCHAR(255) NOT NULL,
			song VARCHAR(255) NOT NULL,
			release_date VARCHAR(255) NOT NULL,
			lyrics TEXT NOT NULL,
			link TEXT NOT NULL
		);
		`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			logger.Log.Error(err)
			return err
		}
	}

	logger.Log.Info("Migrations applied successfully")
	return nil
}
