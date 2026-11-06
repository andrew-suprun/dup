package backend

import (
	"fmt"
	"time"
)

type Backend interface {
	LoadMetas(base string) string
	// StoreMetas(base, content string)
	ScanFolder(path string) (FolderMeta, error)
	HashFile(path string) (string, error)
	// MoveFile(src, dest string) error
	// OpenFile(path string) error
	// ReadFile([]byte) (int, error)
	// CreateFile(path string) error
	// WriteFile([]byte) error
	// CloseFile() error
	// RemoveFile(string)
}

type FileMeta struct {
	Inode   uint64
	Path    string
	Size    int
	ModTime time.Time
}

type FolderMeta struct {
	Files   []FileMeta
	Folders []string
}

func (meta FileMeta) String() string {
	return fmt.Sprintf("%q", meta.Path)
}

// Events

type FileMetas []FileMeta
