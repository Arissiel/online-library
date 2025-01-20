package main

import (
	"net/http"

	"online-library/config"
	"online-library/internal/database"
	"online-library/internal/logger"
	"online-library/internal/routes"
)

func main() {
	// Инициализация логгера
	logger.InitLogger()

	logger.Log.Infof("Logger initialized")

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatalf("Failed to load config: %v", err)
	}

	// Подключение к базе данных
	db, err := database.ConnectDatabase(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		logger.Log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Выполнение миграций
	if err := database.RunMigrations(db); err != nil {
		logger.Log.Fatalf("Failed to apply migrations: %v", err)
	}

	// Инициализация маршрутов
	router := routes.NewRouter(db)

	// Запуск HTTP-сервера
	logger.Log.Infof("Starting server on port %s...", cfg.ServerPort)
	logger.Log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, router))

}
