package main

type LoadMetas struct{}
type StoreMetas FileMetas
type ScanArchive struct{}
type HashFile struct {
	Path string
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
