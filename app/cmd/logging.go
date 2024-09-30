package main

import (
	"github.com/thespider911/filetrackermodification/app/internal/helpers"
	"github.com/thespider911/filetrackermodification/app/internal/service/filetrack"
	"log"
	"os"
)

// logging - create if not exists a file_tracking.log file to records result logs
func (app *application) logging() {
	var err error
	app.logFile, err = os.OpenFile("file_tracking.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		app.errorLog.Fatal(err)
		return
	}

	log.SetOutput(app.logFile)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

// logFileInfo - log to file and in-memory
func (app *application) logFileInfo(fileInfo filetrack.FileInfo) error {
	// update file result to readable info
	fileInfo.ModifiedTime = helpers.ToHumanReadableTime(fileInfo.ModifiedTime)
	fileInfo.AccessedTime = helpers.ToHumanReadableTime(fileInfo.AccessedTime)
	fileInfo.ChangedTime = helpers.ToHumanReadableTime(fileInfo.ChangedTime)
	fileInfo.FileSize = helpers.ToHumanReadableFileSize(fileInfo.FileSize)

	//convert file info to json
	js, err := app.JSON(fileInfo)
	if err != nil {
		return err
	}

	//print to file log
	log.Print(string(js))

	//mutex to ensure safe thread access
	app.logBufferMu.Lock()
	defer app.logBufferMu.Unlock()

	if len(app.logBuffer) >= 1000 {
		app.logBuffer = app.logBuffer[1:]
	}
	app.logBuffer = append(app.logBuffer, fileInfo)

	return nil
}
