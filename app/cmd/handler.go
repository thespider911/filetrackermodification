package main

import (
	"fmt"
	"net/http"
)

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
