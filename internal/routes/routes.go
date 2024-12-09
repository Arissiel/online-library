package routes

import (
	"database/sql"
	"net/http"
	"strconv"

	externalapi "online-library/external_api"
	"online-library/internal/handlers"
	"online-library/internal/logger"
	"online-library/internal/repository"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// NewRouter создает маршрутизатор для всех эндпоинтов
func NewRouter(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	//
	repo := repository.NewPostgresSongRepository(db)

	//
	externalAPI := externalapi.NewExternalAPIClient(viper.GetString("EXTERNAL_API_FULL_URL"), viper.GetString("EXTERNAL_API_METHOD"))

	// Инициализация обработчиков
	songHandler := handlers.NewSongHandler(repo, externalAPI)

	// Определение маршрутов
	mux.HandleFunc("/songs", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			songHandler.GetSongs(w, r)
		case http.MethodPost:
			songHandler.AddSong(w, r)
		default:
			logger.Log.WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
			}).Warn("Method not allowed")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/songs/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		// Получение ID песни из URL
		songIDStr := query.Get("id")
		if songIDStr == "" {
			logger.Log.Warn("Missing query id parameter")
			http.Error(w, "Missing query id parameter", http.StatusBadRequest)
			return
		}

		songID, err := strconv.Atoi(songIDStr)
		if err != nil || songID <= 0 {
			logger.Log.Warn("Invalid query id parameter")
			http.Error(w, "Invalid query id parameter", http.StatusBadRequest)
			return
		}

		logger.Log.Debugf("Parsed song ID from query: %d", songID)

		switch r.Method {
		case http.MethodGet:
			songHandler.GetSongLyrics(w, r)
		case http.MethodPut:
			songHandler.UpdateSong(w, r)
		case http.MethodDelete:
			songHandler.DeleteSong(w, r)
		default:
			logger.Log.WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
			}).Warn("Method not allowed")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
