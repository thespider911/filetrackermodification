package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// startService - this checks if service is running then continue else start the service
func (app *application) startService() {
	app.logBufferMu.Lock()
	defer app.logBufferMu.Unlock()

	if app.isRunning {
		app.appendLog("Service is already running.\n")
		return
	}

	app.isRunning = true
	app.serviceStopper = make(chan struct{})

	// Start workerThread
	go func() {
		app.appendLog("Worker thread starting...\n")
		if err := app.workerThread(); err != nil {
			app.errorLog.Printf("Worker thread error: %v\n", err)
			app.appendLog(fmt.Sprintf("Worker thread error: %v\n", err))
		}
	}()

	// Start timerThread
	go func() {
		app.appendLog("Timer thread starting...\n")
		if err := app.timerThread(); err != nil {
			app.errorLog.Printf("Timer thread error: %v\n", err)
			app.appendLog(fmt.Sprintf("Timer thread error: %v\n", err))
		}
	}()

	app.appendLog("Service started.\n")

	// Trigger an immediate check of the directory
	go app.initialDirectoryCheck()
}

// stopService - this checks if service is running then continue stops if running
func (app *application) stopService() {
	app.logBufferMu.Lock()
	defer app.logBufferMu.Unlock()

	if !app.isRunning {
		app.appendLog("Service is already stopped.\n")
		return
	}
	app.isRunning = false

	if app.serviceStopper != nil {
		close(app.serviceStopper)
	}

	// Wait with timeout
	timeout := time.After(5 * time.Second)
	stoppedChan := make(chan struct{})

	go func() {
		// Wait for both threads to finish
		// This assumes workerThread and timerThread respect app.serviceStopper
		<-app.serviceStopper
		<-app.serviceStopper
		close(stoppedChan)
	}()

	select {
	case <-timeout:
		app.appendLog("Warning: Service stop timed out. Some operations may still be running.\n")
	case <-stoppedChan:
		app.appendLog("All threads stopped successfully.\n")
	}

	app.serviceStopper = nil
	app.appendLog("Service stopped.\n")
}

// initialDirectoryCheck -  check the directory if exists
func (app *application) initialDirectoryCheck() {
	app.appendLog("Starting initial directory check...\n")
	err := filepath.Walk(app.config.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			app.appendLog(fmt.Sprintf("Error accessing path %s: %v\n", path, err))
			return err
		}

		if !info.IsDir() {
			app.appendLog(fmt.Sprintf("Queueing file for check: %s\n", path))
			app.commandQueue <- Command{
				Type: "CHECK_DIRECTORY_FILES",
				Data: path,
			}
		}
		return nil
	})
	if err != nil {
		app.errorLog.Printf("Error in initial directory walk: %v\n", err)
		app.appendLog(fmt.Sprintf("Error in initial directory walk: %v\n", err))
	}
	app.appendLog("Initial directory check completed.\n")
}

// appendLog -
func (app *application) appendLog(text string) {
	select {
	case app.logChan <- text:
		// Message sent successfully
	default:
		// Channel is full, log to error log
		app.errorLog.Printf("Log channel full, couldn't log: %s", text)
	}
}

// updateLogs - to the fynne UI
func (app *application) updateLogs() {
	for text := range app.logChan {
		app.uiLogs.SetText(app.uiLogs.Text + text)
		app.uiLogs.CursorRow = len(app.uiLogs.Text)
		// Also log to the console for debugging
		//fmt.Print(text)
	}
}
