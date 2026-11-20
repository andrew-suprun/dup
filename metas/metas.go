package metas

import (
	"time"
)

type Folder struct {
	Parent     *Folder
	Name       string
	Id         string
	Subfolders Folders
	Files      Files
}

type Folders map[string]*Folder

type File struct {
	Parent  *Folder
	Name    string
	Size    int64
	ModTime time.Time
	Hash    string
}

type Files map[string]*File

func NewFolder(parent *Folder, name string) *Folder {
	return &Folder{
		Parent:     parent,
		Name:       name,
		Subfolders: Folders{},
		Files:      Files{},
	}
}

func NewFile(parent *Folder, name string, size int64, modTime time.Time, hash string) *File {
	return &File{
		Parent:  parent,
		Name:    name,
		Size:    size,
		ModTime: modTime,
		Hash:    hash,
	}
}

func (folder *Folder) Path() []string {
	if folder.Parent == nil {
		return []string{folder.Name}
	}
	return append(folder.Parent.Path(), folder.Name)
}

func (folder *Folder) FindByPath(path []string) *Folder {
	if len(path) == 0 {
		return folder
	}
	name := path[0]
	for _, childFolder := range folder.Subfolders {
		if childFolder.Name == name {
			return childFolder.FindByPath(path[1:])
		}
	}
	return nil
}

func (folder *Folder) FindBy(f func(*Folder) bool) *Folder {
	if f(folder) {
		return folder
	}
	for _, childFolder := range folder.Subfolders {
		if f(childFolder) {
			return childFolder.FindBy(f)
		}
	}
	return nil
}

func (file *File) Path() []string {
	return append(file.Parent.Path(), file.Name)
}

type FolderScanned struct {
	Idx            int
	FilesScanned   int
	FoldersScanned int
}
