package archive

import (
	"dup/backend"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const hashFileName = ".meta.csv"

type Archive[B backend.Backend] struct {
	idx      int
	base     string
	purge    string
	events   chan<- any
	quit     chan struct{}
	commands <-chan any
	backend  B
	err      error
}

func RunArchive[B backend.Backend](idx int, base, purge string, events chan<- any, quit chan struct{}) <-chan any {
	var backend B
	archive := &Archive[B]{
		idx:      idx,
		base:     base,
		purge:    purge,
		events:   events,
		quit:     quit,
		commands: make(chan any),
		backend:  backend,
	}
	go archive.run()
	return archive.commands
}

func (archive *Archive[B]) run() {
mainLoop:
	for {
		select {
		case cmd := <-archive.commands:
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
	}
}

func (archive *Archive[FS]) scanArchive() {
	metasContent := archive.backend.LoadFile(filepath.Join(archive.base, hashFileName))
	loadedMetas := parseMeta(metasContent)

	metas := FileMetas{}
	filesHashed := 0
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
		for i := range folder.Files {
			backendMeta := &folder.Files[i]
			fileMeta := FileMeta{
				Inode:   backendMeta.Inode,
				Path:    backendMeta.Path,
				Size:    backendMeta.Size,
				ModTime: backendMeta.ModTime,
			}
			for j := range loadedMetas {
				loadedMeta := &loadedMetas[j]
				if loadedMeta.Inode == fileMeta.Inode {
					if loadedMeta.Size == fileMeta.Size && loadedMeta.ModTime.Equal(fileMeta.ModTime) && loadedMeta.Hash != "" {
						fileMeta.Hash = loadedMeta.Hash
						filesHashed++
					}
					break
				}
			}
			metas = append(metas, fileMeta)
		}
		foldersScanned++
		foldersToScan = append(foldersToScan, folder.Folders...)
		archive.events <- ScannedFiles{
			Idx:            archive.idx,
			FilesScanned:   len(metas),
			FilesHashed:    filesHashed,
			FoldersScanned: foldersScanned,
			FoldersToScan:  len(folderToScan),
		}
	}

	for i := range metas {
		meta := &metas[i]
		if meta.Hash == "" {
			hash, err := archive.backend.HashFile(meta.Path)
			if err != nil {
				archive.err = err
				archive.events <- archive.err
				close(archive.quit)
			}
			meta.Hash = hash
			filesHashed++
			archive.events <- ScannedFiles{
				Idx:            archive.idx,
				FilesScanned:   len(metas),
				FilesHashed:    filesHashed,
				FoldersScanned: foldersScanned,
				FoldersToScan:  0,
			}
		}
	}
	archive.events <- ScannedArchive{
		Idx:   archive.idx,
		Metas: metas,
	}
}

func (archive *Archive[FS]) shutdown() {
	fmt.Println("TODO: shutdown archive")
}

func parseMeta(content string) FileMetas {
	metas := make(FileMetas, 0)
	records, err := csv.NewReader(strings.NewReader(content)).ReadAll()
	if err != nil || len(records) == 0 {
		return metas
	}

	for _, record := range records[1:] {
		if len(record) == 5 {
			iNode, er1 := strconv.ParseUint(record[0], 10, 64)
			path := record[1]
			size, er2 := strconv.ParseUint(record[2], 10, 64)
			modTime, er3 := time.Parse(time.RFC3339, record[3])
			modTime = modTime.UTC().Round(time.Second)
			hash := record[4]
			if hash == "" || er1 != nil || er2 != nil || er3 != nil {
				continue
			}

			metas = append(metas, FileMeta{
				Inode:   iNode,
				Path:    path,
				Size:    int(size),
				ModTime: modTime,
				Hash:    hash,
			})
		}
	}
	return metas
}
