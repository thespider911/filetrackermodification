package main

import (
	"log"
	"net/http"
)

func route() {
	//api end points
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/health", handleCheckStatus)
	mux.HandleFunc("/v1/logs", handleLogs)
	mux.HandleFunc("/v1/help", handleHelpCommands)

	if err := http.ListenAndServe(":4000", mux); err != nil {
		log.Fatal(err)
	}
}
