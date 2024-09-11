package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
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

type Command struct {
	Type string
	Data interface{}
}

var (
	commandQueue = make(chan Command, 100)
	directory    = "/home/nate/Desktop/test" // monitoring files directory
)

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

	var wg sync.WaitGroup
	wg.Add(2) //specify my two threads

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

	wg.Wait()

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

// fetchFileInfo - get file info from querying the path returning fileInfo
func fetchFileInfo(filePath string) *FileInfo {

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
				fileInfo := fetchFileInfo(filePath)
				if fileInfo != nil {
					// print the result file information
					fmt.Printf("Uuid: %s\nPath: %s\nDirectory: %s\nFilename: %s\nLast Modified: %s\nLast Visit: %s\nLast Change: %s\nSize: %s\nType: %s\nMode: %s\n\n",
						fileInfo.Uuid, fileInfo.Path, fileInfo.Directory, fileInfo.Filename, fileInfo.Mtime, fileInfo.ATime, fileInfo.CTime, fileInfo.Size, fileInfo.Type, fileInfo.Mode)
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
- It then calls the commandQueue "CHECK_FILE" command for each file
*/
func timerThread() {
	//check this thread every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		files, err := filepath.Glob(filepath.Join(directory, "*"))
		if err != nil {
			fmt.Println("Error reading directory:", err)
			continue
		}

		//range the files found calling check directory files command looping each file
		for _, file := range files {
			commandQueue <- Command{
				Type: "CHECK_DIRECTORY_FILES",
				Data: file,
			}
		}
	}
}
