package main

import (
	"fmt"
	"net/http"
	"strings"
)

// CommandInfo - to stores information about each command
type CommandInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
}

// --------------- COMMANDS --------------- //
var commandInfoMap = map[string]CommandInfo{
	"CHECK_DIRECTORY_FILE": {
		Name:        "CHECK_DIRECTORY_FILE",
		Description: "Checks file information for a given file path",
		Usage:       "/execute?command=CHECK_DIRECTORY_FILE&path=/path/to/file",
	},
	"CHECK_FILE_PERMISSION": {
		Name:        "CHECK_FILE_PERMISSION",
		Description: "Checks file permission of a given file",
		Usage:       "/execute?command=CHECK_FILE_PERMISSION&path=/path/to/file",
	},
	"CHECK_FILE_TYPE": {
		Name:        "CHECK_FILE_TYPE",
		Description: "Checks file type for a given path",
		Usage:       "/execute?command=CHECK_FILE_TYPE&path=/path/to/file",
	},
	"CHECK_IS_FILE_TYPE": {
		Name:        "CHECK_IS_FILE_TYPE",
		Description: "Checks documents for a given path",
		Usage:       "/execute?command=CHECK_IS_FILE_TYPE&path=/path/to/file",
	},
	"CHECK_FILE_DATES": {
		Name:        "CHECK_FILE_DATES",
		Description: "Checks file times",
		Usage:       "/execute?command=CHECK_FILE_DATES&path=/path/to/file",
	},
	"CHECK_IF_MODIFIED_FILE": {
		Name:        "CHECK_IF_MODIFIED_FILE",
		Description: "Checks file modified for a given path",
		Usage:       "/execute?command=CHECK_IF_MODIFIED_FILE&path=/path/to/file",
	},
}

// healthCheckHandler - check system health
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]string{
		"status":  "available",
		"service": "running",
	}

	if err := app.writeJSON(w, http.StatusOK, health, nil); err != nil {
		app.serverError(w, r, err)
		return
	}
}

// logsHandler - handle logs
func (app *application) logsHandler(w http.ResponseWriter, r *http.Request) {
	app.logBufferMu.RLock()
	defer app.logBufferMu.RUnlock()

	if err := app.writeJSON(w, http.StatusOK, app.logBuffer, nil); err != nil {
		app.serverError(w, r, err)
		return
	}
}

// helpCommandsHandler - handle help commands
func (app *application) helpCommandsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Help commands available"))
}

// ----------------- COMMANDS ----------------- //

// commandQueryHandler handles queries about commands
func (app *application) commandQueryHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	command := query.Get("command")

	if command == "" {
		if err := app.writeJSON(w, http.StatusOK, commandInfoMap, nil); err != nil {
			app.serverError(w, r, err)
			return
		}
		return
	}

	// convert command to uppercase for case-insensitive matching
	command = strings.ToUpper(command)

	if info, ok := commandInfoMap[command]; ok {
		if err := app.writeJSON(w, http.StatusOK, info, nil); err != nil {
			app.serverError(w, r, err)
			return
		}
	} else {
		app.notFound(w, r)
		return
	}
}

// commandExecuteHandler handles execution of commands
func (app *application) commandExecuteHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	command := query.Get("command")

	if command == "" {
		http.Error(w, "Command parameter is required", http.StatusBadRequest)
		return
	}

	// convert command to uppercase for case-insensitive matching
	command = strings.ToUpper(command)

	params := make(map[string]string)
	for key, values := range query {
		if key != "command" && len(values) > 0 {
			params[key] = values[0]
		}
	}

	result, err := app.service.CommandRunFile.ExecuteCommand(command, params)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusOK, result, nil); err != nil {
		app.serverError(w, r, err)
		return
	}
}

// ----------------- FOR UI SIDE ----------------- //
// startServiceHandler - start work and thread service if not running
func (app *application) startServiceHandler(w http.ResponseWriter, r *http.Request) {
	//if not running start service
	if !app.isRunning {
		app.isRunning = true
		app.serviceStopper = make(chan struct{})
		app.wg.Add(2)

		//start workerThread
		go func() {
			err := app.workerThread()
			if err != nil {
				app.serverError(w, r, err)
				return
			}
		}()

		//start timeThread
		go func() {
			err := app.timerThread()
			if err != nil {
				app.serverError(w, r, err)
				return
			}
		}()

		w.WriteHeader(http.StatusOK)
	} else {
		app.badRequest(w, r, fmt.Errorf("service is already stopped"))
	}
}

// startServiceHandler - stop work and thread service if running
func (app *application) stopServiceHandler(w http.ResponseWriter, r *http.Request) {
	//if running stop service
	if app.isRunning {
		app.isRunning = false
		close(app.serviceStopper)
		app.wg.Wait()

		w.WriteHeader(http.StatusOK)
	} else {
		app.badRequest(w, r, fmt.Errorf("service is not running"))
	}
}
