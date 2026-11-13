package backend

import (
	"fmt"
	"time"
)

type Backend interface {
	ScanFolder(path string) (FolderMeta, error)
	MoveFile(src, dest string) error
	// OpenFile(path string) error
	// ReadFile([]byte) (int, error)
	// CreateFile(path string) error
	// WriteFile([]byte) error
	// CloseFile() error
	// RemoveFile(string)
}

type FileMeta struct {
	Path    string
	Size    int64
	ModTime time.Time
	Hash    string
}

type FileMetas []FileMeta

type FolderMeta struct {
	Metas   FileMetas
	Folders []string
}

func (meta FileMeta) String() string {
	return fmt.Sprintf("%q", meta.Path)
}
