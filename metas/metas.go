package metas

import "time"

type FolderMeta struct {
	Path string
	Id   string
}

type FolderMetas []FolderMeta

type FileMeta struct {
	Path    string
	Size    int64
	ModTime time.Time
	Hash    string
}

type FileMetas []FileMeta
