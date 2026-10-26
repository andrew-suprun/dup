package main

import (
	"fmt"
	"time"
)

type Backend interface {
	LoadMetas() FileMetas
	StoreMetas(FileMetas)
	ScanArchive() (FileMetas, error)
	HashFile(path string) (string, error)
	MoveFile(src, dest string) error
	OpenFile(path string) error
	ReadFile([]byte) (int, error)
	CreateFile(path string) error
	WriteFile([]byte) error
	CloseFile() error
	RemoveFile(string)
}

func (meta FileMeta) String() string {
	return fmt.Sprintf("%q hash:%s", meta.Path, meta.Hash)
}

// Events

type FileMetas []FileMeta
