package fs

import (
	"fmt"
	"time"
)

type FS interface {
	Root() string
	Scan(events chan any)
	Rename(dest, src string)
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

// Commands

type Copy struct {
	Path    string
	Hash    string
	ToRoots []string
}

type Rename struct {
	SourcePath      string
	DestinationPath string
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

type RenamingFile struct {
	Idx  int
	Path string
}

type CopyingFile struct {
	Idx  int
	Path string
	Size int
}

type Synced struct {
	Idx int
}
