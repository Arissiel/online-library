package database

import (
	"database/sql"
	"fmt"
	"online-library/internal/logger"

	_ "github.com/lib/pq"
)

// ConnectDatabase устанавливает соединение с базой данных
func ConnectDatabase(host, port, user, password, dbname string) (*sql.DB, error) {
	// Формирование строки подключения
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Подключение к базе данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Log.Errorf("Failed to connect to database: %v", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		logger.Log.Errorf("Failed to ping database: %v", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Log.Info("Successfully connected to the database")
	return db, nil
}
