package archive

import (
	"dup/backend"
)

type ScanArchive struct{}

type ScannedFiles struct {
	Idx            int
	FilesScanned   int
	FoldersScanned int
	FoldersToScan  int
}

type ScannedArchive struct {
	Idx   int
	Metas backend.FileMetas
}

type TidyUp struct {
	Metas backend.FileMetas
}

type MissingHashes struct {
	Idx    int
	Hashes []uint64
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

type Finished struct {
	Idx int
}
