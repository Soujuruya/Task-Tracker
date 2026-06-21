package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/pkg/hasher"
	"task-tracker-1/internal/repository/auditlog"
	"task-tracker-1/internal/repository/refresh_token"
	"task-tracker-1/internal/repository/task"
	"task-tracker-1/internal/repository/user"
	"task-tracker-1/internal/service"
	"task-tracker-1/internal/transport/handlers"
	"task-tracker-1/internal/transport/middleware"
)

type plainHasher struct{}

func (plainHasher) Hash(ctx context.Context, value string) (string, error) {
	return "hash:" + value, nil
}

func (plainHasher) Compare(ctx context.Context, hash, password string) error {
	if hash != "hash:"+password {
		return domain.ErrInvalidCredentials
	}
	return nil
}

type testHTTPApp struct {
	t      *testing.T
	server *httptest.Server
	client *http.Client
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type registerResponse struct {
	UserID string `json:"user_id"`
}

type taskResponse struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	ProgressStatus string `json:"progress_status"`
	CreatedAt      string `json:"created_at"`
}

type listTasksResponse struct {
	Tasks      []taskResponse `json:"tasks"`
	Pagination struct {
		Total      int `json:"total"`
		Page       int `json:"page"`
		PageSize   int `json:"page_size"`
		TotalPages int `json:"total_pages"`
	} `json:"pagination"`
}

func newTestHTTPApp(t *testing.T) *testHTTPApp {
	t.Helper()

	userRepo := user.NewMemoryUserRepository()
	taskRepo := task.NewMemoryTaskRepository()
	auditLogRepo := auditlog.NewAuditLogRepository()
	refreshRepo := refresh_token.NewRefreshRepository()

	authService := service.NewAuthService(userRepo, plainHasher{})
	taskService := service.NewTaskService(taskRepo, auditLogRepo)
	tokenService := service.NewTokenService(
		[]byte("test-secret"),
		15*time.Minute,
		24*time.Hour,
		refreshRepo,
		hasher.NewSha256Hash(),
	)

	rateLimiter, err := middleware.NewRateLimiter(middleware.RateLimitConfig{
		MaxRequests: 1000,
		WindowSize:  time.Minute,
	})
	if err != nil {
		t.Fatalf("new rate limiter: %v", err)
	}

	server := NewServer(
		handlers.NewAuthHandler(authService, tokenService),
		handlers.NewTaskHandler(taskService),
		tokenService,
		rateLimiter,
		":0",
	)

	testServer := httptest.NewServer(server.httpServer.Handler)
	t.Cleanup(testServer.Close)

	return &testHTTPApp{
		t:      t,
		server: testServer,
		client: testServer.Client(),
	}
}

