package main

import (
	"encoding/json"
	"github.com/thespider911/filetrackermodification/app/internal/config"
	"github.com/thespider911/filetrackermodification/app/internal/service/filetrack"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendToAPI(t *testing.T) {
	// create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// check if the request method is POST
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// check if the content type if is application/json
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// decode the request body
		var receivedInfo filetrack.FileInfo
		err := json.NewDecoder(r.Body).Decode(&receivedInfo)
		if err != nil {
			t.Errorf("Error decoding request body: %v", err)
		}

		// check if the received data matches what we expect
		if receivedInfo.Filename != "testfile.txt" {
			t.Errorf("Expected FileName 'testfile.txt', got '%s'", receivedInfo.Filename)
		}

		// send a 200 OK response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// create a mock application
	app := &application{
		httpClient: http.DefaultClient,
		config: config.Config{
			APIEndpoint: server.URL,
		},
	}

	// create a mock FileInfo
	mockInfo := filetrack.FileInfo{
		Filename:     "testfile.txt",
		Path:         "/path/to/testfile.txt",
		FileSize:     "1024",
		ModifiedTime: time.Now().String(),
		AccessedTime: time.Now().String(),
		ChangedTime:  time.Now().String(),
		Permission:   "rw-r--r--",
	}

	// call the sendToAPI function
	err := app.sendToAPI(mockInfo)
	if err != nil {
		t.Errorf("sendToAPI returned an error: %v", err)
	}

	// test with server returning non-200 status
	server.Close()
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	app.config.APIEndpoint = server.URL
	err = app.sendToAPI(mockInfo)
	if err == nil {
		t.Error("Expected an error when server returns non-200 status, but got nil")
	} else if err.Error() != "API returned non-200 status code: 500" {
		t.Errorf("Expected error 'API returned non-200 status code: 500', got '%v'", err)
	}
}
