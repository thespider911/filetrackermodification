package main

import (
	"encoding/json"
	"net/http"
)

// writeJSON - marshal payload and return a nice JSON format
func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	//convert data to json
	js, err := app.JSON(data)
	if err != nil {
		return err
	}

	//range headers
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return err
	}

	return nil
}

// JSON - convert data to json
func (app *application) JSON(data interface{}) ([]byte, error) {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return nil, err
	}

	//new line
	js = append(js, '\n')
	return js, nil
}
