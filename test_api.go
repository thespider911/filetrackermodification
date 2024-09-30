package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type FileInfo struct {
	UUID      string `json:"uid"`
	Path      string `json:"path"`
	Directory string `json:"directory"`
	Filename  string `json:"filename"`
	Mtime     string `json:"mtime"`
	ATime     string `json:"atime"`
	CTime     string `json:"ctime"`
	Size      string `json:"size"`
	Type      string `json:"type"`
	Mode      string `json:"mode"`
}

type Config struct {
	APIPort int `mapstructure:"api_port" validate:"required,min=4041,max=4045"`
}

type App struct {
	config       Config
	receivedData []FileInfo
	dataMutex    sync.RWMutex
}

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("Error initializing application: %v", err)
	}

	http.HandleFunc("/file-endpoint", app.handleFileUpdates)
	http.HandleFunc("/view-data", app.viewData)

	log.Printf("Starting test API server on port %d\n", app.config.APIPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", app.config.APIPort), nil))
}

func NewApp() (*App, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	return &App{
		config:       config,
		receivedData: []FileInfo{},
	}, nil
}

func loadConfig() (Config, error) {
	var config Config

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return config, fmt.Errorf("error reading config file: %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("error unmarshalling config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return config, fmt.Errorf("config validation error: %w", err)
	}

	return config, nil
}

func (app *App) handleFileUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	//decode file info data
	var fileInfo FileInfo
	if err := json.NewDecoder(r.Body).Decode(&fileInfo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.dataMutex.Lock()
	app.receivedData = append(app.receivedData, fileInfo)
	app.dataMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data received successfully"))
}

func (app *App) viewData(w http.ResponseWriter, r *http.Request) {
	app.dataMutex.RLock()
	defer app.dataMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app.receivedData)
}
