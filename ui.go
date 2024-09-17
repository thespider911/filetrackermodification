package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Service struct {
	running bool
	logs    []string
}

func (s *Service) start() {
	s.running = true
	go func() {
		for s.running {
			s.logs = append(s.logs, fmt.Sprintf("Log entry at %s", time.Now().Format("15:04:05")))
			time.Sleep(time.Second)
		}
	}()
}

func (s *Service) stop() {
	s.running = false
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Service Control")

	service := &Service{}

	logs := widget.NewMultiLineEntry()
	logs.Disable()

	updateLogs := func() {
		logs.SetText("")
		for _, log := range service.logs {
			logs.Text += log + "\n"
		}
		logs.Refresh()
	}

	startButton := widget.NewButton("Start Service", func() {
		if !service.running {
			service.start()
			updateLogs()
		}
	})

	stopButton := widget.NewButton("Stop Service", func() {
		if service.running {
			service.stop()
			updateLogs()
		}
	})

	buttons := container.NewHBox(startButton, stopButton)
	content := container.NewVBox(buttons, logs)

	go func() {
		for {
			time.Sleep(time.Second)
			if service.running {
				updateLogs()
			}
		}
	}()

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(400, 300))
	myWindow.ShowAndRun()
}
