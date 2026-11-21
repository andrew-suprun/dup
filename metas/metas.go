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

func (parent *Folder) AddSubfolder(name string) *Folder {
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

func (parent *Folder) AddFile(name string, size int64, modTime time.Time, hash string) *File {
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

	if _, ok := folder.Files[name]; ok {
		panic("Cannot GetSubfolder(). A file with the same name exists.")
	}
	if subfolder, ok := folder.Subfolders[name]; ok {
		return subfolder
	}
	return folder.AddSubfolder(name)
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
	folder.Remove()
	parentFolder := folder
	for parentFolder.Parent != nil {
		parentFolder = parentFolder.Parent
	}
	for _, name := range path[:len(path)-1] {
		parentFolder = parentFolder.GetSubfolder(name)
	}
	name := path[len(path)-1]
	if _, ok := parentFolder.Subfolders[name]; ok {
		panic("Cannot move Folder. A Folder with the same name exists.")
	}
	if _, ok := parentFolder.Files[name]; ok {
		panic("Cannot move Folder. A File with the same name exists.")
	}
	parentFolder.Subfolders[name] = folder
	folder.Name = name
	folder.Parent = parentFolder
}

// func (folder *Folder) FindByPath(path []string) *Folder {
// 	if len(path) == 0 {
// 		return folder
// 	}
// 	name := path[0]
// 	for _, childFolder := range folder.Subfolders {
// 		if childFolder.Name == name {
// 			return childFolder.FindByPath(path[1:])
// 		}
// 	}
// 	return nil
// }

// func (folder *Folder) FindBy(f func(*Folder) bool) *Folder {
// 	if f(folder) {
// 		return folder
// 	}
// 	for _, childFolder := range folder.Subfolders {
// 		if f(childFolder) {
// 			return childFolder.FindBy(f)
// 		}
// 	}
// 	return nil
// }

func (file *File) Path() []string {
	return append(file.Parent.Path(), file.Name)
}

func (file *File) Remove() {
	delete(file.Parent.Files, file.Name)
}

func (file *File) Move(path []string) {
	file.Remove()
	parentFolder := file.Parent
	for parentFolder.Parent != nil {
		parentFolder = parentFolder.Parent
	}
	for _, name := range path[:len(path)-1] {
		parentFolder = parentFolder.GetSubfolder(name)
	}
	name := path[len(path)-1]
	if _, ok := parentFolder.Subfolders[name]; ok {
		panic("Cannot move File. A Folder with the same name exists.")
	}
	if _, ok := parentFolder.Files[name]; ok {
		panic("Cannot move File. A File with the same name exists.")
	}
	parentFolder.Files[name] = file
	file.Name = name
	file.Parent = parentFolder
}

type FolderScanned struct {
	Idx            int
	FilesScanned   int
	FoldersScanned int
}
