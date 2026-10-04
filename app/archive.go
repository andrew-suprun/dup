package app

import "dup/fs"

type archive struct {
	fsys            fs.FS
	idx             int
	metas           map[string]fs.FileMeta
	updateMetasFile bool
}

func makeArchive(idx int, fs fs.FS) archive {
	return archive{idx: idx, fsys: fs}
}
