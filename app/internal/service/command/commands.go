package command

//
//// FileInfo - file info struct
//type FileInfo struct {
//	Uid       string `json:"uid"`
//	Path      string `json:"path"`
//	Directory string `json:"directory"`
//	Filename  string `json:"filename"`
//	Mtime     string `json:"mtime"`
//	ATime     string `json:"atime"`
//	CTime     string `json:"ctime"`
//	Size      string `json:"size"`
//	Type      string `json:"type"`
//	Mode      string `json:"mode"`
//}
//
//type PermissionModeInfos struct {
//	Type        string `json:"type"`
//	Path        string `json:"path"`
//	Permissions string `json:"mode"`
//}
//
//type FileTypeInfos struct {
//	Path     string `json:"path"`
//	Filename string `json:"filename"`
//	Type     string `json:"type"`
//}
//
//type FileDates struct {
//	Path      string `json:"path"`
//	Filename  string `json:"filename"`
//	Mtime     string `json:"modified_time"`
//	MTimeDiff string `json:"modified_time_diff"`
//	ATime     string `json:"accessed_time"`
//	ATimeDiff string `json:"accessed_time_diff"`
//	CTime     string `json:"changed_time"`
//	CTimeDiff string `json:"changed_time_diff"`
//}
//
//type FileModified struct {
//	Path      string `json:"path"`
//	Filename  string `json:"filename"`
//	CTime     string `json:"changed_time"`
//	CTimeDiff string `json:"changed_time_diff"`
//}

type Command struct {
	Type string
	Data interface{}
}

//// CommandInfo - to stores information about each command
//type CommandInfo struct {
//	Name        string `json:"name"`
//	Description string `json:"description"`
//	Usage       string `json:"usage"`
//}
//
//var (
//	commandQueue   chan Command
//	logFile        *os.File
//	logBuffer      = make([]FileInfo, 0, 1000) // buffer for last 1000 entries
//	logBufferMu    sync.RWMutex
//	httpClient     = &http.Client{Timeout: 10 * time.Second}
//	notAFile       = errors.New("invalid file path, must be a file")
//	isRunning      bool
//	serviceWg      sync.WaitGroup
//	serviceStopper chan struct{}
//)
//
//// --------------- COMMANDS --------------- //
//var commandInfoMap = map[string]CommandInfo{
//	"CHECK_DIRECTORY_FILE": {
//		Name:        "CHECK_DIRECTORY_FILE",
//		Description: "Checks file information for a given file path",
//		Usage:       "/query?command=CHECK_DIRECTORY_FILE&path=/path/to/file",
//	},
//	"CHECK_FILE_PERMISSION": {
//		Name:        "CHECK_FILE_PERMISSION",
//		Description: "Checks file permission of a given file",
//		Usage:       "/query?command=CHECK_FILE_PERMISSION&path=/path/to/file",
//	},
//	"CHECK_FILE_TYPE": {
//		Name:        "CHECK_FILE_TYPE",
//		Description: "Checks file type for a given path",
//		Usage:       "/query?command=CHECK_FILE_TYPE&path=/path/to/file",
//	},
//	"CHECK_IS_FILE_TYPE": {
//		Name:        "CHECK_IS_FILE_TYPE",
//		Description: "Checks documents for a given path",
//		Usage:       "/query?command=CHECK_IS_FILE_TYPE&path=/path/to/file",
//	},
//	"CHECK_FILE_DATES": {
//		Name:        "CHECK_FILE_DATES",
//		Description: "Checks file times",
//		Usage:       "/query?command=CHECK_FILE_DATES&path=/path/to/file",
//	},
//	"CHECK_IF_MODIFIED_FILE": {
//		Name:        "CHECK_IF_MODIFIED_FILE",
//		Description: "Checks file modified for a given path",
//		Usage:       "/query?command=CHECK_IF_MODIFIED_FILE&path=/path/to/file",
//	},
//}
//
////--------------- HANDLERS --------------- //

//// commandQueryHandler handles queries about commands
//func commandQueryHandler(w http.ResponseWriter, r *http.Request) {
//	query := r.URL.Query()
//	command := query.Get("command")
//
//	w.Header().Set("Content-Type", "application/json")
//
//	if command == "" {
//		// if no specific command is queried, return all commands
//		json.NewEncoder(w).Encode(commandInfoMap)
//		return
//	}
//
//	// convert command to uppercase for case-insensitive matching
//	command = strings.ToUpper(command)
//
//	if info, ok := commandInfoMap[command]; ok {
//		json.NewEncoder(w).Encode(info)
//	} else {
//		http.Error(w, "Command not found", http.StatusNotFound)
//	}
//}
//
//// commandExecuteHandler handles execution of commands
//func commandExecuteHandler(w http.ResponseWriter, r *http.Request) {
//	query := r.URL.Query()
//	command := query.Get("command")
//
//	if command == "" {
//		http.Error(w, "Command parameter is required", http.StatusBadRequest)
//		return
//	}
//
//	// convert command to uppercase for case-insensitive matching
//	command = strings.ToUpper(command)
//
//	params := make(map[string]string)
//	for key, values := range query {
//		if key != "command" && len(values) > 0 {
//			params[key] = values[0]
//		}
//	}
//
//	result, err := executeCommand(command, params)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	json.NewEncoder(w).Encode(result)
//}

