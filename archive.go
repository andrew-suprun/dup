package main

import (
	"fmt"
	"io/fs"
)

func runArchive(idx int, fsys fs.FS, backup string, commands <-chan any, events chan<- any, quit <-chan struct{}) {
	archive := &Archive{
		idx:      idx,
		fsys:     fsys,
		backup:   backup,
		commands: commands,
		events:   events,
		quit:     quit,
	}

	for {
		select {
		case cmd := <-archive.commands:
			archive.handle(cmd)
		case <-quit:
			archive.shutdown()
			return
		}
	}
}

func (arcive *Archive) handle(cmd any) {

}

func (arcive *Archive) shutdown() {

}

type Archive struct {
	idx      int
	fsys     fs.FS
	backup   string
	commands <-chan any
	events   chan<- any
	quit     <-chan struct{}
}

type archiveStage int

const (
	scanning archiveStage = iota
	hashing
	copying
)

type fileMeta struct {
	FileMeta
	state fileState
}

func (meta fileMeta) String() string {
	return fmt.Sprintf("%v %v", meta.FileMeta, meta.state)
}

type fileState int

const (
	initial fileState = iota
	backedUp
	toCopy
	done
)

func (state fileState) String() string {
	switch state {
	case initial:
		return "initial"
	case backedUp:
		return "backedUp"
	case toCopy:
		return "toCopy"
	case done:
		return "done"
	}
	return ""
}
