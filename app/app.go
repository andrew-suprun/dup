package app

import (
	"dup/fs"
	"dup/lifecycle"
	"fmt"
	"log"
	"os"
	"time"

	"os/signal"
)

func Run(fss []fs.FS, lc *lifecycle.Lifecycle) {
	app := makeApp(fss, lc)
	defer app.shutdown()

	app.scan()
	app.analyzeArchives()
}

type app struct {
	lc          *lifecycle.Lifecycle
	archives    []*archive
	events      chan any
	backup      string
	interrupts  chan os.Signal
	interrupted bool
}

func makeApp(fss []fs.FS, lc *lifecycle.Lifecycle) app {
	var app = app{
		lc:         lc,
		archives:   make([]*archive, len(fss)),
		backup:     time.Now().Format("~~~060102-150405~~~"),
		interrupts: make(chan os.Signal, 1),
	}

	for i := range app.archives {
		app.archives[i] = newArchive(i, fss[i])
	}

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
				archive := app.archives[e.Idx]
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
