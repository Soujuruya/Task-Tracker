package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/pkg/ctxkeys"
	"task-tracker-1/internal/transport/dto"
	"time"
)

// Добавлена новая функция для парсинга токена из заголовка Authorization
func parseBearerToken(r *http.Request) (string, error) {
	tokenString := r.Header.Get("Authorization")
	token := strings.TrimPrefix(tokenString, "Bearer ")
	if token == "" {
		return "", domain.ErrMissingToken
	}
	return token, nil
}

// Добавлена новая фукнция извлечения из контекста UserID
func getUserID(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(ctxkeys.UserIDKey).(string)
	if !ok || userID == "" {
		return "", domain.ErrInvalidUserID
	}
	return userID, nil
}

// Добавлена фукнция извлечения ID из path
func getPathID(r *http.Request) (string, error) {
	id := r.PathValue("id")
	if id == "" {
		return "", domain.ErrMissingPathID
	}
	return id, nil
}

// Фукнция для декодирования JSON
func decodeJSON(w http.ResponseWriter, r *http.Request, to any) bool {
	if err := json.NewDecoder(r.Body).Decode(to); err != nil {
		slog.Error("invalid request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}

// Функция для кодирования в JSON
func encodeJSON(w http.ResponseWriter, status int, data any) bool {
	var buf bytes.Buffer
	// cначала сериализуем в буфер,если Encode упадёт, в w ещё ничего не ушло
	// и мы можем вернуть 500. если писать напрямую в w,то статус уже не изменить, клиент получит 200 с пустым телом
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		slog.Error("encode response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return false
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(buf.Bytes())
	return true
}

// Функция извлечения request-id
func getRequestID(r *http.Request) string {
	id, _ := r.Context().Value(ctxkeys.RequestIDKey).(string)
	return id
}

// Фукнция для извлечения query-параметров и их валидации
func fromQueryParams(r *http.Request) (dto.TaskFilterRequest, error) {
	var req dto.TaskFilterRequest
	pageInt := 1
	pageSizeInt := 10
	var err error

	status := r.URL.Query().Get("status")
	createdFrom := r.URL.Query().Get("created_from")
	createdTo := r.URL.Query().Get("created_to")

	page := r.URL.Query().Get("page")
	if page != "" {
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			return dto.TaskFilterRequest{}, domain.ErrInvalidQueryParam
		}
	}
	pageSize := r.URL.Query().Get("page_size")
	if pageSize != "" {
		pageSizeInt, err = strconv.Atoi(pageSize)
		if err != nil {
			return dto.TaskFilterRequest{}, domain.ErrInvalidQueryParam
		}
	}

	if isInvalidPagination(pageInt, pageSizeInt) {
		return dto.TaskFilterRequest{}, domain.ErrInvalidQueryParam
	}

	req = dto.TaskFilterRequest{
		Status:      status,
		CreatedFrom: createdFrom,
		CreatedTo:   createdTo,
		Page:        pageInt,
		PageSize:    pageSizeInt,
	}

	return req, nil
}

// Конвертация в доменную модель фильтрации
func convTaskFilterRequest(req dto.TaskFilterRequest) (domain.TaskFilter, error) {
	filter := domain.TaskFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	if req.Status != "" {
		s := domain.ProgressStatus(req.Status)
		if !domain.IsValidStatus(s) {
			return domain.TaskFilter{}, domain.ErrInvalidQueryParam
		}
		filter.Status = &s
	}

	if req.CreatedFrom != "" {
		timeFrom, err := time.Parse(time.RFC3339, req.CreatedFrom)
		if err != nil {
			return domain.TaskFilter{}, domain.ErrInvalidQueryParam
		}
		timeFrom = timeFrom.UTC()
		filter.CreatedFrom = &timeFrom
	}

	if req.CreatedTo != "" {
		timeTo, err := time.Parse(time.RFC3339, req.CreatedTo)
		if err != nil {
			return domain.TaskFilter{}, domain.ErrInvalidQueryParam
		}
		timeTo = timeTo.UTC()
		filter.CreatedTo = &timeTo
	}

	return filter, nil
}
