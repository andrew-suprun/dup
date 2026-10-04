package app

import "dup/fs"

type archive struct {
	fsys            fs.FS
	idx             int
	metas           map[string]fs.FileMeta
	copyCommands    []fs.Copy
	updateMetasFile bool
}

func newArchive(idx int, fsys fs.FS) *archive {
	return &archive{fsys: fsys, idx: idx, metas: map[string]fs.FileMeta{}}
}
