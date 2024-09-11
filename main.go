package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

// FileInfo - file info struct
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

func main() {
	//api end points
	//mux := http.NewServeMux()
	//
	//mux.HandleFunc("/v1/health", handleCheckStatus)
	//mux.HandleFunc("/v1/logs", handleLogs)
	//mux.HandleFunc("/v1/help", handleHelpCommands)
	//
	//if err := http.ListenAndServe(":4000", mux); err != nil {
	//	log.Fatal(err)
	//}

	// monitoring files directory
	directory := "/home/nate/Desktop/test"

	// fetch all files in this directory
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	//range fetching all files
	for _, file := range files {
		if !file.IsDir() {
			fullPath := filepath.Join(directory, file.Name())
			fetchFileInfo(fullPath)
		}
	}
}

// fetchFileInfo - get file info from querying the path
func fetchFileInfo(filePath string) {

	// osquery query and command run
	query := fmt.Sprintf("SELECT uid, path, directory, filename, mtime, atime, ctime, size, type, mode FROM file WHERE path = '%s';", filePath)
	cmd := exec.Command("osqueryi", "--json", query)
	output, err := cmd.Output()

	if err != nil {
		fmt.Println("Error running osquery:", err)
		return
	}

	// decode the output
	var fileInfos []FileInfo
	err = json.Unmarshal(output, &fileInfos)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// print the result file information
	if len(fileInfos) > 0 {
		fileInfo := fileInfos[0]
		fmt.Printf("Uuid: %s\nPath: %s\nDirectory: %s\nFilename: %s\nLast Modified: %s\nLast Visit: %s\nLast Change: %s\nSize: %s\nType: %s\nMode: %s\n\n",
			fileInfo.Uuid, fileInfo.Path, fileInfo.Directory, fileInfo.Filename, fileInfo.Mtime, fileInfo.ATime, fileInfo.CTime, fileInfo.Size, fileInfo.Type, fileInfo.Mode)
	}
}
