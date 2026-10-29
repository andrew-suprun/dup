package fs

import "time"

type LoadMeta struct{}

type Meta struct {
	Id      int
	Content string
}

type StoreMeta struct {
	Content string
}

type ScanArchive struct{}

type ScannedFiles struct {
	metas []FileMeta
}

type FileMeta struct {
	Inode   uint64
	Path    string
	Size    int
	ModTime time.Time
	Hash    string
}

type HashFile struct {
	Path string
}

type FileHash struct {
	Path string
	Hash string
}

type MoveFile struct {
	Src  string
	Dest string
}
type OpenFile struct {
	Path string
}
type ReadFile struct {
	Buf []byte
}
type CreateFile struct {
	Path string
}
type WriteFile struct {
	Buf []byte
}
type CloseFile struct{}
type RemoveFile struct {
	Path string
}
