package archive

import "time"

type ScanArchive struct{}

type ScannedFiles struct {
	Idx            int
	FilesScanned   int
	FilesHashed    int
	FoldersScanned int
	FoldersToScan  int
}

type ScannedArchive struct {
	Idx   int
	Metas FileMetas
}

type TidyUp struct {
	Metas FileMetas
}

type MissingHashes struct {
	Idx    int
	Hashes []uint64
}

type FileMetas []FileMeta

type FileMeta struct {
	Inode   uint64
	Path    string
	Size    int64
	ModTime time.Time
	Hash    string
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
