package app

import (
	"dup/fs"
	"dup/lifecycle"
	"fmt"
	"log"
	"time"
)

func Run(fss []fs.FS, lc *lifecycle.Lifecycle) {
	app := makeApp(fss, lc)
	defer app.shutdown()

	app.scan()
	app.analyzeArchives()
}

type app struct {
	lc       *lifecycle.Lifecycle
	archives []*archive
	events   chan any
	backup   string
}

func makeApp(fss []fs.FS, lc *lifecycle.Lifecycle) app {
	var app = app{
		lc:       lc,
		archives: make([]*archive, len(fss)),
		events:   make(chan any, len(fss)),
		backup:   time.Now().Format("~~~060102-150405~~~"),
	}

	for i := range app.archives {
		app.archives[i] = newArchive(i, fss[i])
	}

	return app
}

func (app *app) scan() {
	for _, arc := range app.archives {
		go arc.fsys.Scan(app.events)
	}

	scanned := 0
scan:
	for {
		if app.lc.ShoudStop() {
			return
		}
		e := <-app.events
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
	}
}

func (app *app) shutdown() {
	fmt.Println("Shutting Down")
}
