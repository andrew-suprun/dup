package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type app struct {
	state    appState
	backup   string
	archives []*archive
	lc       *Lifecycle
	err      error
}

type appState struct {
	stage appStage
}

type appStage int

const (
	appScanning appStage = iota
	appCopying
)

type archive struct {
	fsys  FS
	state *archiveState
	lc    *Lifecycle
}

type archiveState struct {
	metas       FileMetas
	stage       archiveStage
	totalFiles  int
	filesToHash int
	hashedFiles int
}

type archiveStage int

const (
	scanning archiveStage = iota
	hashing
	copying
)

// UI events
type ScanComplete struct {
	TotalFiles  int
	FilesToHash int
}

func Run(fss []FS, backup string) {
	lc := &Lifecycle{}
	archives := make([]*archive, len(fss))
	for i, fs := range fss {
		archives[i] = newArchive(i, fs, lc)
	}
	app := app{
		state:    appState{},
		backup:   backup,
		archives: archives,
		lc:       lc,
	}
	defer app.shutdown()

	go app.ui()
	app.scan()
	app.analyzeArchives()
}

func (app *app) scan() {
	wg := sync.WaitGroup{}
	wg.Add(len(app.archives))
	for _, arc := range app.archives {
		go func() {
			metas, err := arc.fsys.ScanArchive()
			if err != nil {
				app.err = err
				app.lc.Stop()
			}
			arcState := <-arc.state
			arcState.metas = metas
			for i := range metas {
				meta := &metas[i]
				arcState.totalFiles++
				if meta.Hash == "" {
					arcState.filesToHash++
				}
			}
			arcState.stage = hashing
			arc.state <- arcState

			for {
				arcState = <-arc.state
				for i := range arcState.metas {
					meta := &arcState.metas[i]
					// TODO
					_ = meta
				}
			}

			wg.Done()
		}()
	}
	wg.Wait()

	scanned := 0
	for {
		if app.lc.ShoudStop() {
			return
		}
		select {
		case e := <-app.events:
			switch e := e.(type) {
			case FileMetas:
				fmt.Printf("Archive scanned: %#v\n", e.Idx)
				archive := app.archives[e.Idx]
				archive.metas = make([]fileMeta, len(e.Metas))
				for i := range e.Metas {
					archive.metas[i] = fileMeta{FileMeta: e.Metas[i], state: initial}
				}
			case FileHashed:
				fmt.Printf("File hashed: %d: %q - %s\n", e.Idx, e.Path, e.Hash)
				archive := app.archives[e.Idx]
				for i := range archive.metas {
					if archive.metas[i].Path == e.Path {
						archive.metas[i].Hash = e.Hash
						break
					}
				}
				archive.updateMetasFile = true
			case ArchiveHashed:
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
