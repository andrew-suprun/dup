package main

import (
	"encoding/csv"
	"strconv"
	"strings"
	"time"
)

type AppState struct {
	archiveStates []*ArchiveState
	archiveEvents <-chan any
	uiEvents      chan<- any
	quit          <-chan struct{}
}

type ArchiveState struct {
	commands chan<- any
	metas    []FileMeta
}

func run(commands []chan<- any, events <-chan any, uiEvents chan<- any, quit <-chan struct{}) {
	archiveStates := make([]*ArchiveState, len(commands))
	app := &AppState{
		uiEvents:      uiEvents,
		quit:          make(chan struct{}),
		archiveStates: archiveStates,
	}
	for i, cmd := range commands {
		app.archiveStates[i] = &ArchiveState{
			commands: cmd,
		}
	}

	for _, archive := range app.archiveStates {
		archive.commands <- LoadMeta{}
	}
	for {
		event := <-app.archiveEvents
		switch event := event.(type) {
		case Meta:
			app.archiveStates[event.id].metas = parseMeta(event.content)
		}
	}
}

func parseMeta(content string) []FileMeta {
	metas := make([]FileMeta, 0)
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

			metas = append(metas, FileMeta{
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
