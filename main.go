package main

import (
	"context"
	"github.com/harness/runner/logger"
	"github.com/harness/runner/logger/remotelogger"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

const serviceName = "daemonset"

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

	logger.Printf("Daemon server is running on port %s\n", port)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		logger.Printf("Error starting server: %v\n", err)
	}
}

func main() {

	logger.SetReportCaller(true)
	logger.SetFormatter(&logrus.JSONFormatter{})

	isRemoteLoggingEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_REMOTE_LOGGING"))
	if err != nil {
		isRemoteLoggingEnabled = false
	}

	if isRemoteLoggingEnabled {
		startRemoteLogger()
	}

	logger.Infoln("Starting daemonset-simple-task...")
	port := os.Getenv("DAEMON_SERVER_PORT")
	if port == "" {
		logger.Printf("Environment variable DAEMON_SERVER_PORT is not set. Cannot start server\n")
		return
	}

	handler := &Handler{port: port, tasks: make(map[string]chan bool)}
	server := NewServer(handler)
	server.StartServer(":" + port)
}

func startRemoteLogger() {
	managerEndpoint := os.Getenv("DIAL_HOME_URL")
	if managerEndpoint == "" {
		logger.Println("Environment variable DIAL_HOME_URL is not set. Cannot publish logs to remote")
		return
	}
	runnerToken := os.Getenv("DIAL_HOME_TOKEN")
	if runnerToken == "" {
		logger.Println("Environment variable DIAL_HOME_TOKEN is not set. Cannot publish logs to remote")
		return
	}
	accountId := os.Getenv("ACCOUNT_ID")
	if accountId == "" {
		logger.Println("Environment variable ACCOUNT_ID is not set. Cannot publish logs to remote")
		return
	}
	insecure, err := strconv.ParseBool(os.Getenv("DIAL_HOME_INSECURE"))
	if err != nil {
		insecure = true
	}

	remotelogger.Start(context.Background(), accountId, managerEndpoint, runnerToken, serviceName, "daemonset-simple-task", true, insecure)
	logger.Infoln("Publishing daemon set logs to remote")
	logger.UpdateContextInHooks(map[string]string{"service": serviceName})
	return
}
