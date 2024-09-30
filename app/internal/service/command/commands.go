package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thespider911/filetrackermodification/app/internal/helpers"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrNoFile = errors.New("models: no such existing file record found")

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

// CommandRunFile -
type CommandRunFile interface {
	ExecuteCommand(string, map[string]string) (interface{}, error)
	FetchFileInfo(string) (*FileInfo, error)
	FetchFilePermissions(string) (*PermissionModeInfos, error)
	FetchFileType(string) (*FileTypeInfos, error)
	FetchIsFile(string) (bool, error)
	FetchFileDate(string) (*FileDates, error)
	FetchFileIsModified(string) (*FileModified, error)
}

type CommandFileInfo struct{}

// NewCommandFileInfo - new instance of CommandFileInfo
func NewCommandFileInfo() CommandRunFile {
	return &CommandFileInfo{}
}

// validatePath - check if the given path is valid, specific, exists, and is within the Desktop directory
func (cf *CommandFileInfo) validatePath(path string) error {
	// absolute path check
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute")
	}

	// check if the path contains any wildcards or patterns
	if strings.ContainsAny(path, "*?[]") {
		return fmt.Errorf("path must not contain wildcards or patterns")
	}

	// check if the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist")
	}

	// check if the path is within the Desktop directory
	desktopPath, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to determine user's home directory: %v", err)
	}
	desktopPath = filepath.Join(desktopPath, "Desktop")

	if !strings.HasPrefix(path, desktopPath) {
		return fmt.Errorf("path must be within the Desktop directory")
	}

	return nil
}

// --------------- COMMANDS --------------- //

// FetchFileInfo - get file info from querying the path returning fileInfo
func (cf *CommandFileInfo) FetchFileInfo(filePath string) (*FileInfo, error) {
	// osquery query and command run
	query := fmt.Sprintf("SELECT uid, path, directory, filename, mtime, atime, ctime, size, type, mode FROM file WHERE path = '%s';", filepath.Clean(filePath))
	cmd := exec.Command("osqueryi", "--json", query)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// decode the output
	var fileInfos []FileInfo
	if err := json.Unmarshal(output, &fileInfos); err != nil {
		return nil, err
	}

	if len(fileInfos) == 0 {
		return nil, ErrNoFile
	}

	if len(fileInfos) > 1 {
		return nil, fmt.Errorf("multiple files found for path: %s", filePath)
	}

	fileInfo := &fileInfos[0]

	//only accept files
	if fileInfo.Type != "regular" {
		return nil, ErrNoFile
	}

	//format time and size values
	fileInfo.Mtime = helpers.ToHumanReadableTime(fileInfo.Mtime)
	fileInfo.ATime = helpers.ToHumanReadableTime(fileInfo.ATime)
	fileInfo.CTime = helpers.ToHumanReadableTimeDiff(fileInfo.CTime)
	fileInfo.Size = helpers.ToHumanReadableFileSize(fileInfo.Size)

	return fileInfo, nil
}

// FetchFilePermissions - get file permissions
func (cf *CommandFileInfo) FetchFilePermissions(filePath string) (*PermissionModeInfos, error) {
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
		return nil, ErrNoFile
	} else if len(modePermInfos) == 1 {
		newModPermInfo := modePermInfos[0]

		//only accept files
		if newModPermInfo.Type != "regular" {
			return nil, ErrNoFile
		}

		return &newModPermInfo, nil
	}

	return nil, errors.New("no result found")
}

// FetchFileType - get file types
func (cf *CommandFileInfo) FetchFileType(filePath string) (*FileTypeInfos, error) {
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

// FetchIsFile - check file directory
func (cf *CommandFileInfo) FetchIsFile(filePath string) (bool, error) {
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
		return false, ErrNoFile
	} else if len(fileInfos) == 1 {
		newInfo := fileInfos[0]

		//only accept files
		if newInfo.Type != "regular" {
			return false, ErrNoFile
		}

		return true, nil
	}

	return false, errors.New("no result found")
}

// FetchFileDate - get file date
func (cf *CommandFileInfo) FetchFileDate(filePath string) (*FileDates, error) {
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
		return nil, ErrNoFile
	} else if len(fileInfos) == 1 {

		fileInfo := fileInfos[0]

		//only accept files
		if fileInfo.Type != "regular" {
			return nil, ErrNoFile
		}

		return &FileDates{
			Path:      fileInfo.Path,
			Filename:  fileInfo.Filename,
			Mtime:     helpers.ToHumanReadableTime(fileInfo.Mtime),
			MTimeDiff: helpers.ToHumanReadableTimeDiff(fileInfo.Mtime),
			ATime:     helpers.ToHumanReadableTime(fileInfo.ATime),
			ATimeDiff: helpers.ToHumanReadableTimeDiff(fileInfo.ATime),
			CTime:     helpers.ToHumanReadableTime(fileInfo.CTime),
			CTimeDiff: helpers.ToHumanReadableTimeDiff(fileInfo.CTime),
		}, nil
	}

	return nil, errors.New("no result found")
}

// FetchFileIsModified - get file date
func (cf *CommandFileInfo) FetchFileIsModified(filePath string) (*FileModified, error) {
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
		return nil, ErrNoFile
	} else if len(fileInfos) == 1 {

		fileInfo := fileInfos[0]

		//only accept files
		if fileInfo.Type != "regular" {
			return nil, ErrNoFile
		}

		return &FileModified{
			Path:      fileInfo.Path,
			Filename:  fileInfo.Filename,
			CTime:     helpers.ToHumanReadableTime(fileInfo.CTime),
			CTimeDiff: helpers.ToHumanReadableTimeDiff(fileInfo.CTime),
		}, nil
	}

	return nil, errors.New("no result found")
}

// ExecuteCommand -  executes a given command
func (cf *CommandFileInfo) ExecuteCommand(command string, params map[string]string) (interface{}, error) {
	path, ok := params["path"]
	if !ok {
		return nil, fmt.Errorf("%s requires a 'path' parameter", command)
	}

	// Validate the path before executing the command
	if err := cf.validatePath(path); err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	switch command {
	case "CHECK_DIRECTORY_FILE":
		return cf.FetchFileInfo(path)
	case "CHECK_FILE_PERMISSION":
		return cf.FetchFilePermissions(path)
	case "CHECK_FILE_TYPE":
		return cf.FetchFileType(path)
	case "CHECK_IS_FILE_TYPE":
		return cf.FetchIsFile(path)
	case "CHECK_FILE_DATES":
		return cf.FetchFileDate(path)
	case "CHECK_IF_MODIFIED_FILE":
		return cf.FetchFileIsModified(path)
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}
