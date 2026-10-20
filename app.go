package main

import "io/fs"

type AppState struct {
	archives      []ArchiveState
	backup        string
	archiveEvents <-chan any
	uiEvents      chan<- any
	quit          <-chan struct{}
}

type ArchiveState struct {
	root     string
	commands chan<- any
}

func run(roots []string, backup string, fsys fs.FS, uiEvents chan<- any) {
	archiveEvents := make(chan any, len(roots))
	app := &AppState{
		backup:        backup,
		uiEvents:      uiEvents,
		quit:          make(chan struct{}),
		archiveEvents: archiveEvents,
	}
	app.archives = make([]ArchiveState, len(roots))
	for i, root := range roots {
		archiveCommands := make(chan any, 1)
		app.archives[i] = ArchiveState{
			root:     root,
			commands: archiveCommands,
		}
		go runArchive(i, fsys, archiveCommands, archiveEvents, app.quit)
	}

	for _, archive := range app.archives {
		archive.commands <- ScanArchive{}
	}
	for {
		archiveEvent := <-app.archiveEvents
		_ = archiveEvent
		// TODO
	}
}
