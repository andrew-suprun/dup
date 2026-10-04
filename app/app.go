package app

import (
	"dup/archive"
	"dup/fs"
	"fmt"
	"log"
	"os"

	"os/signal"
)

type appState struct {
	archives []archive.Archive
}

func Run(fss []fs.FS) {
	var app appState
	app.archives = make([]archive.Archive, len(fss))
	for i := range app.archives {
		app.archives[i] = archive.MakeArchive(i, fss[i].Root())
	}

	interruptEvents := make(chan os.Signal, 1)
	signal.Notify(interruptEvents, os.Interrupt)
	defer signal.Reset(os.Interrupt)

	// Scanning
	scanEvents := make(chan any)
	for _, fs := range fss {
		go fs.Scan(scanEvents)
	}

	scanned := 0
scan:
	for {
		select {
		case e := <-scanEvents:
			switch e := e.(type) {
			case fs.FileMetas:
				fmt.Printf("Archive scanned: %#v\n", e.Idx)
				app.archives[e.Idx] = archive.Archive{Idx: e.Idx, Metas: e.Metas}
			case fs.FileHashed:
				fmt.Printf("File hashed: %d: %q - %s\n", e.Idx, e.Path, e.Hash)
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
		case <-interruptEvents:
			fmt.Println("Interrupted")
			log.Println("Interrupted")
			break scan
		}
	}
}
