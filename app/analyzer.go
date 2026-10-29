package app

import (
	"fmt"
)

func (app *App) analyzeArchives() {
	if app.shoudStop() {
		return
	}
	fmt.Printf("---- analizing\n")
	for i, archive := range app.archives {
		fmt.Printf("Archive [%d]\n", i)
		for _, meta := range archive.metas {
			fmt.Printf("  ino: %d  path: %q  size: %d  hash: %s\n", meta.Inode, meta.Path, meta.Size, meta.Hash)
		}
	}

	// TODO Analisys

	fmt.Printf("---- analized\n")
	for i, archive := range app.archives {
		fmt.Printf("Archive [%d]\n", i)
		for _, meta := range archive.metas {
			fmt.Printf("  ino: %d  path: %q  size: %d  hash: %s\n", meta.Inode, meta.Path, meta.Size, meta.Hash)
		}
	}
}

func (app *App) shoudStop() bool {
	select {
	case <-app.quit:
		return true
	default:
		return false
	}
}
