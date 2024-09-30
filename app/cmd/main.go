package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/thespider911/filetrackermodification/app/internal/config"
	"github.com/thespider911/filetrackermodification/app/internal/service"
	"github.com/thespider911/filetrackermodification/app/internal/service/filetrack"
	"github.com/thespider911/filetrackermodification/app/internal/testutil"
	"image/color"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// customTheme - theme with black background and white text
type customTheme struct{}

var _ fyne.Theme = (*customTheme)(nil)

func (m customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameBackground {
		if variant == theme.VariantLight {
			return color.White
		}
		return color.Black
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (m customTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m customTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

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
	//setup data to test
	if err := testutil.SetupTestEnvironment(); err != nil {
		fmt.Printf("Error setting up test environment: %v\n", err)
		os.Exit(1)
	}

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
	myApp.Settings().SetTheme(&customTheme{}) // custom theme
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
	application.wg.Add(1)
	go func() {
		defer application.wg.Done()
		if err := application.serveHttp(); err != nil {
			application.errorLog.Println(err)
			os.Exit(0)
		}
	}()

	// Run the UI
	myWindow.ShowAndRun()

	application.wg.Wait()
}
