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

func parseMeta(content string) []fs.FileMeta {
	metas := make([]fs.FileMeta, 0)
	records, err := csv.NewReader(strings.NewReader(content)).ReadAll()
	if err != nil || len(records) == 0 {
		return metas
	}

	for _, record := range records[1:] {
		if len(record) == 5 {
			iNode, er1 := strconv.ParseUint(record[0], 10, 64)
			path := record[1]
			size, er2 := strconv.ParseUint(record[2], 10, 64)
			modTime, er3 := time.Parse(time.RFC3339, record[3])
			modTime = modTime.UTC().Round(time.Second)
			hash := record[4]
			if hash == "" || er1 != nil || er2 != nil || er3 != nil {
				continue
			}

			metas = append(metas, fs.FileMeta{
				Inode:   iNode,
				Path:    path,
				Size:    int(size),
				ModTime: modTime,
				Hash:    hash,
			})
		}
	}
	return metas
}
