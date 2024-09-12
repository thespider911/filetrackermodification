package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FileInfo - file info struct
type FileInfo struct {
	Uid       string `json:"uid"`
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

type PermissionModeInfos struct {
	Type        string `json:"type"`
	Path        string `json:"path"`
	Permissions string `json:"mode"`
}

type FileTypeInfos struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Type     string `json:"type"`
}

type FileDates struct {
	Path      string `json:"path"`
	Filename  string `json:"filename"`
	Mtime     string `json:"modified_time"`
	MTimeDiff string `json:"modified_time_diff"`
	ATime     string `json:"accessed_time"`
	ATimeDiff string `json:"accessed_time_diff"`
	CTime     string `json:"changed_time"`
	CTimeDiff string `json:"changed_time_diff"`
}

type FileModified struct {
	Path      string `json:"path"`
	Filename  string `json:"filename"`
	CTime     string `json:"changed_time"`
	CTimeDiff string `json:"changed_time_diff"`
}

type Command struct {
	Type string
	Data interface{}
}

// Config -
type Config struct {
	Directory     string `mapstructure:"directory" validate:"required,dir"`
	CheckInterval int    `mapstructure:"check_interval" validate:"required,min=1"`
	QueueSize     int    `mapstructure:"queue_size" validate:"required,min=1"`
	Port          int    `mapstructure:"port" validate:"required,min=4000,max=4040"`
	APIEndpoint   string `mapstructure:"api_endpoint" validate:"required,url"`
}

// CommandInfo - to stores information about each command
type CommandInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
}

var (
	commandQueue chan Command
	config       Config
	logFile      *os.File
	logBuffer    = make([]FileInfo, 0, 1000) // buffer for last 1000 entries
	logBufferMu  sync.RWMutex
	httpClient   = &http.Client{Timeout: 10 * time.Second}
	notAFile     = errors.New("invalid file path, must be a file")
)

// --------------- COMMANDS --------------- //
var commandInfoMap = map[string]CommandInfo{
	"CHECK_DIRECTORY_FILE": {
		Name:        "CHECK_DIRECTORY_FILE",
		Description: "Checks file information for a given file path",
		Usage:       "/query?command=CHECK_DIRECTORY_FILE&path=/path/to/file",
	},
	"CHECK_FILE_PERMISSION": {
		Name:        "CHECK_FILE_PERMISSION",
		Description: "Checks file permission of a given file",
		Usage:       "/query?command=CHECK_FILE_PERMISSION&path=/path/to/file",
	},
	"CHECK_FILE_TYPE": {
		Name:        "CHECK_FILE_TYPE",
		Description: "Checks file type for a given path",
		Usage:       "/query?command=CHECK_FILE_TYPE&path=/path/to/file",
	},
	"CHECK_IS_FILE_TYPE": {
		Name:        "CHECK_IS_FILE_TYPE",
		Description: "Checks documents for a given path",
		Usage:       "/query?command=CHECK_IS_FILE_TYPE&path=/path/to/file",
	},
	"CHECK_FILE_DATES": {
		Name:        "CHECK_FILE_DATES",
		Description: "Checks file times",
		Usage:       "/query?command=CHECK_FILE_DATES&path=/path/to/file",
	},
	"CHECK_IF_MODIFIED_FILE": {
		Name:        "CHECK_IF_MODIFIED_FILE",
		Description: "Checks file modified for a given path",
		Usage:       "/query?command=CHECK_IF_MODIFIED_FILE&path=/path/to/file",
	},
}

//--------------- HANDLERS --------------- //

// healthCheckHandler - check system health
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// logsHandler - handle logs
func logsHandler(w http.ResponseWriter, r *http.Request) {
	logBufferMu.RLock()
	defer logBufferMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logBuffer)
}

// helpCommandsHandler - handle help commands
func helpCommandsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Help commands available"))
}

// commandQueryHandler handles queries about commands
func commandQueryHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	command := query.Get("command")

	w.Header().Set("Content-Type", "application/json")

	if command == "" {
		// if no specific command is queried, return all commands
		json.NewEncoder(w).Encode(commandInfoMap)
		return
	}

	// convert command to uppercase for case-insensitive matching
	command = strings.ToUpper(command)

	if info, ok := commandInfoMap[command]; ok {
		json.NewEncoder(w).Encode(info)
	} else {
		http.Error(w, "Command not found", http.StatusNotFound)
	}
}

