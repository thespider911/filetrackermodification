package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"
)

/*
	workerThread

- this is to keep running continuous unless stopped
- the thread listens to file logs the files info
- it loops the checking directory files command
*/
func (app *application) workerThread() error {
	defer func() {
		app.wg.Done()
		atomic.AddInt32(&app.wgCount, -1)
		app.appendLog("Worker thread stopped\n")
	}()

	for {
		select {
		case cmd, ok := <-app.commandQueue:
			if !ok {
				return nil
			}
			switch cmd.Type {
			// check files command and print file info
			case "CHECK_DIRECTORY_FILES":
				if filePath, ok := cmd.Data.(string); ok {
					fileInfo, err := app.service.FileTracker.FetchFilesInfo(filePath)
					if err != nil {
						return err
					}

					if fileInfo != nil {
						// print the result file information
						err = app.logFileInfo(*fileInfo)
						if err != nil {
							return err
						}

						// Update UI logs
						jsonData, err := app.JSON(fileInfo)
						if err != nil {
							app.errorLog.Printf("Error marshalling file info to JSON: %v\n", err)
						}

						// Log the JSON string
						app.appendLog(fmt.Sprintf("File Info:\n%s", string(jsonData)))

						//send to api the file info
						err = app.sendToAPI(*fileInfo)
						if err != nil {
							// if the api is not running
							if errors.Is(err, syscall.ECONNREFUSED) {
								app.errorLog.Println("api service not running")
							} else {
								app.errorLog.Printf("Error sending to API: %v\n", err)
							}
						}
					}
				}
			default:
				app.errorLog.Printf("Unknown command type: %s\n", cmd.Type)
			}
		case <-app.serviceStopper:
			return nil
		}
	}
}

/*
*
timeThread - this runs every minute checking all files in the specified directory
- It then calls the commandQueue looping through
*/
func (app *application) timerThread() error {
	defer func() {
		app.wg.Done()
		atomic.AddInt32(&app.wgCount, -1)
		app.appendLog("Timer thread stopped\n")
	}()

	// Ensure CheckInterval is positive
	checkInterval := time.Duration(app.config.CheckInterval) * time.Second
	if checkInterval <= 0 {
		checkInterval = time.Minute // Default to 1 minute if not set or invalid
		app.appendLog(fmt.Sprintf("Warning: Invalid CheckInterval (%d). Using default of 1 minute.\n", app.config.CheckInterval))
	}

	//check this thread every minute
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := filepath.Walk(app.config.Directory, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() {
					app.commandQueue <- Command{
						Type: "CHECK_DIRECTORY_FILES",
						Data: path,
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		case <-app.serviceStopper:
			return nil
		}
	}
}
