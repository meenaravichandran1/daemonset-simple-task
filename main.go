package main

import (
	"context"
	"github.com/harness/runner/delegateshell/client"
	"github.com/harness/runner/logger/gcplogger"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
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

	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	isRemoteLoggingEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_REMOTE_LOGGING"))
	if err != nil {
		isRemoteLoggingEnabled = false
	}

	var remoteLogger *gcplogger.GCPLogger
	if isRemoteLoggingEnabled {
		remoteLogger = startRemoteLogger()
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

func startRemoteLogger() *gcplogger.GCPLogger {
	managerEndpoint := os.Getenv("MANAGER_HOST_AND_PORT")
	if managerEndpoint == "" {
		logrus.Println("Environment variable MANAGER_HOST_AND_PORT is not set. Cannot publish logs to remote")
		return nil
	}
	delegateToken := os.Getenv("DELEGATE_TOKEN")
	if delegateToken == "" {
		logrus.Println("Environment variable DELEGATE_TOKEN is not set. Cannot publish logs to remote")
		return nil
	}
	accountId := os.Getenv("ACCOUNT_ID")
	if accountId == "" {
		logrus.Println("Environment variable ACCOUNT_ID is not set. Cannot publish logs to remote")
		return nil
	}
	insecure, err := strconv.ParseBool(os.Getenv("SERVER_INSECURE"))
	if err != nil {
		insecure = true
	}

	managerClient := client.NewManagerClient(managerEndpoint, accountId, delegateToken,
		insecure, "")

	additionalFields := map[string]string{"service": "daemonset-simple-task"}
	remoteLogger := gcplogger.NewGCPLogger(logrus.StandardLogger(), additionalFields, managerClient)

	_, err = remoteLogger.StartGcpLogger(context.TODO())
	if err != nil {
		return nil
	}
	logrus.Infoln("Publishing daemon set logs to remote")
	return remoteLogger
}
