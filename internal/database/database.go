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

// CheckAndCreateDatabase проверяет существование базы данных и создает её, если она отсутствует
func CheckAndCreateDatabase(host, port, user, password, dbname string) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable dbname=postgres",
		host, port, user, password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Log.Errorf("Failed to connect to PostgreSQL to check database: %v", err)
		return fmt.Errorf("failed to connect to PostgreSQL to check database: %w", err)
	}
	defer db.Close()

	// Проверяем, существует ли база данных
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", dbname)
	err = db.QueryRow(query).Scan(&exists)
	if err != nil {
		logger.Log.Errorf("Error checking database existence: %v", err)
		return fmt.Errorf("error checking database existence: %w", err)
	}

	if exists {
		logger.Log.Infof("Database %s already exists", dbname)
		return nil
	}

	// Создаем базу данных, если она не существует
	createQuery := fmt.Sprintf("CREATE DATABASE %s", dbname)
	_, err = db.Exec(createQuery)
	if err != nil {
		logger.Log.Errorf("Failed to create database %s: %v", dbname, err)
		return fmt.Errorf("failed to create database %s: %w", dbname, err)
	}

	logger.Log.Infof("Database %s created successfully", dbname)
	return nil
}
