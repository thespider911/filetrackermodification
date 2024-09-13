package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

/*
	workerThread

- this is to keep running continuous unless stopped
- the thread listens to file logs the files info
- it loops the checking directory files command
*/
func (app *application) workerThread() error {
	defer app.wg.Done()
	for {
		select {
		case cmd := <-app.commandQueue:
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

						//send to api the file info
						err = app.sendToAPI(*fileInfo)
						if err != nil {
							return err
						}
					}
				}
			default:
				return errors.New(fmt.Sprintf("unknown command type: %s\n", cmd.Type))
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
	defer app.wg.Done()
	//check this thread every minute
	ticker := time.NewTicker(time.Duration(app.config.CheckInterval) * time.Second)
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
