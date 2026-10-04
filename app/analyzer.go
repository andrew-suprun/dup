package app

import (
	"dup/fs"
	"path/filepath"
)

func (app *app) analyzeArchives() {
	app.ignoreIdenticalFiles()
	app.backupExcessFiles()
	app.resolveConflicts()
	app.renameAndCopyFiles()
}

func (app *app) ignoreIdenticalFiles() {
	identicalFiles := []string{}
	for _, original := range app.archives[0].metas {
		hasIdentical := true
		for _, archive := range app.archives[1:] {
			copy, ok := archive.metas[original.Path]
			if !ok || original.Size != copy.Size || original.Hash != copy.Hash {
				hasIdentical = false
			}
		}
		if hasIdentical {
			identicalFiles = append(identicalFiles, original.Path)
		}
	}
	for _, archive := range app.archives {
		for _, path := range identicalFiles {
			delete(archive.metas, path)
		}
	}
}

func (app *app) backupExcessFiles() {
	hashes := map[string]struct{}{}
	for _, archive := range app.archives {
		for _, file := range archive.metas {
			hashes[file.Hash] = struct{}{}
		}
	}

	originals := app.archives[0].byHash()
	for _, archive := range app.archives[1:] {
		copies := archive.byHash()
		for hash := range hashes {
			originalFiles := originals[hash]
			copyFiles := copies[hash]

			if len(originalFiles) >= len(copyFiles) {
				continue
			}
			for i := len(originalFiles); i < len(copyFiles); i++ {
				path := copyFiles[i]
				archive.commands = append(archive.commands, fs.Rename{
					SourcePath:      path,
					DestinationPath: filepath.Join(app.backup, path),
				})
				delete(archive.files, path)
			}
		}
	}
}

func (app *app) resolveConflicts() {
	for _, file := range app.archives[0].files {
		for _, archive := range app.archives[1:] {
			if other, ok := archive.files[file.path]; ok {
				if file.hash == other.hash {
					continue
				}
				dir, name := filepath.Split(other.path)
				newPath := filepath.Join(dir, app.backup+name)
				archive.commands = append(archive.commands, fs.Rename{
					SourcePath:      other.path,
					DestinationPath: newPath,
				})
				other.path = newPath
				archive.files[newPath] = other
				delete(archive.files, file.path)
			}
		}
	}
}

func (app *app) renameAndCopyFiles() {
	toCopy := map[string][]string{}
	originalsByHash := app.archives[0].byHash()
	for _, archive := range app.archives[1:] {
		copiesByHash := archive.byHash()
		for hash, originals := range originalsByHash {
			copies := copiesByHash[hash]
			for i, original := range originals {
				if i < len(copies) {
					archive.commands = append(archive.commands, fs.Rename{
						SourcePath:      copies[i],
						DestinationPath: original,
					})
				} else {
					newRoots := toCopy[original]
					newRoots = append(newRoots, archive.fs.Root())
					toCopy[original] = newRoots
				}
			}
		}
	}
	for path, roots := range toCopy {
		archive := app.archives[0]
		if len(roots) > 0 {
			archive.commands = append(archive.commands, fs.Copy{
				Path:    path,
				Hash:    archive.files[path].hash,
				ToRoots: roots,
			})
		}

	}
}

func (arc *archive) byHash() map[string][]string {
	result := map[string][]string{}
	for _, file := range arc.files {
		paths := result[file.hash]
		paths = append(paths, file.path)
		result[file.hash] = paths
	}
	return result
}
