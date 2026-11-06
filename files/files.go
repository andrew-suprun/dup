package files

import (
	"dup/backend"
)

type Files struct{}

func (f *Files) LoadMetas(base string) string {
	return ""
}

// func (f *Files) StoreMetas(base, content string) {}

func (f *Files) ScanFolder(path string) (backend.FolderMeta, error) {
	return backend.FolderMeta{}, nil
}

func (f *Files) HashFile(path string) (string, error) {
	return "", nil
}
