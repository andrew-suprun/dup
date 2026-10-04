package app

import "dup/fs"

type archive struct {
	Fs              fs.FS
	Idx             int
	Metas           []fs.FileMeta
	UpdateMetasFile bool
}

func makeArchive(idx int, fs fs.FS) archive {
	return archive{Idx: idx, Fs: fs}
}
