package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"

	"net/http"

	externalapi "online-library/external_api"

	"online-library/internal/logger"
	"online-library/internal/models"
	"online-library/internal/repository"
	"strconv"
	"strings"
)

// SongHandlerInterface определяет контракт для обработки запросов песен.
type SongHandlerInterface interface {
	GetSongs(w http.ResponseWriter, r *http.Request)
	GetSongLyrics(w http.ResponseWriter, r *http.Request)
	AddSong(w http.ResponseWriter, r *http.Request)
	UpdateSong(w http.ResponseWriter, r *http.Request)
	DeleteSong(w http.ResponseWriter, r *http.Request)
}

// SongHandler реализует SongHandlerInterface.
type SongHandler struct {
	Repo        repository.SongRepository
	ExternalAPI externalapi.ExternalAPI
}

// ResponseLyrics структура ответа с текстом песни и пагинацией
type ResponseLyrics struct {
	Song     string   `json:"song"`      //название песни
	SongID   string   `json:"song_id"`   //id песни
	Lyrics   []string `json:"lyrics"`    //текст песни
	Page     int      `json:"page"`      //страница
	PageSize int      `json:"page_size"` //размер страницы
}

func NewSongHandler(repo repository.SongRepository, api externalapi.ExternalAPI) *SongHandler {
	return &SongHandler{
		Repo:        repo,
		ExternalAPI: api,
	}
}

