package main

import "net/http"

// routes - http requests
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", app.healthCheckHandler) //check system health
	mux.HandleFunc("/logs", app.logsHandler)          //log result

	mux.HandleFunc("/help", app.commandQueryHandler)      //display commands
	mux.HandleFunc("/execute", app.commandExecuteHandler) // execute commands

	mux.HandleFunc("/start", app.startServiceHandler) //start service
	mux.HandleFunc("/stop", app.stopServiceHandler)   //stop service

	return mux
}
