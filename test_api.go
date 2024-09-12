package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"sync"
)

type FileInfo1 struct {
	Uuid      string `json:"uid"`
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

type Config1 struct {
	ApiPort int `mapstructure:"api_port" validate:"required,min=4041,max=4045"`
}

var (
	receivedData []FileInfo1
	dataMutex    sync.RWMutex
	config1      Config1
)

func main() {
	// load config
	if err := loadConfig1(); err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	http.HandleFunc("/file-updates", handleFileUpdates)
	http.HandleFunc("/view-data", viewData)

	//starting microservice server
	fmt.Printf("Starting test API server on port %d\n", config1.ApiPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config1.ApiPort), nil))
}

// loadConfig1 - yaml file with viper and validate
func loadConfig1() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file - %w", err)
	}

	if err := viper.Unmarshal(&config1); err != nil {
		return fmt.Errorf("error unmarshalling config - %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(config1); err != nil {
		return err
	}

	return nil
}

// handleFileUpdates - receive file updates
func handleFileUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var fileInfo FileInfo1
	err := json.NewDecoder(r.Body).Decode(&fileInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dataMutex.Lock()
	receivedData = append(receivedData, fileInfo)
	dataMutex.Unlock()

	//fmt.Printf("Received data: %+v\n", fileInfo)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data received successfully"))
}

// viewData - view all received data from microservice
func viewData(w http.ResponseWriter, r *http.Request) {
	dataMutex.RLock()
	defer dataMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(receivedData)
}
