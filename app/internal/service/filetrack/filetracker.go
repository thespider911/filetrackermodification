package filetrack

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// FileInfo - file info struct
type FileInfo struct {
	Uid          string `json:"uid"`
	Path         string `json:"path"`
	Directory    string `json:"directory"`
	Filename     string `json:"filename"`
	ModifiedTime string `json:"mtime"`
	AccessedTime string `json:"atime"`
	ChangedTime  string `json:"ctime"`
	FileSize     string `json:"size"`
	FileType     string `json:"type"`
	Permission   string `json:"mode"`
}

type FileTracker struct {
}

// FetchFilesInfo - get files info from querying the path returning fileInfo
func (data FileTracker) FetchFilesInfo(filePath string) (*FileInfo, error) {
	var fileInfos []FileInfo

	// osquery query and command run
	query := fmt.Sprintf("SELECT uid, path, directory, filename, mtime, atime, ctime, size, type, mode FROM file WHERE path = '%s';", filePath)
	cmd := exec.Command("osqueryi", "--json", query)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// decode the output
	err = json.Unmarshal(output, &fileInfos)
	if err != nil {
		return nil, err
	}

	// return all file infos
	if len(fileInfos) > 0 {
		return &fileInfos[0], nil
	}

	return nil, nil
}
