package main

import "net/http"

// routes - http requests
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", app.healthCheckHandler)
	mux.HandleFunc("/logs", app.logsHandler)
	mux.HandleFunc("/help", app.helpCommandsHandler)
	//http.HandleFunc("/query", commandQueryHandler)
	//http.HandleFunc("/execute", commandExecuteHandler)
	//
	mux.HandleFunc("/start", app.startServiceHandler)
	mux.HandleFunc("/stop", app.stopServiceHandler)

	return mux
}
