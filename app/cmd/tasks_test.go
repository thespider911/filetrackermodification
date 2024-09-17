package main

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/thespider911/filetrackermodification/app/internal/service"
	"github.com/thespider911/filetrackermodification/app/internal/service/filetrack"
)

// MockApplication is a mock implementation of the application struct
type MockApplication struct {
	commandQueue    chan Command
	serviceStopper  chan struct{}
	wg              sync.WaitGroup
	service         service.Service
	config          Config
	logFileInfoFunc func(filetrack.FileInfo) error
	sendToAPIFunc   func(filetrack.FileInfo) error
	errorLog        *MockLogger
}

type MockLogger struct {
	lastLoggedMessage string
}

func (m *MockLogger) Println(v ...interface{}) {
	m.lastLoggedMessage = v[0].(string)
}

// Config - app config
type Config struct {
	CheckInterval int
	Directory     string
	APIEndpoint   string
}

// MockFileTracker - mock implementation of the FileTracker interface
type MockFileTracker struct {
	FetchFilesInfoFunc func(string) (*filetrack.FileInfo, error)
}

func (m MockFileTracker) FetchFilesInfo(path string) (*filetrack.FileInfo, error) {
	return m.FetchFilesInfoFunc(path)
}

// workerThread - mock implementation of the worker thread
func workerThread(app *MockApplication) error {
	select {
	case cmd := <-app.commandQueue:
		if cmd.Type == "CHECK_DIRECTORY_FILES" {
			fileInfo, err := app.service.FileTracker.FetchFilesInfo(cmd.Data.(string))
			if err != nil {
				return err
			}
			if fileInfo != nil {
				err = app.logFileInfoFunc(*fileInfo)
				if err != nil {
					return err
				}
				err = app.sendToAPIFunc(*fileInfo)
				if err != nil {
					if errors.Is(err, syscall.ECONNREFUSED) {
						app.errorLog.Println("api service not running")
					} else {
						return err
					}
				}
			}
		}
	case <-app.serviceStopper:
		return nil
	}
	return nil
}

// timerThread - mock implementation of the timer thread
func timerThread(app *MockApplication) error {
	select {
	case <-time.After(time.Duration(app.config.CheckInterval) * time.Second):
		app.commandQueue <- Command{Type: "CHECK_DIRECTORY_FILES", Data: app.config.Directory}
	case <-app.serviceStopper:
		return nil
	}
	return nil
}

// TestWorkerThread - test the workerThread function
func TestWorkerThread(t *testing.T) {
	mockApp := &MockApplication{
		commandQueue:   make(chan Command),
		serviceStopper: make(chan struct{}),
		errorLog:       &MockLogger{},
		service:        service.NewService(),
	}

	mockFileTracker := MockFileTracker{
		FetchFilesInfoFunc: func(path string) (*filetrack.FileInfo, error) {
			return &filetrack.FileInfo{Filename: "test.txt"}, nil
		},
	}

	mockApp.service.FileTracker = mockFileTracker

	mockApp.logFileInfoFunc = func(info filetrack.FileInfo) error {
		return nil
	}

	mockApp.sendToAPIFunc = func(info filetrack.FileInfo) error {
		return nil
	}

	go func() {
		mockApp.commandQueue <- Command{Type: "CHECK_DIRECTORY_FILES", Data: "test/path"}
		close(mockApp.serviceStopper)
	}()

	err := workerThread(mockApp)
	if err != nil {
		t.Errorf("workerThread returned an error: %v", err)
	}
}

// TestWorkerThreadAPIError - tests the workerThread function when the API is not running
func TestWorkerThreadAPIError(t *testing.T) {
	mockApp := &MockApplication{
		commandQueue:   make(chan Command),
		serviceStopper: make(chan struct{}),
		errorLog:       &MockLogger{},
		service:        service.NewService(),
	}

	mockFileTracker := MockFileTracker{
		FetchFilesInfoFunc: func(path string) (*filetrack.FileInfo, error) {
			return &filetrack.FileInfo{Filename: "test.txt"}, nil
		},
	}

	mockApp.service.FileTracker = mockFileTracker

	mockApp.logFileInfoFunc = func(info filetrack.FileInfo) error {
		return nil
	}

	mockApp.sendToAPIFunc = func(info filetrack.FileInfo) error {
		return syscall.ECONNREFUSED
	}

	go func() {
		mockApp.commandQueue <- Command{Type: "CHECK_DIRECTORY_FILES", Data: "test/path"}
		close(mockApp.serviceStopper)
	}()

	err := workerThread(mockApp)
	if err != nil {
		t.Errorf("workerThread returned an error: %v", err)
	}

	if mockApp.errorLog.lastLoggedMessage != "api service not running" {
		t.Errorf("Expected error log 'api service not running', got '%s'", mockApp.errorLog.lastLoggedMessage)
	}
}

// TestTimerThread tests the timerThread function
func TestTimerThread(t *testing.T) {
	mockApp := &MockApplication{
		commandQueue:   make(chan Command),
		serviceStopper: make(chan struct{}),
		config: Config{
			CheckInterval: 1, // 1 second for faster testing
			Directory:     t.TempDir(),
		},
	}

	// Create a test file in the temporary directory
	testFile := filepath.Join(mockApp.config.Directory, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	go func() {
		time.Sleep(2 * time.Second)
		close(mockApp.serviceStopper)
	}()

	commandReceived := make(chan bool)
	go func() {
		cmd := <-mockApp.commandQueue
		if cmd.Type == "CHECK_DIRECTORY_FILES" && cmd.Data.(string) == mockApp.config.Directory {
			commandReceived <- true
		}
	}()

	err := timerThread(mockApp)
	if err != nil {
		t.Errorf("timerThread returned an error: %v", err)
	}

	select {
	case <-commandReceived:
		// Test passed
	case <-time.After(3 * time.Second):
		t.Error("Timed out waiting for command")
	}
}
