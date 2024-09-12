package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type MyWindow struct {
	*walk.MainWindow
	logView   *walk.TextEdit
	startBtn  *walk.PushButton
	stopBtn   *walk.PushButton
	isRunning bool
}

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

func main() {
	mw := &MyWindow{isRunning: false}

	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "File Monitor Service",
		MinSize:  Size{600, 400},
		Layout:   VBox{},
		Children: []Widget{
			PushButton{
				AssignTo: &mw.startBtn,
				Text:     "Start",
				OnClicked: func() {
					mw.startService()
				},
			},
			PushButton{
				AssignTo: &mw.stopBtn,
				Text:     "Stop",
				Enabled:  false,
				OnClicked: func() {
					mw.stopService()
				},
			},
			TextEdit{
				AssignTo: &mw.logView,
				ReadOnly: true,
				VScroll:  true,
			},
		},
	}.Create()); err != nil {
		fmt.Println("Error:", err)
		return
	}

	mw.Run()
}

func (mw *MyWindow) startService() {
	resp, err := http.Get("http://localhost:4000/start")
	if err != nil {
		walk.MsgBox(mw, "Error", "Failed to start service: "+err.Error(), walk.MsgBoxIconError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		mw.isRunning = true
		mw.startBtn.SetEnabled(false)
		mw.stopBtn.SetEnabled(true)
		mw.updateLogs()
	} else {
		walk.MsgBox(mw, "Error", "Failed to start service", walk.MsgBoxIconError)
	}
}

func (mw *MyWindow) stopService() {
	resp, err := http.Get("http://localhost:4000/stop")
	if err != nil {
		walk.MsgBox(mw, "Error", "Failed to stop service: "+err.Error(), walk.MsgBoxIconError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		mw.isRunning = false
		mw.startBtn.SetEnabled(true)
		mw.stopBtn.SetEnabled(false)
	} else {
		walk.MsgBox(mw, "Error", "Failed to stop service", walk.MsgBoxIconError)
	}
}

func (mw *MyWindow) updateLogs() {
	go func() {
		for mw.isRunning {
			resp, err := http.Get("http://localhost:4000/logs")
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			var logs []FileInfo
			err = json.Unmarshal(body, &logs)
			if err != nil {
				continue
			}

			mw.Synchronize(func() {
				mw.logView.SetText("")
				for _, log := range logs {
					mw.logView.AppendText(fmt.Sprintf("File: %s, Modified: %s\r\n", log.Path, log.Mtime))
				}
			})

			time.Sleep(5 * time.Second)
		}
	}()
}
