package main

type MetasLoaded struct {
	Idx   int
	Metas FileMetas
}

type MetasStored struct {
	Idx int
}

type MetasScanned struct {
	Idx   int
	Metas FileMetas
}

type FileHashed struct {
	Idx  int
	Path string
	Hash string
}

type FileMoved struct {
	Idx  int
	Src  string
	Dest string
}

type FileOpened struct {
	Idx  int
	Path string
}

type FileRead struct {
	Idx  int
	Size int
}

type FileCreated struct {
	Idx  int
	Path string
}

type FileWritten struct {
	Idx  int
	Size int
}

type FileClosed struct {
	Idx int
}

type FileRemoved struct {
	Idx  int
	Path string
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

type Quit struct{}
