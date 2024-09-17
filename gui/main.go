package main

//
//import (
//	"bytes"
//	"github.com/thespider911/filetrackermodification/app/internal/config"
//	"github.com/thespider911/filetrackermodification/app/internal/service"
//	"github.com/thespider911/filetrackermodification/app/internal/service/filetrack"
//	"log"
//	"net/http"
//	"os"
//	"sync"
//	"time"
//)
//
//type Command struct {
//	Type string
//	Data interface{}
//}
//
//type application struct {
//	infoLog        *log.Logger
//	errorLog       *log.Logger
//	config         config.Config
//	wg             sync.WaitGroup
//	wgCount        int32
//	logFile        *os.File
//	service        service.Service
//	commandQueue   chan Command
//	logBufferMu    sync.RWMutex
//	logBuffer      []filetrack.FileInfo
//	httpClient     *http.Client
//	isRunning      bool
//	serviceStopper chan struct{}
//	logChan        chan string
//}
//
//func main() {
//	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
//	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime)
//
//	// config instance
//	cfg, err := config.GetConfig()
//	if err != nil {
//		errorLog.Fatal(err)
//	}
//
//	app := &application{
//		infoLog:        infoLog,
//		errorLog:       errorLog,
//		config:         *cfg,
//		service:        service.NewService(),
//		commandQueue:   make(chan Command, cfg.QueueSize),
//		logBuffer:      make([]filetrack.FileInfo, 0, 1000),
//		httpClient:     &http.Client{Timeout: 10 * time.Second},
//		isRunning:      true,
//		serviceStopper: make(chan struct{}),
//		logChan:        make(chan string, 100),
//	}
//
//	//set up logging
//	app.logging()
//	defer app.logFile.Close()
//
//	// waitGroup to run services on diff threads
//	var wg sync.WaitGroup
//	wg.Add(3) //
//
//	// start worker thread (running continuous executing available commands)
//	go func() {
//		defer wg.Done()
//		if err := app.workerThread(); err != nil {
//			app.errorLog.Println(err)
//		}
//	}()
//
//	// start timer thread (which is checking files modified in one minute lapse sending commands)
//	go func() {
//		defer wg.Done()
//		if err := app.timerThread(); err != nil {
//			app.errorLog.Println(err)
//		}
//	}()
//
//	// start HTTP server
//	go func() {
//		defer wg.Done()
//		if err := app.serveHttp(); err != nil {
//			app.errorLog.Println(err)
//			os.Exit(1)
//		}
//	}()
//
//	wg.Wait()
//}
//
//// --------------- API --------------- //
//
//// sentToApi - convert file into to json then send as response to api endpoint that it has access
//func (app *application) sendToAPI(info filetrack.FileInfo) error {
//	//file info to json
//	jsonData, err := app.JSON(info)
//	if err != nil {
//		return err
//	}
//
//	// use httpClient to send a post response to api endpoint
//	resp, err := app.httpClient.Post(app.config.APIEndpoint, "application/json", bytes.NewBuffer(jsonData))
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return err
//	}
//
//	return nil
//}
//
//func (app *application) appendLog(text string) {
//	select {
//	case app.logChan <- text:
//		// Message sent successfully
//	default:
//		// Channel is full, log to error log
//		app.errorLog.Printf("Log channel full, couldn't log: %s", text)
//	}
//}
