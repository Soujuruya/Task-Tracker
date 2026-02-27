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

func NewServer(authHandler *handlers.AuthHandler, taskHandler *handlers.TaskHandler, addr string) *Server {
	mux := http.NewServeMux()
	//Auth-service
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("/validate", authHandler.ValidateToken)
	//Task-service
	taskMux := http.NewServeMux()
	taskMux.HandleFunc("POST /tasks", taskHandler.CreateTask)
	taskMux.HandleFunc("GET /tasks", taskHandler.GetListTasks)
	taskMux.HandleFunc("PUT /tasks/{id}", taskHandler.UpdateTask)
	taskMux.HandleFunc("DELETE /tasks/{id}", taskHandler.DeleteTask)

	//Auth-Middleware
	authServiceURL := "http://" + addr
	taskMiddleware := middleware.ValidateTokenMiddleware(authServiceURL)
	mux.Handle("/tasks", taskMiddleware(taskMux))
	mux.Handle("/tasks/", taskMiddleware(taskMux))

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
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
