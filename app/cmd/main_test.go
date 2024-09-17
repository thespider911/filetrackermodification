package main

//
//import (
//	"encoding/json"
//	"fyne.io/fyne/v2/widget"
//	"github.com/thespider911/filetrackermodification/app/internal/service"
//	"io/ioutil"
//	"log"
//	"net/http"
//	"net/http/httptest"
//	"os"
//	"path/filepath"
//	"sync/atomic"
//	"testing"
//	"time"
//
//	"github.com/thespider911/filetrackermodification/app/internal/config"
//	"github.com/thespider911/filetrackermodification/app/internal/service/filetrack"
//)
//
//// MockEntry is a simplified mock implementation of widget.Entry
//type MockEntry struct {
//	text string
//}
//
//func NewMockEntry() *widget.Entry {
//	return &widget.Entry{}
//}
//
//// Helper function to create a new application with a mock entry
//func newTestApplication() *application {
//	return &application{
//		commandQueue:   make(chan Command, 100),
//		serviceStopper: make(chan struct{}),
//		isRunning:      false,
//		config: config.Config{
//			Directory:     os.TempDir(),
//			CheckInterval: 60, // 1 min for testing purposes
//		},
//		uiLogs:     NewMockEntry(),
//		logChan:    make(chan string, 100),
//		infoLog:    log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
//		errorLog:   log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
//		service:    service.NewService(),
//		httpClient: &http.Client{},
//	}
//}
//
//// MockHTTPClient - mock for the http.Client
//type MockHTTPClient struct {
//	DoFunc func(req *http.Request) (*http.Response, error)
//}
//
//func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
//	return m.DoFunc(req)
//}
//
//func TestTimerThread(t *testing.T) {
//	app := newTestApplication()
//
//	// Start the timer thread
//	app.wg.Add(1)
//	atomic.AddInt32(&app.wgCount, 1)
//	go app.timerThread()
//
//	// Wait for a short period to allow the thread to start
//	time.Sleep(time.Second * 2)
//
//	// Stop the service
//	close(app.serviceStopper)
//
//	// Wait for the thread to stop
//	app.wg.Wait()
//
//	// Check if any commands were queued
//	select {
//	case cmd := <-app.commandQueue:
//		if cmd.Type != "CHECK_DIRECTORY_FILES" {
//			t.Errorf("Expected command type CHECK_DIRECTORY_FILES, got %s", cmd.Type)
//		}
//	default:
//		// It's possible no commands were queued if the directory was empty
//		t.Log("No commands were queued during the test period")
//	}
//}
//
//func TestSendToAPI(t *testing.T) {
//	// Create a test server
//	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		// Read the request body
//		body, err := ioutil.ReadAll(r.Bo//package maindy)
//		if err != nil {
//			t.Errorf("Error reading request body: %v", err)
//			return
//		}
//		defer r.Body.Close()
//
//		// Unmarshal the JSON to verify the data
//		var receivedInfo filetrack.FileInfo
//		err = json.Unmarshal(body, &receivedInfo)
//		if err != nil {
//			t.Errorf("Error unmarshaling request body: %v", err)
//			return
//		}
//
//		// Verify the received data
//		if receivedInfo.Filename != "test.txt" || receivedInfo.FileSize != "1024" {
//			t.Errorf("Received unexpected data: %+v", receivedInfo)
//		}
//
//		w.WriteHeader(http.StatusOK)
//		w.Write([]byte(`OK`))
//	}))
//	defer server.Close()
//
//	app := &application{
//		config: config.Config{
//			APIEndpoint: server.URL, // Use the test server URL
//		},
//		httpClient: &http.Client{},
//	}
//
//	fileInfo := filetrack.FileInfo{
//		Filename:     "test.txt",
//		FileSize:     "1024",
//		ModifiedTime: time.Now().Format(time.RFC3339),
//	}
//
//	err := app.sendToAPI(fileInfo)
//	if err != nil {
//		t.Errorf("sendToAPI() returned an unexpected error: %v", err)
//	}
//}
//
//func TestStartStopService(t *testing.T) {
//	app := newTestApplication()
//
//	// Test starting the service
//	app.startService()
//	if !app.isRunning {
//		t.Error("Service did not start")
//	}
//
//	// Give some time for goroutines to start
//	time.Sleep(100 * time.Millisecond)
//
//	// Test stopping the service
//	app.stopService()
//	if app.isRunning {
//		t.Error("Service did not stop")
//	}
//}
//
//func TestUpdateLogs(t *testing.T) {
//	app := newTestApplication()
//
//	go app.updateLogs()
//
//	testLogs := []string{"Log 1", "Log 2", "Log 3"}
//	for _, log := range testLogs {
//		app.logChan <- log
//	}
//
//	// Give some time for logs to be processed
//	time.Sleep(100 * time.Millisecond)
//
//	expectedLog := "Log 1Log 2Log 3"
//	if app.uiLogs.Text != expectedLog {
//		t.Errorf("Expected logs '%s', but got '%s'", expectedLog, app.uiLogs.Text)
//	}
//}
//
//func TestCheckDirectory(t *testing.T) {
//	// Create a temporary directory for testing
//	tempDir, err := ioutil.TempDir("", "testdir")
//	if err != nil {
//		t.Fatalf("Failed to create temp directory: %v", err)
//	}
//	defer os.RemoveAll(tempDir)
//
//	app := &application{
//		config: config.Config{
//			Directory: tempDir,
//		},
//	}
//
//	// Test with existing directory
//	err = app.checkDirectory()
//	if err != nil {
//		t.Errorf("checkDirectory() returned an error for existing directory: %v", err)
//	}
//
//	// Test with non-existing directory
//	app.config.Directory = "/non/existent/directory"
//	err = app.checkDirectory()
//	if err == nil {
//		t.Error("checkDirectory() did not return an error for non-existent directory")
//	}
//}
//
//func TestInitialDirectoryCheck(t *testing.T) {
//	// Create a temporary directory with some files
//	tempDir := t.TempDir()
//	testFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
//	for _, file := range testFiles {
//		err := ioutil.WriteFile(filepath.Join(tempDir, file), []byte("test content"), 0644)
//		if err != nil {
//			t.Fatalf("Failed to create test file: %v", err)
//		}
//	}
//
//	app := newTestApplication()
//	app.config.Directory = tempDir
//
//	// Run the initial directory check
//	app.initialDirectoryCheck()
//
//	// Check if all files were queued
//	queuedFiles := 0
//	for i := 0; i < len(testFiles); i++ {
//		select {
//		case cmd := <-app.commandQueue:
//			if cmd.Type == "CHECK_DIRECTORY_FILES" {
//				queuedFiles++
//			}
//		case <-time.After(100 * time.Millisecond):
//			// Timeout if not all files were queued
//			break
//		}
//	}
//
//	if queuedFiles != len(testFiles) {
//		t.Errorf("Expected %d files to be queued, but got %d", len(testFiles), queuedFiles)
//	}
//}
//
//func TestAppendLog(t *testing.T) {
//	app := &application{
//		logChan: make(chan string, 10),
//	}
//
//	testLog := "Test log message"
//	app.appendLog(testLog)	tempDir := t.TempDir()
//	testFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
//	for _, file := range testFiles {
//		err := ioutil.WriteFile(filepath.Join(tempDir, file), []byte("test content"), 0644)
//		if err != nil {
//			t.Fatalf("Failed to create test file: %v", err)
//		}
//	}
//
//	app := newTestApplication()
//	app.config.Directory = tempDir
//
//	// Run the initial directory check
//	app.initialDirectoryCheck()
//
//	// Check if all files were queued
//	queuedFiles := 0
//	for i := 0; i < len(testFiles); i++ {
//		select {
//		case cmd := <-app.commandQueue:
//			if cmd.Type == "CHECK_DIRECTORY_FILES" {
//				queuedFiles++
//			}
//		case <-time.After(100 * time.Millisecond):
//			// Timeout if not all files were queued
//			break
//		}
//	}
//
//	if queuedFiles != len(testFiles) {
//		t.Errorf("Expected %d files to be queued, but got %d", len(testFiles), queuedFiles)
//	}
//
//
//	select {
//	case logMsg := <-app.logChan:
//		if logMsg != testLog {
//			t.Errorf("Expected log message '%s', but got '%s'", testLog, logMsg)
//		}
//	case <-time.After(100 * time.Millisecond):
//		t.Error("Timed out waiting for log message")
//	}
//}
