package main

import (
	"fmt"
	"path/filepath"
)

func (app *AppState) analyzeArchives() {
	if app.lc.ShoudStop() {
		return
	}
	fmt.Printf("---- analizing\n")
	for i, archive := range app.archives {
		fmt.Printf("Archive [%d] %v\n", i, archive.state)
		for _, meta := range archive.state.metas {
			fmt.Printf("  state: %v - %q %s\n", meta.state, meta.Path, meta.Hash)
		}
	}

	app.backupConflicts()
	app.markIdentical()

	fmt.Printf("---- analized\n")
	for i, archive := range app.archives {
		fmt.Printf("Archive [%d] %v\n", i, archive.state)
		for _, meta := range archive.metas {
			fmt.Printf("  state: %v - %q %s\n", meta.state, meta.Path, meta.Hash)
		}
	}
}

func (app *AppState) backupConflicts() {
	for _, archive := range app.archives[1:] {
		for i := range archive.metas {
			copy := &archive.metas[i]
			foundConflict := true
			for metaIdx := range app.archives[0].metas {
				original := &app.archives[0].metas[metaIdx]
				if original.Path == copy.Path {
					if original.Size == copy.Size && original.Hash == copy.Hash {
						foundConflict = false
					}
					break
				}
			}
			if foundConflict {
				if app.lc.ShoudStop() {
					return
				}
				newPath := filepath.Join(app.backup, copy.Path)
				archive.fsys.Move(newPath, copy.Path)
				copy.Path = newPath
				copy.state = backedUp
			}
		}
	}
}

func (app *AppState) markIdentical() {
	for metaIdx := range app.archives[0].metas {
		original := &app.archives[0].metas[metaIdx]

		allIdentical := true
		for _, archive := range app.archives[1:] {
			found := false
			for i := range archive.metas {
				copy := &archive.metas[i]
				if original.Path == copy.Path {
					copy.state = done
					found = true
					break
				}
			}
			if !found {
				for i := range archive.metas {
					copy := &archive.metas[i]
					if copy.state == backedUp && copy.Hash == original.Hash {
						if app.lc.ShoudStop() {
							return
						}
						copy.state = done
						found = true
						archive.fsys.Move(original.Path, copy.Path)
						copy.Path = original.Path
						break
					}
				}
			}
			if !found {
				allIdentical = false
			}
		}
		if allIdentical {
			original.state = done
		}
	}
}
