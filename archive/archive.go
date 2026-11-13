package archive

import (
	"dup/backend"
	"dup/dump"
	"fmt"
	"path/filepath"
)

type Archive[B backend.Backend] struct {
	idx      int
	base     string
	purge    string
	events   chan<- any
	quit     chan struct{}
	commands <-chan any
	backend  B
	metas    backend.FileMetas
	err      error

	shouldStoreMetas bool
}

type ByHashPath map[string]map[string]*backend.FileMeta

func RunArchive[B backend.Backend](idx int, base, purge string, events chan<- any, quit chan struct{}) chan<- any {
	commands := make(chan any, 2)
	archive := &Archive[B]{
		idx:      idx,
		base:     base,
		purge:    purge,
		events:   events,
		quit:     quit,
		commands: commands,
	}
	go archive.run()
	return commands
}

func (archive *Archive[B]) run() {
	fmt.Println("Archive is running")
mainLoop:
	for {
		select {
		case cmd, ok := <-archive.commands:
			fmt.Printf("Archive got command: %T\n", cmd)
			if !ok {
				break mainLoop
			}
			archive.handleCommand(cmd)
		case <-archive.quit:
			break mainLoop
		}
	}
	archive.shutdown()
}

func (archive *Archive[FS]) handleCommand(cmd any) {
	switch cmd := cmd.(type) {
	case ScanArchive:
		_ = cmd
		archive.scanArchive()
	case TidyUp:
		archive.tidyUp(cmd.Metas)
	}
}

func (archive *Archive[FS]) scanArchive() {
	foldersToScan := []string{archive.base}
	foldersScanned := 0
	for len(foldersToScan) > 0 {
		folderToScan := foldersToScan[0]
		foldersToScan = foldersToScan[1:]
		folder, err := archive.backend.ScanFolder(folderToScan)
		if err != nil {
			archive.err = err
			archive.events <- archive.err
			close(archive.quit)
		}
		archive.metas = append(archive.metas, folder.Metas...)

		foldersScanned++
		foldersToScan = append(foldersToScan, folder.Folders...)
		archive.events <- ScannedFiles{
			Idx:            archive.idx,
			FilesScanned:   len(archive.metas),
			FoldersScanned: foldersScanned,
			FoldersToScan:  len(foldersToScan),
		}
	}

	archive.events <- ScannedArchive{
		Idx:   archive.idx,
		Metas: archive.metas,
	}
}

func (archive *Archive[FS]) tidyUp(metas backend.FileMetas) {
	selfMetas := byHashPath(archive.metas)
	originMetas := byHashPath(metas)
	fmt.Printf("SELF: %s\nORIGIN: %s\n", dump.Dump(selfMetas), dump.Dump(originMetas))
	for hash, metaMap := range selfMetas {
		if _, ok := originMetas[hash]; !ok {
			for path := range metaMap {
				archive.purgeFile(path)
			}
		}
	}
}

func byHashPath(metas backend.FileMetas) ByHashPath {
	result := ByHashPath{}
	for i := range metas {
		meta := &metas[i]
		if metaMap, ok := result[meta.Hash]; ok {
			metaMap[meta.Path] = meta
		} else {
			result[meta.Hash] = map[string]*backend.FileMeta{
				meta.Path: meta,
			}
		}
	}
	return result
}

func (archive *Archive[FS]) purgeFile(path string) {
	archive.backend.MoveFile(
		filepath.Join(archive.base, path),
		filepath.Join(archive.base, archive.purge, path),
	)
}

func (archive *Archive[FS]) shutdown() {
	fmt.Println("TODO: shutdown archive")
	archive.events <- Finished{Idx: archive.idx}
}
