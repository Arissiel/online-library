package database

import (
	"database/sql"
	"online-library/internal/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// RunMigrations выполняет миграции для создания структуры базы данных
func RunMigrations(db *sql.DB) error {
	logger.Log.Info("Running migrations...")

	// Экземпляр драйвера для работы с базой
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Log.Errorf("Failed to create migration driver: %v", err)
		return err
	}

	// Создаем объект миграций
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"songs_db",
		driver,
	)
	if err != nil {
		logger.Log.Errorf("Failed to initialize migrations: %v", err)
		return err
	}

	// Выполняем миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Log.Errorf("Failed to apply migrations: %v", err)
		return err
	}

	logger.Log.Info("Migrations applied successfully")
	return nil

}