// --------------- MAIN --------------- //

//
//// --------------- COMMANDS --------------- //
//// fetchFileInfo - get file info from querying the path returning fileInfo
//func fetchFileInfo(filePath string) (*FileInfo, error) {
//	// osquery query and command run
//	query := fmt.Sprintf("SELECT uid, path, directory, filename, mtime, atime, ctime, size, type, mode FROM file WHERE path = '%s';", filePath)
//	cmd := exec.Command("osqueryi", "--json", query)
//
//	output, err := cmd.Output()
//	if err != nil {
//		return nil, err
//	}
//
//	// decode the output
//	var fileInfos []FileInfo
//	err = json.Unmarshal(output, &fileInfos)
//	if err != nil {
//		return nil, err
//	}
//
//	// return all file infos
//	if len(fileInfos) > 1 {
//		return nil, notAFile
//	} else if len(fileInfos) == 1 {
//		newFileInfos := fileInfos[0]
//
//		//only accept files
//		if newFileInfos.Type != "regular" {
//			return nil, notAFile
//		}
//
//		//format time and size values
//		newFileInfos.Mtime = toHumanReadableTime(newFileInfos.Mtime)
//		newFileInfos.ATime = toHumanReadableTime(newFileInfos.ATime)
//		newFileInfos.CTime = toHumanReadableTimeDiff(newFileInfos.CTime)
//		newFileInfos.Size = toHumanReadableFileSize(newFileInfos.Size)
//		return &newFileInfos, nil
//	}
//
//	return nil, errors.New("no result found")
//}
//
//// fetchFilePermissions - get file permissions
//func fetchFilePermissions(filePath string) (*PermissionModeInfos, error) {
//	query := fmt.Sprintf("SELECT path, mode, type FROM file WHERE path = '%s';", filePath)
//	cmd := exec.Command("osqueryi", "--json", query)
//
//	output, err := cmd.Output()
//	if err != nil {
//		return nil, err
//	}
//
//	// decode the output
//	var modePermInfos []PermissionModeInfos
//	err = json.Unmarshal(output, &modePermInfos)
//	if err != nil {
//		return nil, err
//	}
//
//	// return all file infos
//	if len(modePermInfos) > 1 {
//		return nil, notAFile
//	} else if len(modePermInfos) == 1 {
//		newModPermInfo := modePermInfos[0]
//
//		//only accept files
//		if newModPermInfo.Type != "regular" {
//			return nil, notAFile
//		}
//
//		return &newModPermInfo, nil
//	}
//
//	return nil, errors.New("no result found")
//}
//
//// fetchFileType - get file types
//func fetchFileType(filePath string) (*FileTypeInfos, error) {
//	query := fmt.Sprintf("SELECT path, filename, type FROM file WHERE path = '%s';", filePath)
//	cmd := exec.Command("osqueryi", "--json", query)
//
//	output, err := cmd.Output()
//	if err != nil {
//		return nil, err
//	}
//
//	// decode the output
//	var fileInfos []FileTypeInfos
//	err = json.Unmarshal(output, &fileInfos)
//	if err != nil {
//		return nil, err
//	}
//
//	// return all file infos
//	if len(fileInfos) > 0 {
//		return &fileInfos[0], nil
//	}
//
//	return nil, errors.New("no result found")
//}
//
//// fetchFileIsDirectory - check file directory
//func fetchIsFile(filePath string) (bool, error) {
//	query := fmt.Sprintf("SELECT path, filename, type FROM file WHERE path = '%s';", filePath)
//	cmd := exec.Command("osqueryi", "--json", query)
//
//	output, err := cmd.Output()
//	if err != nil {
//		return false, err
//	}
//
//	// decode the output
//	var fileInfos []FileTypeInfos
//	err = json.Unmarshal(output, &fileInfos)
//	if err != nil {
//		return false, err
//	}
//
//	// return all file infos
//	if len(fileInfos) > 1 {
//		return false, notAFile
//	} else if len(fileInfos) == 1 {
//		newInfo := fileInfos[0]
//
//		//only accept files
//		if newInfo.Type != "regular" {
//			return false, notAFile
//		}
//
//		return true, nil
//	}
//
//	return false, errors.New("no result found")
//}
//
//// fetchFileDate - get file date
//func fetchFileDate(filePath string) (*FileDates, error) {
//	query := fmt.Sprintf("SELECT path, filename, mtime, atime, ctime, type  FROM file WHERE path = '%s';", filePath)
//	cmd := exec.Command("osqueryi", "--json", query)
//
//	output, err := cmd.Output()
//	if err != nil {
//		return nil, err
//	}
//
//	// decode the output
//	var fileInfos []struct {
//		Path     string `json:"path"`
//		Filename string `json:"filename"`
//		Mtime    string `json:"mtime"`
//		ATime    string `json:"atime"`
//		CTime    string `json:"ctime"`
//		Type     string `json:"type"`
//	}
//	err = json.Unmarshal(output, &fileInfos)
//	if err != nil {
//		return nil, err
//	}
//
//	// return all file infos
//	if len(fileInfos) > 1 {
//		return nil, notAFile
//	} else if len(fileInfos) == 1 {
//
//		fileInfo := fileInfos[0]
//
//		//only accept files
//		if fileInfo.Type != "regular" {
//			return nil, notAFile
//		}
//
//		return &FileDates{
//			Path:      fileInfo.Path,
//			Filename:  fileInfo.Filename,
//			Mtime:     toHumanReadableTime(fileInfo.Mtime),
//			MTimeDiff: toHumanReadableTimeDiff(fileInfo.Mtime),
//			ATime:     toHumanReadableTime(fileInfo.ATime),
//			ATimeDiff: toHumanReadableTimeDiff(fileInfo.ATime),
//			CTime:     toHumanReadableTime(fileInfo.CTime),
//			CTimeDiff: toHumanReadableTimeDiff(fileInfo.CTime),
//		}, nil
//	}
//
//	return nil, errors.New("no result found")
//}
//
//// fetchFileIsModified - get file date
//func fetchFileIsModified(filePath string) (*FileModified, error) {
//	query := fmt.Sprintf("SELECT path, filename, ctime, type  FROM file WHERE path = '%s';", filePath)
//	cmd := exec.Command("osqueryi", "--json", query)
//
//	output, err := cmd.Output()
//	if err != nil {
//		return nil, err
//	}
//
//	// decode the output
//	var fileInfos []struct {
//		Path     string `json:"path"`
//		Filename string `json:"filename"`
//		CTime    string `json:"ctime"`
//		Type     string `json:"type"`
//	}
//	err = json.Unmarshal(output, &fileInfos)
//	if err != nil {
//		return nil, err
//	}
//
//	// return all file infos
//	if len(fileInfos) > 1 {
//		return nil, notAFile
//	} else if len(fileInfos) == 1 {
//
//		fileInfo := fileInfos[0]
//
//		//only accept files
//		if fileInfo.Type != "regular" {
//			return nil, notAFile
//		}
//
//		return &FileModified{
//			Path:      fileInfo.Path,
//			Filename:  fileInfo.Filename,
//			CTime:     toHumanReadableTime(fileInfo.CTime),
//			CTimeDiff: toHumanReadableTimeDiff(fileInfo.CTime),
//		}, nil
//	}
//
//	return nil, errors.New("no result found")
//}

