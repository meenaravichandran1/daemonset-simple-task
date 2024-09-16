package main

import (
	"fmt"
	"os"
	"time"
)

type Params struct {
	Message string `json:"message"`
}

func startTask(taskId string, params Params, quit chan bool) {
	message := "My current version is 0.0.1"
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			writeToFile(fmt.Sprintf("%v [taskId=%s]", message, taskId))
		}
	}
}

func writeToFile(message string) {
	file, err := os.OpenFile("output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening or creating the file: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(message + "\n"); err != nil {
		fmt.Printf("Error writing to the file: %v\n", err)
	}
}
