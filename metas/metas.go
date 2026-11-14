package metas

import "time"

type FileMeta struct {
	Path    string
	Size    int64
	ModTime time.Time
	Hash    string
}

type FileMetas []FileMeta
