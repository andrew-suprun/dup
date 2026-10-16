package main

import (
	"fmt"
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

func newArchive(idx int, fsys FS, lc *Lifecycle) *archive {
	return &archive{fsys: fsys, idx: idx, state: make(chan archiveState, 1), lc: lc}
}
