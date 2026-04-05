package transport

import (
	"context"
	"net/http"
	"task-tracker-1/internal/transport/handlers"
	"task-tracker-1/internal/transport/middleware"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(authHandler *handlers.AuthHandler, taskHandler *handlers.TaskHandler, tokenValidator middleware.AccessTokenValidator, rateLimiter *middleware.RateLimiter, addr string) *Server {
	mux := http.NewServeMux()
	//Auth-service
	//Открытые роуты
	// Инициализируем Rate Limiter Middleware
	rateLimiterMiddleware := middleware.RateLimitMiddleware(rateLimiter)
	// Оборачиаем нужные роуты
	mux.Handle("POST /login", rateLimiterMiddleware(http.HandlerFunc(authHandler.Login)))
	mux.Handle("POST /refresh", rateLimiterMiddleware(http.HandlerFunc(authHandler.Refresh)))
	mux.Handle("POST /register", rateLimiterMiddleware(http.HandlerFunc(authHandler.Register)))
	// Не трогаем
	mux.HandleFunc("GET /validate", authHandler.ValidateToken)

	handler := middleware.RequestIDMiddleware(middleware.RequestLoggerMiddleware(mux))
	//Task-service
	taskMux := http.NewServeMux()
	taskMux.HandleFunc("POST /tasks", taskHandler.CreateTask)
	taskMux.HandleFunc("GET /tasks", taskHandler.GetListTasks)
	taskMux.HandleFunc("PUT /tasks/{id}", taskHandler.UpdateTask)
	taskMux.HandleFunc("DELETE /tasks/{id}", taskHandler.DeleteTask)
	taskMux.HandleFunc("GET /tasks/{id}/history", taskHandler.GetTaskHistory)

	//Auth-Middleware
	authMiddleware := middleware.ValidateTokenMiddleware(tokenValidator)
	mux.Handle("/tasks", authMiddleware(taskMux))
	mux.Handle("/tasks/", authMiddleware(taskMux))
	//Защищенный logout
	authMux := http.NewServeMux()
	authMux.HandleFunc("POST /logout", authHandler.Logout)
	mux.Handle("/logout", authMiddleware(authMux))

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		httpServer: srv,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) GracefulShutdown(ctx context.Context) error {

	return s.httpServer.Shutdown(ctx)
}
