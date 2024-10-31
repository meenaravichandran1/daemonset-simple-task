package main

import (
	"context"
	"github.com/harness/runner/delegateshell/client"
	"github.com/harness/runner/logger/gcplogger"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type Server struct {
	handler *Handler
}

func NewServer(handler *Handler) *Server {
	return &Server{handler: handler}
}

func (s *Server) StartServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", s.handler.HandleTasks)

	httpServer := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	logrus.Printf("Daemon server is running on port %s\n", port)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		logrus.Printf("Error starting server: %v\n", err)
	}
}

func main() {
	// TODO set it from env var
	managerClient := client.NewManagerClient("https://qa.harness.io", "px7xd_BFRCi-pfWPYXVjvw", "token",
		true, "")

	// TODO if remote logging is enabled in env var, set the bool from env var
	remoteLogger := gcplogger.NewGCPLogger(logrus.StandardLogger(), managerClient)

	// TODO set the bool from env var
	_, err := remoteLogger.StartGcpLogger(context.TODO())
	if err != nil {
		return
	}

	logrus.Infoln("Starting daemonset-simple-task...")
	port := os.Getenv("DAEMON_SERVER_PORT")
	if port == "" {
		logrus.Printf("Environment variable DAEMON_SERVER_PORT is not set. Cannot start server\n")
		return
	}

	handler := &Handler{port: port, tasks: make(map[string]chan bool), remoteLogger: remoteLogger}
	server := NewServer(handler)
	server.StartServer(":" + port)
}
