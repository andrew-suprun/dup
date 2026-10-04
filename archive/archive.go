package archive

import "dup/fs"

type Archive struct {
	Idx   int
	Root  string
	Metas []fs.FileMeta
}

func MakeArchive(idx int, root string) Archive {
	return Archive{Idx: idx, Root: root}
}
