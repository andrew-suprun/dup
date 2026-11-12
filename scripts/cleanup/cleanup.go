package main

import (
	"dup/dump"
	"fmt"
	"os"
	"path/filepath"
)

type Folder struct {
	Name     string
	Children []*Folder
	Empty    bool
}

func main() {
	tree := collectStats("/Volumes/SanDisk/Fiction/")
	markNonEmptyParents(tree)
	clearEmpty(tree)
	fmt.Println(dump.Dump(tree))
}

func collectStats(root string) *Folder {
	tree := &Folder{Name: root}
	recCollectStats(root, tree)
	return tree
}

func recCollectStats(path string, folder *Folder) {
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			child := &Folder{Name: entry.Name(), Empty: true}
			folder.Children = append(folder.Children, child)
			recCollectStats(filepath.Join(path, entry.Name()), child)
		} else if entry.Name() != ".DS_Store" {
			folder.Empty = false
		}
	}
}

func markNonEmptyParents(folder *Folder) {
	for _, child := range folder.Children {
		markNonEmptyParents(child)
		if !child.Empty {
			folder.Empty = false
		}
	}
}

func clearEmpty(root *Folder) {
	recClearEmpty("", root)
}

func recClearEmpty(path string, folder *Folder) {
	childPath := filepath.Join(path, folder.Name)
	if folder.Empty {
		fmt.Println("Empty", childPath)
		err := os.RemoveAll(childPath)
		if err != nil {
			panic(err)
		}
	} else {
		for _, child := range folder.Children {
			recClearEmpty(childPath, child)
		}
	}
}
