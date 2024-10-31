package main

import (
	"github.com/sirupsen/logrus"
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
			logrus.WithFields(logrus.Fields{"taskId": taskId}).Info(message)
		}
	}
}
