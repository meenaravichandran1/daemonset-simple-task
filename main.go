package main

import (
	"context"
	"fmt"
	"github.com/harness/runner/delegateshell/client"
	"github.com/meenaravichandran1/runner-logger/logger"
	"github.com/meenaravichandran1/runner-logger/runnerlogs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	handler      *Handler
	httpServer   *http.Server
	remoteLogger *runnerlogs.RemoteLogger
}

func NewServer(handler *Handler, remoteLogger *runnerlogs.RemoteLogger) *Server {
	return &Server{handler: handler, remoteLogger: remoteLogger}
}

func (s *Server) StartServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", s.handler.HandleTasks)

	// Define http.Server to have control over shutdown
	s.httpServer = &http.Server{
		Addr:    port,
		Handler: mux,
	}

	// Run the server in a goroutine so it doesn't block and we can listen for shutdown signals
	go func() {
		logger.Printf("Daemon server is running on port %s\n", port)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Printf("Error starting server: %v\n", err)
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Perform any additional cleanup here, like closing client connections
	logger.Println("Shutting down server...")
	err := s.remoteLogger.StopRemoteLogger()
	if err != nil {
		return err
	}
	return s.httpServer.Shutdown(ctx)
}

func main() {
	runnerlogs.StartLocalLogger()

	// TODO set it from env var
	managerClient := client.NewManagerClient("https://qa.harness.io", "px7xd_BFRCi-pfWPYXVjvw", "b219e91636d72f254b2349cbb69e1a90",
		true, "")

	// TODO set the bool from env var
	remoteLogger := runnerlogs.InitRemoteLogger(context.TODO(), true, managerClient)

	port := os.Getenv("DAEMON_SERVER_PORT")
	if port == "" {
		logger.Printf("Environment variable DAEMON_SERVER_PORT is not set. Cannot start server\n")
		return
	}

	handler := &Handler{port: port, tasks: make(map[string]chan bool)}
	server := NewServer(handler, remoteLogger)
	server.StartServer(":" + port)

	// Setup signal catching
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-stop

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Perform server shutdown and cleanup
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Error during shutdown: %v\n", err)
	}
}