// commandExecuteHandler handles execution of commands
func commandExecuteHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	command := query.Get("command")

	if command == "" {
		http.Error(w, "Command parameter is required", http.StatusBadRequest)
		return
	}

	// convert command to uppercase for case-insensitive matching
	command = strings.ToUpper(command)

	params := make(map[string]string)
	for key, values := range query {
		if key != "command" && len(values) > 0 {
			params[key] = values[0]
		}
	}

	result, err := executeCommand(command, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// --------------- MAIN --------------- //

func main() {
	// load config
	if err := loadConfig(); err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	//set up logging
	setupLogging()
	defer logFile.Close()

	commandQueue = make(chan Command, config.QueueSize)

	// waitGroup to run services on diff threads
	var wg sync.WaitGroup
	wg.Add(3) //specify my two threads

	// start worker thread (running continuous executing available commands)
	go func() {
		defer wg.Done()
		workerThread()
	}()

	// start timer thread (which is checking files modified in one minute lapse sending commands)
	go func() {
		defer wg.Done()
		timerThread()
	}()

	// start HTTP server
	go func() {
		defer wg.Done()
		startHTTPServer()
	}()

	wg.Wait()
}

// --------------- SERVER AND HTTP REQUESTS --------------- //

func startHTTPServer() {
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/logs", logsHandler)
	http.HandleFunc("/help", helpCommandsHandler)
	http.HandleFunc("/query", commandQueryHandler)
	http.HandleFunc("/execute", commandExecuteHandler)

	log.Printf("Starting HTTP server on port %d\n", config.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil); err != nil {
		log.Fatal(err)
	}
}

// --------------- CONFIG --------------- //

// LoadConfig - yaml file with viper and validate
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

// fetchFilesInfo - get files info from querying the path returning fileInfo
func fetchFilesInfo(filePath string) *FileInfo {
	// osquery query and command run
	query := fmt.Sprintf("SELECT uid, path, directory, filename, mtime, atime, ctime, size, type, mode FROM file WHERE path = '%s';", filePath)
	cmd := exec.Command("osqueryi", "--json", query)

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running osquery:", err)
		return nil
	}

	// decode the output
	var fileInfos []FileInfo
	err = json.Unmarshal(output, &fileInfos)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}

	// return all file infos
	if len(fileInfos) > 0 {
		return &fileInfos[0]
	}

	return nil
}

// --------------- COMMANDS --------------- //
// fetchFileInfo - get file info from querying the path returning fileInfo
func fetchFileInfo(filePath string) (*FileInfo, error) {
	// osquery query and command run
	query := fmt.Sprintf("SELECT uid, path, directory, filename, mtime, atime, ctime, size, type, mode FROM file WHERE path = '%s';", filePath)
	cmd := exec.Command("osqueryi", "--json", query)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// decode the output
	var fileInfos []FileInfo
	err = json.Unmarshal(output, &fileInfos)
	if err != nil {
		return nil, err
	}

	// return all file infos
	if len(fileInfos) > 1 {
		return nil, notAFile
	} else if len(fileInfos) == 1 {
		newFileInfos := fileInfos[0]

		//only accept files
		if newFileInfos.Type != "regular" {
			return nil, notAFile
		}

		//format time and size values
		newFileInfos.Mtime = toHumanReadableTime(newFileInfos.Mtime)
		newFileInfos.ATime = toHumanReadableTime(newFileInfos.ATime)
		newFileInfos.CTime = toHumanReadableTimeDiff(newFileInfos.CTime)
		newFileInfos.Size = toHumanReadableFileSize(newFileInfos.Size)
		return &newFileInfos, nil
	}

	return nil, errors.New("no result found")
}

// fetchFilePermissions - get file permissions
func fetchFilePermissions(filePath string) (*PermissionModeInfos, error) {
	query := fmt.Sprintf("SELECT path, mode, type FROM file WHERE path = '%s';", filePath)
	cmd := exec.Command("osqueryi", "--json", query)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// decode the output
	var modePermInfos []PermissionModeInfos
	err = json.Unmarshal(output, &modePermInfos)
	if err != nil {
		return nil, err
	}

	// return all file infos
	if len(modePermInfos) > 1 {
		return nil, notAFile
	} else if len(modePermInfos) == 1 {
		newModPermInfo := modePermInfos[0]

		//only accept files
		if newModPermInfo.Type != "regular" {
			return nil, notAFile
		}

		return &newModPermInfo, nil
	}

	return nil, errors.New("no result found")
}

