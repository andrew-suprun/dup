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

func AddFolder(parent *Folder, name string) *Folder {
	folder := &Folder{
		Parent:     parent,
		Name:       name,
		Subfolders: Folders{},
		Files:      Files{},
	}
	if parent != nil {
		parent.Subfolders[name] = folder
	}
	return folder
}

func AddFile(parent *Folder, name string, size int64, modTime time.Time, hash string) *File {
	file := &File{
		Parent:  parent,
		Name:    name,
		Size:    size,
		ModTime: modTime,
		Hash:    hash,
	}
	if parent != nil {
		parent.Files[name] = file
	}
	return file
}

func (folder *Folder) GetSubfolder(name string) *Folder {
	if subfolder, ok := folder.Subfolders[name]; ok {
		return subfolder
	}
	return AddFolder(folder, name)
}

func (folder *Folder) Path() []string {
	if folder.Parent == nil {
		return []string{folder.Name}
	}
	return append(folder.Parent.Path(), folder.Name)
}

func (folder *Folder) Remove() {
	delete(folder.Parent.Subfolders, folder.Name)
}

func (folder *Folder) Move(path []string) {

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

func (file *File) Remove() {
	delete(file.Parent.Files, file.Name)
}

func (file *File) Move(path []string) {

}

type FolderScanned struct {
	Idx            int
	FilesScanned   int
	FoldersScanned int
}
