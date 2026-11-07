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

	app.run()
}

func (app *App) run() {
	fmt.Println("App is running, commands", len(app.commands))
	for _, cmd := range app.commands {
		cmd <- archive.ScanArchive{}
	}
	fmt.Println("App sent scan commands")
	for {
		event := <-app.events
		fmt.Printf("Event %T %#v\n", event, event)
		switch event := event.(type) {
		case archive.ScannedFiles:
		case archive.ScannedArchive:
			fmt.Println("Event: archive.ScannedArchive", event.Idx)
			for _, meta := range event.Metas {
				fmt.Println(meta)
			}
		}
	}
}
