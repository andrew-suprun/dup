package app

import (
	"dup/archive"
	"fmt"
)

type App struct {
	commands []chan<- any
	events   <-chan any
	quit     chan struct{}
	finished []bool
}

func Run(commands []chan<- any, events <-chan any, quit <-chan struct{}) {
	app := &App{
		commands: commands,
		events:   events,
		quit:     make(chan struct{}),
		finished: make([]bool, len(commands)),
	}

	app.run()
}

func (app *App) run() {
	for _, cmd := range app.commands {
		cmd <- archive.ScanArchive{}
	}
	for {
		allFinished := true
		for _, f := range app.finished {
			if !f {
				allFinished = false
				break
			}
		}
		if allFinished {
			fmt.Println("FINISHED")
			return
		}
		fmt.Println("Waiting for event")
		event := <-app.events
		switch event := event.(type) {
		case archive.ScannedFiles:
			fmt.Printf("Event: archive.ScannedFiles %#v\n", event)
		case archive.ScannedArchive:
			fmt.Println("Event: archive.ScannedArchive", event.Idx)
			for _, meta := range event.Metas {
				fmt.Println(meta)
			}

			close(app.commands[event.Idx]) // TODO temporary
		case archive.Finished:
			fmt.Println("Event: archive.Finished", event.Idx)
			app.finished[event.Idx] = true
		default:
			fmt.Printf("### Unknown Event: %T", event)
		}
	}
}
