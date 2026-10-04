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

type app struct {
	archives    []archive
	events      chan any
	interrupts  chan os.Signal
	interrupted bool
}

func makeApp(fss []fs.FS) app {
	var app app
	app.archives = make([]archive, len(fss))
	for i := range app.archives {
		app.archives[i] = makeArchive(i, fss[i])
	}

	app.interrupts = make(chan os.Signal, 1)
	signal.Notify(app.interrupts, os.Interrupt)
	return app
}

func (app *app) scan() {
	scanEvents := make(chan any)
	for _, arc := range app.archives {
		go arc.fsys.Scan(scanEvents)
	}

	scanned := 0
scan:
	for {
		select {
		case e := <-scanEvents:
			switch e := e.(type) {
			case fs.FileMetas:
				fmt.Printf("Archive scanned: %#v\n", e.Idx)
				archive := &app.archives[e.Idx]
				for _, meta := range e.Metas {
					archive.metas[meta.Path] = meta
				}
			case fs.FileHashed:
				fmt.Printf("File hashed: %d: %q - %s\n", e.Idx, e.Path, e.Hash)
				app.archives[e.Idx].updateMetasFile = true
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
			app.interrupted = true
			fmt.Println("Interrupted")
			log.Println("Interrupted")
			break scan
		}
	}
}

func (app *app) shutdown() {
	fmt.Println("Shutting Down")
}
