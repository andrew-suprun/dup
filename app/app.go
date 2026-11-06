package app

import (
	"dup/archive"
	"fmt"
)

type App struct {
	commands []chan<- any
	events   <-chan any
	quit     <-chan struct{}
}

func Run(commands []chan<- any, events <-chan any, quit <-chan struct{}) {
	app := &App{
		commands: commands,
		events:   events,
		quit:     make(chan struct{}),
	}

	go app.run()
}

func (app *App) run() {
	for _, cmd := range app.commands {
		cmd <- archive.ScanArchive{}
	}
	for {
		event := <-app.events
		fmt.Println(event)
		switch event := event.(type) {
		case archive.ScannedFiles:
			_ = event
			fmt.Println("Event: archive.ScannedFile")
		}
	}
}
