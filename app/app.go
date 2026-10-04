package app

import (
	"dup/fs"
	"fmt"
	"log"
	"os"

	"os/signal"
)

func Run(fss []fs.FS) {
	app := makeApp(fss)
	defer app.shutdown()

	app.scan()
}

type appState struct {
	archives   []archive
	events     chan any
	interrupts chan os.Signal
}

func makeApp(fss []fs.FS) appState {
	var app appState
	app.archives = make([]archive, len(fss))
	for i := range app.archives {
		app.archives[i] = makeArchive(i, fss[i])
	}

	app.interrupts = make(chan os.Signal, 1)
	signal.Notify(app.interrupts, os.Interrupt)
	return app
}

func (app *appState) scan() {
	scanEvents := make(chan any)
	for _, arc := range app.archives {
		go arc.Fs.Scan(scanEvents)
	}

	scanned := 0
scan:
	for {
		select {
		case e := <-scanEvents:
			switch e := e.(type) {
			case fs.FileMetas:
				fmt.Printf("Archive scanned: %#v\n", e.Idx)
				app.archives[e.Idx] = archive{Idx: e.Idx, Metas: e.Metas}
			case fs.FileHashed:
				fmt.Printf("File hashed: %d: %q - %s\n", e.Idx, e.Path, e.Hash)
				app.archives[e.Idx].UpdateMetasFile = true
			case fs.ArchiveHashed:
				fmt.Printf("Archive hashed: %d\n", e.Idx)
				scanned++
				if scanned == len(app.archives) {
					break scan
				}
			default:
				log.Printf("### Scan event: %v\n", e)
				panic(e)
			}
		case <-app.interrupts:
			signal.Reset(os.Interrupt)
			fmt.Println("Interrupted")
			log.Println("Interrupted")
			break scan
		}
	}
}

func (app *appState) shutdown() {
	fmt.Println("Shutting Down")
}
