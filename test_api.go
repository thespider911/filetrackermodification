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

type FileInfo struct {
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

type Config struct {
	ApiPort int `mapstructure:"api_port" validate:"required,min=4041,max=4045"`
}

var (
	receivedData []FileInfo
	dataMutex    sync.RWMutex
	config       Config
)

func main() {
	// load config
	if err := loadConfig(); err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	http.HandleFunc("/file-endpoint", handleFileUpdates)
	http.HandleFunc("/view-data", viewData)

	//starting microservice server
	fmt.Printf("Starting test API server on port %d\n", config.ApiPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.ApiPort), nil))
}

// loadConfig - yaml file with viper and validate
func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file - %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("error unmarshalling config - %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return err
	}

	return nil
}

// handleFileUpdates - receive file info updates
func handleFileUpdates(w http.ResponseWriter, r *http.Request) {
	//only allow post method
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	//decode file info data
	var fileInfo FileInfo
	err := json.NewDecoder(r.Body).Decode(&fileInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//lock writing
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
