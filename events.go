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
