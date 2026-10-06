package app

import (
	"dup/fs"
	"fmt"
)

type archive struct {
	idx             int
	fsys            fs.FS
	events          chan any
	metas           []fileMeta
	state           archiveState
	updateMetasFile bool
}

type fileMeta struct {
	fs.FileMeta
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

type archiveState int

const (
	scanning archiveState = iota
	hashing
	copying
)

func newArchive(idx int, fsys fs.FS, events chan any) *archive {
	return &archive{fsys: fsys, idx: idx, events: events}
}
