package main

import (
	"bytes"
	"fmt"
	"github.com/thespider911/filetrackermodification/app/internal/service/filetrack"
	"net/http"
)

// sentToApi - convert file into to json then send as response to api endpoint that it has access
func (app *application) sendToAPI(info filetrack.FileInfo) error {
	//file info to json
	jsonData, err := app.JSON(info)
	if err != nil {
		return fmt.Errorf("error converting file info to JSON: %w", err)
	}

	// use httpClient to send a post response to api endpoint
	resp, err := app.httpClient.Post(app.config.APIEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}
