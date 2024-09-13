package main

//
//import (
//	"encoding/json"
//	"fmt"
//	"log"
//	"net/http"
//	"sync"
//	"time"
//)
//
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
//var (
//	isRunning bool
//	mu        sync.Mutex //ensure thread safe access to my variables
//	logs      []FileInfo
//)
//
//func main() {
//	http.HandleFunc("/start", startHandler)
//	http.HandleFunc("/stop", stopHandler)
//	http.HandleFunc("/logs", logsHandler)
//
//	log.Println("Starting server on :4070")
//	log.Fatal(http.ListenAndServe(":4070", nil))
//}
//
//// startHandler - starts the file monitoring service
//func startHandler(w http.ResponseWriter, r *http.Request) {
//	mu.Lock()
//	defer mu.Unlock()
//
//	if isRunning {
//		http.Error(w, "Service is already running", http.StatusBadRequest)
//		return
//	}
//
//	isRunning = true
//	go monitorFiles()
//
//	w.WriteHeader(http.StatusOK)
//	fmt.Fprint(w, "File monitoring service started")
//}
//
//// stopHandler -stopping the file monitoring service
//func stopHandler(w http.ResponseWriter, r *http.Request) {
//	mu.Lock()
//	defer mu.Unlock()
//
//	if !isRunning {
//		http.Error(w, "service is not running", http.StatusBadRequest)
//		return
//	}
//
//	isRunning = false
//
//	w.WriteHeader(http.StatusOK)
//	fmt.Fprint(w, "file monitoring service stopped")
//}
//
//// logsHandler -log result
//func logsHandler(w http.ResponseWriter, r *http.Request) {
//	mu.Lock()
//	defer mu.Unlock()
//
//	w.Header().Set("Content-Type", "application/json")
//	json.NewEncoder(w).Encode(logs)
//}
//
//func monitorFiles() {
//	for isRunning {
//		mu.Lock()
//		// simulating file monitoring
//		newLog := FileInfo{ //test
//			Uid:       "123",
//			Path:      "/path/to/file.txt",
//			Directory: "/path/to",
//			Filename:  "file.txt",
//			Mtime:     time.Now().Format(time.RFC3339),
//			Size:      "1024",
//			Type:      "file",
//		}
//		logs = append(logs, newLog)
//		mu.Unlock()
//
//		time.Sleep(5 * time.Second)
//	}
//}