// GetSongs возвращает список песен с фильтрацией.
// @Summary Get Songs
// @Description Получение списка песен с возможностью фильтрации по группе и названию.
// @Tags songs
// @Accept json
// @Produce json
// @Param group query string false "Название группы" example("Queen")
// @Param title query string false "Название песни" example("Bohemian Rhapsody")
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(10)
// @Success 200 {array} models.Song "Список песен"
// @Failure 400 {object} map[string]string "Ошибочные параметры запроса"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /songs [get]
func (h *SongHandler) GetSongs(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("GetSong handler invoked")

	// Чтение query параметров
	filters := r.URL.Query()
	group := filters.Get("group")
	title := filters.Get("title")
	pageStr := filters.Get("page")
	limitStr := filters.Get("limit")

	logger.Log.Debugf("Received query parameters: group=%s, title=%s, page=%s, limit=%s", group, title, pageStr, limitStr)

	if group == "" || title == "" || pageStr == "" || limitStr == "" {
		logger.Log.Warn("Missing query parameters")
		http.Error(w, "Missing parameters", http.StatusBadRequest)
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		logger.Log.Warn("Invalid or missing page parameter, defaulting to 1")
		page = 1 //по умолчанию
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		logger.Log.Warn("Invalid or missing limit parameter, defaulting to 10")
		limit = 10 //по умолчанию
	}

	//Получение данных из БД
	logger.Log.Debugf("Fetching songs from DB: group=%s, title=%s, page=%d, limit=%d", group, title, page, limit)
	songs, err := h.Repo.GetFilteredSongs(group, title, page, limit)
	if err != nil {
		logger.Log.Errorf("Failed to fetch songs from DB: %v", err)
		http.Error(w, "Failed to fetch songs", http.StatusInternalServerError)
		return
	}

	logger.Log.Infof("Successfully fetched %d songs from DB", len(songs))

	//Ответ клиенту
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(songs); err != nil {
		logger.Log.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// GetSongLyrics возвращает текст песни с возможностью пагинации.
// @Summary Get Song Lyrics
// @Description Получение текста песни по ID с возможностью разбивки на страницы.
// @Tags songs
// @Accept json
// @Produce json
// @Param id query int true "ID песни" example(1)
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество куплетов на странице" default(5)
// @Success 200 {object} ResponseLyrics "Текст песни с пагинацией"
// @Failure 400 {object} map[string]string "Ошибочные параметры запроса"
// @Failure 404 {object} map[string]string "Песня не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /songs/{id}/lyrics [get]
func (h *SongHandler) GetSongLyrics(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("GetSongLyrics handler invoked")

	query := r.URL.Query()
	// Получение ID песни из URL
	songIDStr := query.Get("id")
	if songIDStr == "" {
		logger.Log.Error("Missing query id parameter")
		http.Error(w, "Missing query id parameter", http.StatusBadRequest)
		return
	}

	logger.Log.Debugf("Received query id parameter: id=%s", songIDStr)
	songID, err := strconv.Atoi(songIDStr)
	if err != nil || songID <= 0 {
		logger.Log.Warnf("Invalid query id parameter: id=%s", songIDStr)
		http.Error(w, "Invalid query id parameter", http.StatusBadRequest)
		return
	}

	//Извлечение параметров для пагинации
	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		logger.Log.Warn("Invalid or missing page parameter, defaulting to 1")
		page = 1
	}

	size, err := strconv.Atoi(query.Get("limit"))
	if err != nil || size < 1 {
		logger.Log.Warn("Invalid or missing limit parameter, defaulting to 5")
		size = 5
	}

	//Извлечение текста песни из базы данных
	logger.Log.Debugf("Fetching lyrics for song ID %d with pagination: page=%d, size=%d", songID, page, size)
	song, lyrics, err := h.Repo.GetSongLyricsByID(songID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Если песня вообще не найдена
			logger.Log.Warnf("Song with ID %d not found", songID)
			http.Error(w, "Song not found", http.StatusNotFound)
		} else {
			// Любая другая ошибка
			logger.Log.Errorf("Failed to retrieve song details for ID %d: %v", songID, err)
			http.Error(w, "Failed to retrieve song details", http.StatusInternalServerError)
		}
		return
	}

	// Проверяем, есть ли текст песни
	if lyrics == "" {
		logger.Log.Infof("No lyrics found for song ID %d", songID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"song":    song,
			"song_id": songID,
			"error":   "Lyrics not found",
		})
		return
	}

	// Если песня и текст найдены:

	//Разделение текста на куплеты
	stanzas := strings.Split(lyrics, "\n\n") //Разбиваем текст по двойным переносам строк
	totalStanzas := len(stanzas)

	//Определение границ пагинации
	start := (page - 1) * size
	end := start + size

	if start >= totalStanzas {
		logger.Log.Warn("Page out of range")
		http.Error(w, "Page out of range", http.StatusBadRequest)
		return
	}

	if end > totalStanzas {
		end = totalStanzas
	}

	//Формирование ответа
	logger.Log.Infof("Successfully retrieved lyrics for song ID %d", songID)
	response := ResponseLyrics{
		Song:     song,
		SongID:   songIDStr,
		Lyrics:   stanzas[start:end],
		Page:     page,
		PageSize: size,
	}

	//Отправка ответа клиенту
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Log.Errorf("Failed to encode responsefor song ID %d: %v", songID, err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// AddSong добавляет новую песню в базу данных.
// @Summary Add Song
// @Description Добавление новой песни в базу данных. Данные о песне подтягиваются из внешнего API.
// @Tags songs
// @Accept json
// @Produce json
// @Param song body models.Song true "Данные песни"
// @Success 201 {object} map[string]int "ID добавленной песни"
// @Failure 400 {object} map[string]string "Ошибочный запрос"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /songs [post]
func (h *SongHandler) AddSong(w http.ResponseWriter, r *http.Request) {
	var song models.Song
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		logger.Log.Errorf("Invalid request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	logger.Log.Debugf("Received query parameters: group=%s, song=%s", song.Group, song.Song)

	if song.Group == "" || song.Song == "" {
		logger.Log.Warn("Group and Song are required fields")
		http.Error(w, "Group and Song are required fields", http.StatusBadRequest)
	}

	apiSongDetails, err := h.ExternalAPI.GetSongDetails(song.Group, song.Song)
	if err != nil {
		logger.Log.Errorf("Failed to fetch song details from external API: %v", err)
		http.Error(w, "Failed to fetch song details from external API", http.StatusInternalServerError)
		return
	}

	// Сохранение в базе данных
	songID, err := h.Repo.AddSong(song.Group, song.Song, apiSongDetails.ReleaseDate, apiSongDetails.Text, apiSongDetails.Link)
	if err != nil {
		logger.Log.Errorf("Failed to save song in database: %v", err)
		http.Error(w, "Failed to save song in database", http.StatusInternalServerError)
		return
	}

	// Ответ клиенту с ID новой песни
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": songID})
}

// UpdateSong обновляет данные существующей песни.
// @Summary Update Song
// @Description Обновление данных песни в базе по её ID.
// @Tags songs
// @Accept json
// @Produce json
// @Param id query int true "ID песни" example(1)
// @Param song body models.Song true "Обновленные данные песни"
// @Success 200 {string} string "Песня успешно обновлена"
// @Failure 400 {object} map[string]string "Ошибочный запрос"
// @Failure 404 {object} map[string]string "Песня не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /songs/{id} [put]
func (h *SongHandler) UpdateSong(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("UpdateSong handler invoked")

	// Извлечение song_id из запроса
	query := r.URL.Query()
	songIDStr := query.Get("id")
	if songIDStr == "" {
		logger.Log.Error("Missing query id parameter")
		http.Error(w, "Missing query id parameter'", http.StatusBadRequest)
		return
	}

	songID, err := strconv.Atoi(songIDStr)
	if err != nil || songID <= 0 {
		logger.Log.Warnf("Invalid query id parameter: id=%s", songIDStr)
		http.Error(w, "Invalid query id parameter", http.StatusBadRequest)
		return
	}

	// Чтение данных из тела запроса
	var updatedSong models.Song
	if err := json.NewDecoder(r.Body).Decode(&updatedSong); err != nil {
		logger.Log.Errorf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверка обязательных полей
	if updatedSong.Group == "" || updatedSong.Song == "" {
		logger.Log.Error("Group and Song fields are required")
		http.Error(w, "Group and Song fields are required", http.StatusBadRequest)
		return
	}

	logger.Log.Debugf("Received parameters: group=%s, song=%s, releaseDate=%s, lyrics=%s, link=%s", updatedSong.Group, updatedSong.Song, updatedSong.ReleaseDate, updatedSong.Lyrics, updatedSong.Link)
	// Вызов метода репозитория для обновления записи
	err = h.Repo.UpdateSong(songID, updatedSong.Group, updatedSong.Song, updatedSong.ReleaseDate, updatedSong.Lyrics, updatedSong.Link)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Warnf("Song with ID %d not found", songID)
			http.Error(w, "Song not found", http.StatusNotFound)
		} else {
			logger.Log.Errorf("Failed to update song with ID %d: %v", songID, err)
			http.Error(w, "Failed to update song", http.StatusInternalServerError)
		}
		return
	}

	// Успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Song updated successfully"))

}

// DeleteSong удаляет песню из базы данных по её ID.
// @Summary Delete Song
// @Description Удаление песни из базы данных по её ID.
// @Tags songs
// @Accept json
// @Produce json
// @Param id query int true "ID песни" example(1)
// @Success 200 {string} string "Песня успешно удалена"
// @Failure 400 {object} map[string]string "Ошибочный запрос"
// @Failure 404 {object} map[string]string "Песня не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /songs/{id} [delete]
func (h *SongHandler) DeleteSong(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("DeleteSong handler invoked")

	// Извлечение song_id из запроса
	query := r.URL.Query()
	songIDStr := query.Get("id")
	if songIDStr == "" {
		logger.Log.Error("Missing query id parameter")
		http.Error(w, "Missing query id parameter'", http.StatusBadRequest)
		return
	}

	songID, err := strconv.Atoi(songIDStr)
	if err != nil || songID <= 0 {
		logger.Log.Warnf("Invalid query id parameter: id=%s", songIDStr)
		http.Error(w, "Invalid query id parameter", http.StatusBadRequest)
		return
	}

	// Вызов метода репозитория для удаления записи
	err = h.Repo.DeleteSong(songID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Warnf("Song with ID %d not found", songID)
			http.Error(w, "Song not found", http.StatusNotFound)
		} else {
			logger.Log.Errorf("Failed to delete song with ID %d: %v", songID, err)
			http.Error(w, "Failed to delete song", http.StatusInternalServerError)
		}
		return
	}

	// Успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Song deleted successfully"))
}