// fetchFileType - get file types
func fetchFileType(filePath string) (*FileTypeInfos, error) {
	query := fmt.Sprintf("SELECT path, filename, type FROM file WHERE path = '%s';", filePath)
	cmd := exec.Command("osqueryi", "--json", query)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// decode the output
	var fileInfos []FileTypeInfos
	err = json.Unmarshal(output, &fileInfos)
	if err != nil {
		return nil, err
	}

	// return all file infos
	if len(fileInfos) > 0 {
		return &fileInfos[0], nil
	}

	return nil, errors.New("no result found")
}

// fetchFileIsDirectory - check file directory
func fetchIsFile(filePath string) (bool, error) {
	query := fmt.Sprintf("SELECT path, filename, type FROM file WHERE path = '%s';", filePath)
	cmd := exec.Command("osqueryi", "--json", query)

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// decode the output
	var fileInfos []FileTypeInfos
	err = json.Unmarshal(output, &fileInfos)
	if err != nil {
		return false, err
	}

	// return all file infos
	if len(fileInfos) > 1 {
		return false, notAFile
	} else if len(fileInfos) == 1 {
		newInfo := fileInfos[0]

		//only accept files
		if newInfo.Type != "regular" {
			return false, notAFile
		}

		return true, nil
	}

	return false, errors.New("no result found")
}

// fetchFileDate - get file date
func fetchFileDate(filePath string) (*FileDates, error) {
	query := fmt.Sprintf("SELECT path, filename, mtime, atime, ctime, type  FROM file WHERE path = '%s';", filePath)
	cmd := exec.Command("osqueryi", "--json", query)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// decode the output
	var fileInfos []struct {
		Path     string `json:"path"`
		Filename string `json:"filename"`
		Mtime    string `json:"mtime"`
		ATime    string `json:"atime"`
		CTime    string `json:"ctime"`
		Type     string `json:"type"`
	}
	err = json.Unmarshal(output, &fileInfos)
	if err != nil {
		return nil, err
	}

	// return all file infos
	if len(fileInfos) > 1 {
		return nil, notAFile
	} else if len(fileInfos) == 1 {

		fileInfo := fileInfos[0]

		//only accept files
		if fileInfo.Type != "regular" {
			return nil, notAFile
		}

		return &FileDates{
			Path:      fileInfo.Path,
			Filename:  fileInfo.Filename,
			Mtime:     toHumanReadableTime(fileInfo.Mtime),
			MTimeDiff: toHumanReadableTimeDiff(fileInfo.Mtime),
			ATime:     toHumanReadableTime(fileInfo.ATime),
			ATimeDiff: toHumanReadableTimeDiff(fileInfo.ATime),
			CTime:     toHumanReadableTime(fileInfo.CTime),
			CTimeDiff: toHumanReadableTimeDiff(fileInfo.CTime),
		}, nil
	}

	return nil, errors.New("no result found")
}

// fetchFileIsModified - get file date
func fetchFileIsModified(filePath string) (*FileModified, error) {
	query := fmt.Sprintf("SELECT path, filename, ctime, type  FROM file WHERE path = '%s';", filePath)
	cmd := exec.Command("osqueryi", "--json", query)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// decode the output
	var fileInfos []struct {
		Path     string `json:"path"`
		Filename string `json:"filename"`
		CTime    string `json:"ctime"`
		Type     string `json:"type"`
	}
	err = json.Unmarshal(output, &fileInfos)
	if err != nil {
		return nil, err
	}

	// return all file infos
	if len(fileInfos) > 1 {
		return nil, notAFile
	} else if len(fileInfos) == 1 {

		fileInfo := fileInfos[0]

		//only accept files
		if fileInfo.Type != "regular" {
			return nil, notAFile
		}

		return &FileModified{
			Path:      fileInfo.Path,
			Filename:  fileInfo.Filename,
			CTime:     toHumanReadableTime(fileInfo.CTime),
			CTimeDiff: toHumanReadableTimeDiff(fileInfo.CTime),
		}, nil
	}

	return nil, errors.New("no result found")
}

// --------------- LOGGING --------------- //

