package main

import (
	"bytes"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/thespider911/filetrackermodification/app/internal/config"
	"github.com/thespider911/filetrackermodification/app/internal/service"
	"github.com/thespider911/filetrackermodification/app/internal/service/filetrack"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type Command struct {
	Type string
	Data interface{}
}

type application struct {
	infoLog        *log.Logger
	errorLog       *log.Logger
	config         config.Config
	wg             sync.WaitGroup
	wgCount        int32
	logFile        *os.File
	service        service.Service
	commandQueue   chan Command
	logBufferMu    sync.RWMutex
	logBuffer      []filetrack.FileInfo
	httpClient     *http.Client
	isRunning      bool
	serviceStopper chan struct{}
	uiLogs         *widget.Entry
	logChan        chan string
}

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime)

	// config instance
	cfg, err := config.GetConfig()
	if err != nil {
		errorLog.Fatal(err)
	}

	application := &application{
		infoLog:        infoLog,
		errorLog:       errorLog,
		config:         *cfg,
		service:        service.NewService(),
		commandQueue:   make(chan Command, cfg.QueueSize),
		logBuffer:      make([]filetrack.FileInfo, 0, 1000),
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		isRunning:      false,
		serviceStopper: make(chan struct{}),
		logChan:        make(chan string, 100),
	}

	//set up logging
	application.logging()
	defer application.logFile.Close()

	err = application.checkDirectory()
	if err != nil {
		errorLog.Printf("Directory check failed: %v\n", err)
		return
	}

	// Create Fyne application
	myApp := app.New()
	myWindow := myApp.NewWindow("File Tracker Service")

	// Create UI elements
	application.uiLogs = widget.NewMultiLineEntry()
	application.uiLogs.Disable()

	startButton := widget.NewButton("Start Service", func() {
		if !application.isRunning {
			application.startService()
		}
	})

	stopButton := widget.NewButton("Stop Service", func() {
		if application.isRunning {
			application.stopService()
		}
	})

	buttons := container.NewHBox(startButton, stopButton)
	content := container.NewBorder(buttons, nil, nil, nil, application.uiLogs)

	// Set up window
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(600, 400))

	// Start log update goroutine
	go application.updateLogs()

	// start HTTP server
	go func() {
		if err := application.serveHttp(); err != nil {
			application.errorLog.Println(err)
			os.Exit(1)
		}
	}()

	// start the service
	application.startService()

	// Run the UI
	myWindow.ShowAndRun()
}

//START AND STOP SERVICE

func (app *application) startService() {
	app.logBufferMu.Lock()
	defer app.logBufferMu.Unlock()

	if app.isRunning {
		app.appendLog("Service is already running.\n")
		return
	}

	app.isRunning = true
	app.serviceStopper = make(chan struct{})

	// Reset WaitGroup count
	atomic.StoreInt32(&app.wgCount, 0)

	// Start workerThread
	atomic.AddInt32(&app.wgCount, 1)
	app.wg.Add(1)
	go func() {
		app.appendLog("Worker thread starting...\n")
		if err := app.workerThread(); err != nil {
			app.errorLog.Printf("Worker thread error: %v\n", err)
			app.appendLog(fmt.Sprintf("Worker thread error: %v\n", err))
		}
	}()

	// Start timerThread
	atomic.AddInt32(&app.wgCount, 1)
	app.wg.Add(1)
	go func() {
		app.appendLog("Timer thread starting...\n")
		if err := app.timerThread(); err != nil {
			app.errorLog.Printf("Timer thread error: %v\n", err)
			app.appendLog(fmt.Sprintf("Timer thread error: %v\n", err))
		}
	}()

	app.appendLog("Service started.\n")

	// Trigger an immediate check of the directory
	go app.initialDirectoryCheck()
}

func (app *application) stopService() {
	app.logBufferMu.Lock()
	defer app.logBufferMu.Unlock()

	if !app.isRunning {
		app.appendLog("Service is already stopped.\n")
		return
	}
	app.isRunning = false

	if app.serviceStopper != nil {
		close(app.serviceStopper)
	}

	// Log the current WaitGroup count
	app.appendLog(fmt.Sprintf("Current WaitGroup count before waiting: %d\n", atomic.LoadInt32(&app.wgCount)))

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		app.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		app.appendLog("All threads stopped successfully.\n")
	case <-time.After(5 * time.Second):
		app.appendLog("Warning: Service stop timed out. Some operations may still be running.\n")
		// Log the current WaitGroup count after timeout
		app.appendLog(fmt.Sprintf("Current WaitGroup count after timeout: %d\n", atomic.LoadInt32(&app.wgCount)))
	}

	app.serviceStopper = nil

	app.appendLog("Service stopped.\n")
}

func (app *application) initialDirectoryCheck() {
	app.appendLog("Starting initial directory check...\n")
	err := filepath.Walk(app.config.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			app.appendLog(fmt.Sprintf("Error accessing path %s: %v\n", path, err))
			return err
		}

		if !info.IsDir() {
			app.appendLog(fmt.Sprintf("Queueing file for check: %s\n", path))
			app.commandQueue <- Command{
				Type: "CHECK_DIRECTORY_FILES",
				Data: path,
			}
		}
		return nil
	})
	if err != nil {
		app.errorLog.Printf("Error in initial directory walk: %v\n", err)
		app.appendLog(fmt.Sprintf("Error in initial directory walk: %v\n", err))
	}
	app.appendLog("Initial directory check completed.\n")
}

func (app *application) appendLog(text string) {
	select {
	case app.logChan <- text:
		// Message sent successfully
	default:
		// Channel is full, log to error log
		app.errorLog.Printf("Log channel full, couldn't log: %s", text)
	}
}

func (app *application) updateLogs() {
	for text := range app.logChan {
		app.uiLogs.SetText(app.uiLogs.Text + text)
		app.uiLogs.CursorRow = len(app.uiLogs.Text)
		// Also log to the console for debugging
		fmt.Print(text)
	}
}

// Add this method to check if the directory exists and is accessible
func (app *application) checkDirectory() error {
	_, err := os.Stat(app.config.Directory)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", app.config.Directory)
	}
	if err != nil {
		return fmt.Errorf("error accessing directory: %s, error: %v", app.config.Directory, err)
	}
	return nil
}

// --------------- API --------------- //

// sentToApi - convert file into to json then send as response to api endpoint that it has access
func (app *application) sendToAPI(info filetrack.FileInfo) error {
	//file info to json
	jsonData, err := app.JSON(info)
	if err != nil {
		return err
	}

	// use httpClient to send a post response to api endpoint
	resp, err := app.httpClient.Post(app.config.APIEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	return nil
}
