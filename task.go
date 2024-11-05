package main

import (
	"github.com/harness/runner/logger"
	"time"
)

func startTask(taskId string, quit chan bool) {
	message := "Current daemon set is daemon-simple-task...!!!"
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			logger.WithFields(map[string]interface{}{"taskId": taskId}).Info(message)
		}
	}
}
