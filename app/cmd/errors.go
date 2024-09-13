package main

import (
	"net/http"
	"strings"
)

// errorMessage - log error as readable text
func (app *application) errorMessage(w http.ResponseWriter, r *http.Request, status int, message string, headers http.Header) {
	http.Error(w, http.StatusText(status), status)

	message = strings.ToUpper(message[:1]) + message[1:]

	err := app.writeJSON(w, status, map[string]string{"Error": message}, headers)
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// serverError -
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.errorLog.Println(err)

	message := "Oop! something went wrong. Try again."
	app.errorMessage(w, r, http.StatusInternalServerError, message, nil)
}

// notFound -
func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorMessage(w, r, http.StatusNotFound, message, nil)
}

// badRequest
func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	app.errorMessage(w, r, http.StatusBadRequest, err.Error(), nil)
}
