package app

import (
	"dup/fs"
	"encoding/csv"
	"strconv"
	"strings"
	"time"
)

type App struct {
	archives []*Archive
	events   <-chan any
	quit     <-chan struct{}
	purge    string
}

type Archive struct {
	commands chan<- any
	metas    []fs.FileMeta
}

func run(commands []chan<- any, events <-chan any, purge string, quit <-chan struct{}) {
	archives := make([]*Archive, len(commands))
	app := &App{
		archives: archives,
		events:   events,
		purge:    purge,
		quit:     make(chan struct{}),
	}
	for i, cmd := range commands {
		app.archives[i] = &Archive{
			commands: cmd,
		}
	}

	for _, archive := range app.archives {
		archive.commands <- fs.LoadMeta{}
	}
	for {
		event := <-app.events
		switch event := event.(type) {
		case fs.Meta:
			app.archives[event.Id].metas = parseMeta(event.Content)
		}
	}
}