func (app *testHTTPApp) request(method, path, token string, body any) *http.Response {
	app.t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			app.t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, app.server.URL+path, reader)
	if err != nil {
		app.t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := app.client.Do(req)
	if err != nil {
		app.t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func (app *testHTTPApp) rawRequest(method, path, authHeader, body string) *http.Response {
	app.t.Helper()

	req, err := http.NewRequest(method, app.server.URL+path, strings.NewReader(body))
	if err != nil {
		app.t.Fatalf("new raw request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := app.client.Do(req)
	if err != nil {
		app.t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func decodeBody[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()

	var value T
	if err := json.NewDecoder(resp.Body).Decode(&value); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return value
}

func closeBody(t *testing.T, resp *http.Response) {
	t.Helper()
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close response body: %v", err)
	}
}

func assertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		t.Fatalf("status = %d, want %d, body = %q", resp.StatusCode, want, string(body))
	}
}

func (app *testHTTPApp) register(username, password string) registerResponse {
	app.t.Helper()
	resp := app.request(http.MethodPost, "/register", "", map[string]string{
		"username": username,
		"password": password,
	})
	assertStatus(app.t, resp, http.StatusCreated)
	return decodeBody[registerResponse](app.t, resp)
}

func (app *testHTTPApp) login(username, password string) loginResponse {
	app.t.Helper()
	resp := app.request(http.MethodPost, "/login", "", map[string]string{
		"username": username,
		"password": password,
	})
	assertStatus(app.t, resp, http.StatusOK)
	return decodeBody[loginResponse](app.t, resp)
}

func (app *testHTTPApp) createTask(token, title, description, status string) taskResponse {
	app.t.Helper()
	resp := app.request(http.MethodPost, "/tasks", token, map[string]string{
		"title":           title,
		"description":     description,
		"progress_status": status,
	})
	assertStatus(app.t, resp, http.StatusCreated)
	return decodeBody[taskResponse](app.t, resp)
}

func TestAuthEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T, app *testHTTPApp)
	}{
		{
			name: "register creates user",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.register("demo", "password123")
				if resp.UserID == "" {
					t.Fatal("empty user_id")
				}
			},
		},
		{
			name: "register rejects invalid json",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.rawRequest(http.MethodPost, "/register", "", "{")
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "register rejects empty username",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodPost, "/register", "", map[string]string{"password": "password123"})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "register rejects empty password",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodPost, "/register", "", map[string]string{"username": "demo"})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "register rejects duplicate username",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				resp := app.request(http.MethodPost, "/register", "", map[string]string{
					"username": "demo",
					"password": "password123",
				})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusConflict)
			},
		},
		{
			name: "login returns token pair",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				resp := app.login("demo", "password123")
				if resp.AccessToken == "" || resp.RefreshToken == "" {
					t.Fatal("empty token pair")
				}
			},
		},
		{
			name: "login rejects invalid json",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.rawRequest(http.MethodPost, "/login", "", "{")
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "login rejects missing username",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodPost, "/login", "", map[string]string{"password": "password123"})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "login rejects missing password",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodPost, "/login", "", map[string]string{"username": "demo"})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "login rejects unknown user",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodPost, "/login", "", map[string]string{
					"username": "missing",
					"password": "password123",
				})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "login rejects wrong password",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				resp := app.request(http.MethodPost, "/login", "", map[string]string{
					"username": "demo",
					"password": "wrong-password",
				})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "validate accepts access token",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodGet, "/validate", tokens.AccessToken, nil)
				assertStatus(t, resp, http.StatusOK)
				body := decodeBody[map[string]any](t, resp)
				if body["valid"] != true {
					t.Fatalf("valid = %v, want true", body["valid"])
				}
			},
		},
		{
			name: "validate rejects missing token",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodGet, "/validate", "", nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "validate rejects invalid token",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodGet, "/validate", "not-a-jwt", nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "validate rejects unsupported auth scheme",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.rawRequest(http.MethodGet, "/validate", "Basic abc", "")
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "refresh returns new token pair",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodPost, "/refresh", "", map[string]string{
					"refresh_token": tokens.RefreshToken,
				})
				assertStatus(t, resp, http.StatusOK)
				refreshed := decodeBody[loginResponse](t, resp)
				if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
					t.Fatal("empty refreshed token pair")
				}
				if refreshed.RefreshToken == tokens.RefreshToken {
					t.Fatal("refresh token was not rotated")
				}
			},
		},
		{
			name: "refresh rejects invalid json",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.rawRequest(http.MethodPost, "/refresh", "", "{")
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "refresh rejects missing token",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodPost, "/refresh", "", map[string]string{})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "refresh rejects unknown token",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodPost, "/refresh", "", map[string]string{
					"refresh_token": "unknown",
				})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "refresh rejects reused rotated token",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodPost, "/refresh", "", map[string]string{
					"refresh_token": tokens.RefreshToken,
				})
				assertStatus(t, resp, http.StatusOK)
				closeBody(t, resp)

				resp = app.request(http.MethodPost, "/refresh", "", map[string]string{
					"refresh_token": tokens.RefreshToken,
				})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "logout revokes refresh token",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodPost, "/logout", tokens.AccessToken, nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusNoContent)

				resp = app.request(http.MethodPost, "/refresh", "", map[string]string{
					"refresh_token": tokens.RefreshToken,
				})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "logout rejects missing token",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodPost, "/logout", "", nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "logout rejects invalid token",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodPost, "/logout", "invalid", nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.run(t, newTestHTTPApp(t))
		})
	}
}

func TestTaskEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T, app *testHTTPApp)
	}{
		{
			name: "create task with explicit status",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				task := app.createTask(tokens.AccessToken, "Task", "Description", "todo")
				if task.ID == "" || task.ProgressStatus != "todo" {
					t.Fatalf("unexpected task: %+v", task)
				}
			},
		},
		{
			name: "create task uses default todo status",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				task := app.createTask(tokens.AccessToken, "Task", "Description", "")
				if task.ProgressStatus != "todo" {
					t.Fatalf("status = %q, want todo", task.ProgressStatus)
				}
			},
		},
		{
			name: "create task accepts blocked status",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				task := app.createTask(tokens.AccessToken, "Task", "Description", "blocked")
				if task.ProgressStatus != "blocked" {
					t.Fatalf("status = %q, want blocked", task.ProgressStatus)
				}
			},
		},
		{
			name: "create task rejects missing auth",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodPost, "/tasks", "", map[string]string{"title": "Task"})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "create task rejects invalid token",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodPost, "/tasks", "invalid", map[string]string{"title": "Task"})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "create task rejects invalid json",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.rawRequest(http.MethodPost, "/tasks", "Bearer "+tokens.AccessToken, "{")
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "create task rejects empty title",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodPost, "/tasks", tokens.AccessToken, map[string]string{"description": "Description"})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "create task rejects invalid status",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodPost, "/tasks", tokens.AccessToken, map[string]string{
					"title":           "Task",
					"progress_status": "invalid",
				})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "list tasks returns pagination",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				app.createTask(tokens.AccessToken, "First", "Description", "todo")
				app.createTask(tokens.AccessToken, "Second", "Description", "blocked")

				resp := app.request(http.MethodGet, "/tasks", tokens.AccessToken, nil)
				assertStatus(t, resp, http.StatusOK)
				body := decodeBody[listTasksResponse](t, resp)
				if body.Pagination.Total != 2 || len(body.Tasks) != 2 {
					t.Fatalf("unexpected list response: %+v", body)
				}
			},
		},
		{
			name: "list tasks filters by status",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				app.createTask(tokens.AccessToken, "First", "Description", "todo")
				app.createTask(tokens.AccessToken, "Second", "Description", "blocked")

				resp := app.request(http.MethodGet, "/tasks?status=blocked", tokens.AccessToken, nil)
				assertStatus(t, resp, http.StatusOK)
				body := decodeBody[listTasksResponse](t, resp)
				if body.Pagination.Total != 1 || body.Tasks[0].ProgressStatus != "blocked" {
					t.Fatalf("unexpected filtered response: %+v", body)
				}
			},
		},
		{
			name: "list tasks filters by created_from",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				app.createTask(tokens.AccessToken, "Task", "Description", "todo")
				createdFrom := time.Now().Add(-time.Minute).UTC().Format(time.RFC3339)

				resp := app.request(http.MethodGet, "/tasks?created_from="+createdFrom, tokens.AccessToken, nil)
				assertStatus(t, resp, http.StatusOK)
				body := decodeBody[listTasksResponse](t, resp)
				if body.Pagination.Total != 1 {
					t.Fatalf("total = %d, want 1", body.Pagination.Total)
				}
			},
		},
		{
			name: "list tasks filters by created_to",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				app.createTask(tokens.AccessToken, "Task", "Description", "todo")
				createdTo := time.Now().Add(time.Minute).UTC().Format(time.RFC3339)

				resp := app.request(http.MethodGet, "/tasks?created_to="+createdTo, tokens.AccessToken, nil)
				assertStatus(t, resp, http.StatusOK)
				body := decodeBody[listTasksResponse](t, resp)
				if body.Pagination.Total != 1 {
					t.Fatalf("total = %d, want 1", body.Pagination.Total)
				}
			},
		},
		{
			name: "list tasks applies page size",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				app.createTask(tokens.AccessToken, "First", "Description", "todo")
				app.createTask(tokens.AccessToken, "Second", "Description", "todo")

				resp := app.request(http.MethodGet, "/tasks?page=1&page_size=1", tokens.AccessToken, nil)
				assertStatus(t, resp, http.StatusOK)
				body := decodeBody[listTasksResponse](t, resp)
				if body.Pagination.Total != 2 || len(body.Tasks) != 1 || body.Pagination.TotalPages != 2 {
					t.Fatalf("unexpected paginated response: %+v", body)
				}
			},
		},
		{
			name: "list tasks returns empty page outside range",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				app.createTask(tokens.AccessToken, "Task", "Description", "todo")

				resp := app.request(http.MethodGet, "/tasks?page=3&page_size=10", tokens.AccessToken, nil)
				assertStatus(t, resp, http.StatusOK)
				body := decodeBody[listTasksResponse](t, resp)
				if body.Pagination.Total != 1 || len(body.Tasks) != 0 {
					t.Fatalf("unexpected out-of-range response: %+v", body)
				}
			},
		},
		{
			name: "list tasks rejects missing auth",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodGet, "/tasks", "", nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "list tasks rejects invalid status filter",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodGet, "/tasks?status=invalid", tokens.AccessToken, nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "list tasks rejects invalid created_from",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodGet, "/tasks?created_from=bad-date", tokens.AccessToken, nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "list tasks rejects invalid created_to",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodGet, "/tasks?created_to=bad-date", tokens.AccessToken, nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "list tasks rejects invalid page",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodGet, "/tasks?page=0", tokens.AccessToken, nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "list tasks rejects invalid page size",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodGet, "/tasks?page_size=101", tokens.AccessToken, nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "update task title",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				task := app.createTask(tokens.AccessToken, "Old", "Description", "todo")
				resp := app.request(http.MethodPut, "/tasks/"+task.ID, tokens.AccessToken, map[string]string{
					"title": "New",
				})
				assertStatus(t, resp, http.StatusOK)
				updated := decodeBody[taskResponse](t, resp)
				if updated.Title != "New" {
					t.Fatalf("title = %q, want New", updated.Title)
				}
			},
		},
		{
			name: "update task description",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				task := app.createTask(tokens.AccessToken, "Task", "Old", "todo")
				resp := app.request(http.MethodPut, "/tasks/"+task.ID, tokens.AccessToken, map[string]string{
					"description": "New",
				})
				assertStatus(t, resp, http.StatusOK)
				updated := decodeBody[taskResponse](t, resp)
				if updated.Description != "New" {
					t.Fatalf("description = %q, want New", updated.Description)
				}
			},
		},
		{
			name: "update task status",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				task := app.createTask(tokens.AccessToken, "Task", "Description", "todo")
				resp := app.request(http.MethodPut, "/tasks/"+task.ID, tokens.AccessToken, map[string]string{
					"progress_status": "in_progress",
				})
				assertStatus(t, resp, http.StatusOK)
				updated := decodeBody[taskResponse](t, resp)
				if updated.ProgressStatus != "in_progress" {
					t.Fatalf("status = %q, want in_progress", updated.ProgressStatus)
				}
			},
		},
		{
			name: "update task rejects missing auth",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodPut, "/tasks/task-id", "", map[string]string{"title": "New"})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "update task rejects invalid json",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				task := app.createTask(tokens.AccessToken, "Task", "Description", "todo")
				resp := app.rawRequest(http.MethodPut, "/tasks/"+task.ID, "Bearer "+tokens.AccessToken, "{")
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "update task rejects empty title",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				task := app.createTask(tokens.AccessToken, "Task", "Description", "todo")
				resp := app.request(http.MethodPut, "/tasks/"+task.ID, tokens.AccessToken, map[string]string{"title": ""})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "update task rejects invalid transition from done",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				task := app.createTask(tokens.AccessToken, "Task", "Description", "todo")
				resp := app.request(http.MethodPut, "/tasks/"+task.ID, tokens.AccessToken, map[string]string{"progress_status": "in_progress"})
				assertStatus(t, resp, http.StatusOK)
				closeBody(t, resp)
				resp = app.request(http.MethodPut, "/tasks/"+task.ID, tokens.AccessToken, map[string]string{"progress_status": "done"})
				assertStatus(t, resp, http.StatusOK)
				closeBody(t, resp)

				resp = app.request(http.MethodPut, "/tasks/"+task.ID, tokens.AccessToken, map[string]string{"progress_status": "todo"})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusConflict)
			},
		},
		{
			name: "update task rejects not found task",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodPut, "/tasks/missing", tokens.AccessToken, map[string]string{"title": "New"})
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusNotFound)
			},
		},
		{
			name: "delete task removes existing task",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				task := app.createTask(tokens.AccessToken, "Task", "Description", "todo")
				resp := app.request(http.MethodDelete, "/tasks/"+task.ID, tokens.AccessToken, nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusNoContent)
			},
		},
		{
			name: "delete task rejects missing auth",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodDelete, "/tasks/task-id", "", nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "delete task rejects not found task",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodDelete, "/tasks/missing", tokens.AccessToken, nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusNotFound)
			},
		},
		{
			name: "delete task rejects done task",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				task := app.createTask(tokens.AccessToken, "Task", "Description", "todo")
				resp := app.request(http.MethodPut, "/tasks/"+task.ID, tokens.AccessToken, map[string]string{"progress_status": "in_progress"})
				assertStatus(t, resp, http.StatusOK)
				closeBody(t, resp)
				resp = app.request(http.MethodPut, "/tasks/"+task.ID, tokens.AccessToken, map[string]string{"progress_status": "done"})
				assertStatus(t, resp, http.StatusOK)
				closeBody(t, resp)

				resp = app.request(http.MethodDelete, "/tasks/"+task.ID, tokens.AccessToken, nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusBadRequest)
			},
		},
		{
			name: "history returns task changes",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				task := app.createTask(tokens.AccessToken, "Old", "Description", "todo")
				resp := app.request(http.MethodPut, "/tasks/"+task.ID, tokens.AccessToken, map[string]string{"title": "New"})
				assertStatus(t, resp, http.StatusOK)
				closeBody(t, resp)

				resp = app.request(http.MethodGet, "/tasks/"+task.ID+"/history", tokens.AccessToken, nil)
				assertStatus(t, resp, http.StatusOK)
				entries := decodeBody[[]map[string]any](t, resp)
				if len(entries) == 0 {
					t.Fatal("empty history")
				}
			},
		},
		{
			name: "history rejects missing auth",
			run: func(t *testing.T, app *testHTTPApp) {
				resp := app.request(http.MethodGet, "/tasks/task-id/history", "", nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusUnauthorized)
			},
		},
		{
			name: "history rejects not found task",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("demo", "password123")
				tokens := app.login("demo", "password123")
				resp := app.request(http.MethodGet, "/tasks/missing/history", tokens.AccessToken, nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusNotFound)
			},
		},
		{
			name: "history hides another user's task",
			run: func(t *testing.T, app *testHTTPApp) {
				app.register("owner", "password123")
				ownerTokens := app.login("owner", "password123")
				task := app.createTask(ownerTokens.AccessToken, "Task", "Description", "todo")

				app.register("other", "password123")
				otherTokens := app.login("other", "password123")
				resp := app.request(http.MethodGet, "/tasks/"+task.ID+"/history", otherTokens.AccessToken, nil)
				defer closeBody(t, resp)
				assertStatus(t, resp, http.StatusNotFound)
			},
		},
	}

	if len(tests) < 30 {
		t.Fatalf("task endpoint test table is too small: %d", len(tests))
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.run(t, newTestHTTPApp(t))
		})
	}
}

func TestUnsupportedRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "unknown route", method: http.MethodGet, path: "/unknown"},
		{name: "unsupported register method", method: http.MethodGet, path: "/register"},
		{name: "unsupported task method", method: http.MethodPatch, path: "/tasks/task-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := newTestHTTPApp(t)
			resp := app.request(tt.method, tt.path, "", nil)
			defer closeBody(t, resp)
			if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("%s %s status = %d", tt.method, tt.path, resp.StatusCode)
			}
		})
	}
}

func TestRateLimitedAuthEndpoints(t *testing.T) {
	t.Parallel()

	userRepo := user.NewMemoryUserRepository()
	taskRepo := task.NewMemoryTaskRepository()
	auditLogRepo := auditlog.NewAuditLogRepository()
	refreshRepo := refresh_token.NewRefreshRepository()
	authService := service.NewAuthService(userRepo, plainHasher{})
	taskService := service.NewTaskService(taskRepo, auditLogRepo)
	tokenService := service.NewTokenService([]byte("test-secret"), time.Minute, time.Hour, refreshRepo, hasher.NewSha256Hash())
	rateLimiter, err := middleware.NewRateLimiter(middleware.RateLimitConfig{
		MaxRequests: 1,
		WindowSize:  time.Minute,
	})
	if err != nil {
		t.Fatalf("new rate limiter: %v", err)
	}
	server := NewServer(
		handlers.NewAuthHandler(authService, tokenService),
		handlers.NewTaskHandler(taskService),
		tokenService,
		rateLimiter,
		":0",
	)
	testServer := httptest.NewServer(server.httpServer.Handler)
	t.Cleanup(testServer.Close)
	app := &testHTTPApp{t: t, server: testServer, client: testServer.Client()}

	resp := app.request(http.MethodPost, "/login", "", map[string]string{
		"username": "demo",
		"password": "password123",
	})
	closeBody(t, resp)

	resp = app.request(http.MethodPost, "/login", "", map[string]string{
		"username": "demo",
		"password": "password123",
	})
	defer closeBody(t, resp)
	assertStatus(t, resp, http.StatusTooManyRequests)
}