// --------------- COMMANDS --------------- //

//// executeCommand executes a given command
//func executeCommand(command string, params map[string]string) (interface{}, error) {
//	switch command {
//	case "CHECK_DIRECTORY_FILE":
//		if path, ok := params["path"]; ok {
//			result, err := fetchFileInfo(path)
//			if err != nil {
//				return nil, err
//			}
//			return result, nil
//		}
//		return nil, fmt.Errorf("path parameter is required for CHECK_DIRECTORY_FILE")
//
//	case "CHECK_FILE_PERMISSION":
//		if path, ok := params["path"]; ok {
//			result, err := fetchFilePermissions(path)
//			if err != nil {
//				return nil, err
//			}
//			return result, nil
//		}
//		return nil, fmt.Errorf("path parameter is required for CHECK_FILE_PERMISSION")
//
//	case "CHECK_FILE_TYPE":
//		if path, ok := params["path"]; ok {
//			result, err := fetchFileType(path)
//			if err != nil {
//				return nil, err
//			}
//			return result, nil
//		}
//		return nil, fmt.Errorf("path parameter is required for CHECK_FILE_TYPE")
//
//	case "CHECK_IS_FILE_TYPE":
//		if path, ok := params["path"]; ok {
//			result, err := fetchIsFile(path)
//			if err != nil {
//				return nil, err
//			}
//			return result, nil
//		}
//		return nil, fmt.Errorf("path parameter is required for CHECK_IS_FILE_TYPE")
//
//	case "CHECK_FILE_DATES":
//		if path, ok := params["path"]; ok {
//			result, err := fetchFileDate(path)
//			if err != nil {
//				return nil, err
//			}
//			return result, nil
//		}
//		return nil, fmt.Errorf("path parameter is required for CHECK_FILE_DATES")
//
//	case "CHECK_IF_MODIFIED_FILE":
//		if path, ok := params["path"]; ok {
//			result, err := fetchFileIsModified(path)
//			if err != nil {
//				return nil, err
//			}
//			return result, nil
//		}
//		return nil, fmt.Errorf("path parameter is required for CHECK_IF_MODIFIED_FILE")
//
//	default:
//		return nil, fmt.Errorf("unknown command: %s", command)
//	}
//}
//
