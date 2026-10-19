package main

import "io/fs"

type App struct {
	archiveCommands []chan<- any
	archiveEvents   <-chan any
	uiEvents        chan<- any
	quit            <-chan struct{}
}

func run(fss []fs.FS, uiEvents chan<- any, backup string) {
	archiveEvents := make(chan any, len(fss))
	app := &App{
		uiEvents:      uiEvents,
		quit:          make(chan struct{}),
		archiveEvents: archiveEvents,
	}
	for i, fsys := range fss {
		archiveCommands := make(chan any, 1)
		go runArchive(i, fsys, backup, archiveCommands, archiveEvents, app.quit)
		app.archiveCommands[i] = archiveCommands
	}

	for _, cmd := range app.archiveCommands {
		cmd <- ScanArchive{}
	}
	for {
		archiveEvent := <-app.archiveEvents
		_ = archiveEvent
		// TODO
	}
}
