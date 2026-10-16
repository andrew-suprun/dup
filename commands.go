package main

type LoadMetas struct{}
type StoreMetas FileMetas
type ScanArchive struct{}
type HashFile struct {
	Path string
}

// LoadMetas() FileMetas
// StoreMetas(FileMetas)
// ScanArchive() (FileMetas, error)
// HashFile(path string) (string, error)
// MoveFile(src, dest string) error
// OpenFile(path string) error
// ReadFile([]byte) error
// CreateFile(path string) error
// WriteFile([]byte) error
// CloseFile() error
// RemoveFile(string)
