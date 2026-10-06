package fs

import (
	"fmt"
	"time"
)

type FS interface {
	Scan()
	Rename(dest, src string) error
	Copy(roots []string, src string)
	StoreMetas(FileMetas)
}

type FileMeta struct {
	Path    string
	Size    int
	ModTime time.Time
	Hash    string
}

func (meta FileMeta) String() string {
	return fmt.Sprintf("%q hash:%s", meta.Path, meta.Hash)
}

// Events

type FileMetas struct {
	Idx   int
	Metas []FileMeta
}

type FileHashed struct {
	Idx  int
	Path string
	Hash string
}

type ArchiveHashed struct {
	Idx int
}

type CopyingFile struct {
	Idx  int
	Path string
	Size int
}

type FileCopied struct {
	Idx  int
	Path string
	Size int
}
