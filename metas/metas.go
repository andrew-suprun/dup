package metas

import (
	"path/filepath"
	"time"
)

type Folder struct {
	Parent     *Folder
	Name       string
	Id         string
	Subfolders Folders
	Files      Files
}

func (file *Folder) Path() string {
	if file.Parent == nil {
		return file.Name
	}
	return filepath.Join(file.Parent.Path(), file.Name)
}

type Folders []*Folder

type File struct {
	Parent  *Folder
	Name    string
	Size    int64
	ModTime time.Time
	Hash    string
}

func (file *File) Path() string {
	return filepath.Join(file.Parent.Path(), file.Name)
}

type Files []*File

type FolderScanned struct {
	Root           string
	FilesScanned   int
	FoldersScanned int
}
