package app

import (
	"dup/fs"
	"dup/lifecycle"
	"fmt"
	"log"
	"os"
	"time"
)

type app struct {
	lc       *lifecycle.Lifecycle
	archives []*archive
	events   chan any
	backup   string
}

func Run(fss []fs.FS, backup string, lc *lifecycle.Lifecycle, events chan any) {
	var app = app{
		lc:       lifecycle.New(),
		archives: make([]*archive, len(fss)),
		events:   events,
		backup:   backup,
	}

	for i := range app.archives {
		app.archives[i] = newArchive(i, fss[i], app.events)
	}

	defer app.shutdown()

	app.scan()
	app.analyzeArchives()
}

func (app *app) scan() {
	for _, arc := range app.archives {
		go arc.fsys.Scan()
	}

	scanned := 0
	for {
		if app.lc.ShoudStop() {
			return
		}
		select {
		case e := <-app.events:
			switch e := e.(type) {
			case fs.FileMetas:
				fmt.Printf("Archive scanned: %#v\n", e.Idx)
				archive := app.archives[e.Idx]
				archive.metas = make([]fileMeta, len(e.Metas))
				for i := range e.Metas {
					archive.metas[i] = fileMeta{FileMeta: e.Metas[i], state: initial}
				}
			case fs.FileHashed:
				fmt.Printf("File hashed: %d: %q - %s\n", e.Idx, e.Path, e.Hash)
				archive := app.archives[e.Idx]
				for i := range archive.metas {
					if archive.metas[i].Path == e.Path {
						archive.metas[i].Hash = e.Hash
						break
					}
				}
				archive.updateMetasFile = true
			case fs.ArchiveHashed:
				fmt.Printf("Archive hashed: %d\n", e.Idx)
				scanned++
				if scanned == len(app.archives) {
					return
				}
			case error:
				fmt.Fprintf(os.Stderr, "Error: %W\n", e)
				app.lc.Stop()
			default:
				log.Printf("### Scan event: %v\n", e)
				panic(e)
			}
		case <-time.After(20 * time.Millisecond):
			// TODO: UI update
		}
	}
}

func (app *app) shutdown() {
	fmt.Println("Shutting Down")
}