// logging to file
func setupLogging() {
	var err error
	logFile, err = os.OpenFile("file_monitor.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

// log to file and in-memory
func logFileInfo(fileInfo FileInfo) {
	log.Printf("Uid: %s\nPath: %s\nDirectory: %s\nFilename: %s\nLast Modified: %s\nLast Visit: %s\nLast Change: %s\nSize: %s\nType: %s\nMode: %s\n\n",
		fileInfo.Uid, fileInfo.Path, fileInfo.Directory, fileInfo.Filename, toHumanReadableTime(fileInfo.Mtime), toHumanReadableTime(fileInfo.ATime), toHumanReadableTimeDiff(fileInfo.CTime), toHumanReadableFileSize(fileInfo.Size), fileInfo.Type, fileInfo.Mode)

	//mutex to ensure safe thread access
	logBufferMu.Lock()
	defer logBufferMu.Unlock()

	if len(logBuffer) >= 1000 {
		logBuffer = logBuffer[1:]
	}
	logBuffer = append(logBuffer, fileInfo)
}

// executeCommand executes a given command
func executeCommand(command string, params map[string]string) (interface{}, error) {
	switch command {
	case "CHECK_DIRECTORY_FILE":
		if path, ok := params["path"]; ok {
			result, err := fetchFileInfo(path)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		return nil, fmt.Errorf("path parameter is required for CHECK_DIRECTORY_FILE")

	case "CHECK_FILE_PERMISSION":
		if path, ok := params["path"]; ok {
			result, err := fetchFilePermissions(path)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		return nil, fmt.Errorf("path parameter is required for CHECK_FILE_PERMISSION")

	case "CHECK_FILE_TYPE":
		if path, ok := params["path"]; ok {
			result, err := fetchFileType(path)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		return nil, fmt.Errorf("path parameter is required for CHECK_FILE_TYPE")

	case "CHECK_IS_FILE_TYPE":
		if path, ok := params["path"]; ok {
			result, err := fetchIsFile(path)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		return nil, fmt.Errorf("path parameter is required for CHECK_IS_FILE_TYPE")

	case "CHECK_FILE_DATES":
		if path, ok := params["path"]; ok {
			result, err := fetchFileDate(path)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		return nil, fmt.Errorf("path parameter is required for CHECK_FILE_DATES")

	case "CHECK_IF_MODIFIED_FILE":
		if path, ok := params["path"]; ok {
			result, err := fetchFileIsModified(path)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		return nil, fmt.Errorf("path parameter is required for CHECK_IF_MODIFIED_FILE")

	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// --------------- THREADS --------------- //

/*
* workerThread - this loops through the commands in queue and is to run continuous listening to any commandQueue
 */
func workerThread() {
	//range commands executed
	for cmd := range commandQueue {
		switch cmd.Type {
		// check files command and print file info
		case "CHECK_DIRECTORY_FILES":
			if filePath, ok := cmd.Data.(string); ok {
				fileInfo := fetchFilesInfo(filePath)
				if fileInfo != nil {
					// print the result file information
					logFileInfo(*fileInfo)
					//send to api the file info
					sendToAPI(*fileInfo)
				}
			}
		default:
			fmt.Printf("Unknown command type: %s\n", cmd.Type)
		}
	}
}

/*
*
timeThread - this runs every minute checking all files in the specified directory
- It then calls the commandQueue looping through
*/
func timerThread() {
	//check this thread every minute
	ticker := time.NewTicker(time.Duration(config.CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		err := filepath.Walk(config.Directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				commandQueue <- Command{
					Type: "CHECK_DIRECTORY_FILES",
					Data: path,
				}
			}
			return nil
		})
		if err != nil {
			log.Printf("Error walking through directory: %v\n", err)
		}
	}
}

// --------------- API --------------- //

// sentToApi
func sendToAPI(info FileInfo) {
	jsonData, err := json.Marshal(info)
	if err != nil {
		log.Printf("Error marshaling JSON for API: %v\n", err)
		return
	}

	resp, err := httpClient.Post(config.APIEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending data to API: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API returned non-OK status: %v\n", resp.Status)
	}
}

// --------------- HELPERS --------------- //
func toHumanReadableTime(val string) string {
	unix, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Error converting time: %v\n", err)
	}

	t := time.Unix(int64(unix), 0)
	return t.Format("Monday 02 January, 2006 03:04 PM")
}

func toHumanReadableFileSize(val string) string {
	size, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Error converting file size: %v\n", err)
	}

	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	return fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])
}

// toHumanReadableTimeDiff - calculate time difference
func toHumanReadableTimeDiff(val string) string {
	unixTime, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Error converting time: %v\n", err)
	}

	now := time.Now()
	t := time.Unix(int64(unixTime), 0)

	//get diff
	duration := now.Sub(t)
	seconds := int(duration.Seconds())
	minutes := int(duration.Minutes())
	hours := int(duration.Hours())
	days := int(hours / 24)

	switch {
	case seconds < 60:
		return fmt.Sprintf("%d seconds ago", seconds)
	case minutes < 60:
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours < 24:
		return fmt.Sprintf("%d hours %d minutes ago", hours, minutes%60)
	default:
		return fmt.Sprintf("%d days %d hours ago", days, hours%24)
	}
}
