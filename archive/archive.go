package archive

import (
	"bytes"
	"dup/backend"
	"dup/dump"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/text/unicode/norm"
)

type Archive[B backend.Backend] struct {
	idx      int
	base     string
	purge    string
	events   chan<- any
	quit     chan struct{}
	commands <-chan any
	backend  B
	metas    FileMetas
	err      error

	shouldStoreMetas bool
}

type ByHashPath map[string]map[string]*FileMeta

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
	metasContent := archive.backend.LoadMetas(archive.base)
	loadedMetas := parseMeta(metasContent)

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
			archive.metas = append(archive.metas, fileMeta)
		}
		foldersScanned++
		foldersToScan = append(foldersToScan, folder.Folders...)
		archive.events <- ScannedFiles{
			Idx:            archive.idx,
			FilesScanned:   len(archive.metas),
			FilesHashed:    filesHashed,
			FoldersScanned: foldersScanned,
			FoldersToScan:  len(foldersToScan),
		}
	}

	for i := range archive.metas {
		meta := &archive.metas[i]
		if meta.Hash == "" {
			archive.shouldStoreMetas = true
			hash, err := archive.backend.HashFile(meta.Path, meta.Size)
			if err != nil {
				archive.err = err
				archive.events <- archive.err
				close(archive.quit)
			}
			meta.Hash = hash
			filesHashed++
			archive.events <- ScannedFiles{
				Idx:            archive.idx,
				FilesScanned:   len(archive.metas),
				FilesHashed:    filesHashed,
				FoldersScanned: foldersScanned,
				FoldersToScan:  0,
			}
		}
	}
	archive.events <- ScannedArchive{
		Idx:   archive.idx,
		Metas: archive.metas,
	}
}

func (archive *Archive[FS]) tidyUp(metas FileMetas) {
	selfMetas := byHashPath(archive.metas)
	originMetas := byHashPath(metas)
	fmt.Printf("SELF: %s\nORIGIN: %s\n", dump.Dump(selfMetas), dump.Dump(originMetas))

}

func byHashPath(metas FileMetas) ByHashPath {
	result := ByHashPath{}
	for i := range metas {
		meta := &metas[i]
		if metaMap, ok := result[meta.Hash]; ok {
			metaMap[meta.Path] = meta
		} else {
			result[meta.Hash] = map[string]*FileMeta{
				meta.Path: meta,
			}
		}
	}
	return result
}

func (archive *Archive[FS]) shutdown() {
	fmt.Println("TODO: shutdown archive")
	if archive.shouldStoreMetas {
		archive.backend.StoreMetas(archive.base, serializeMeta(archive.metas))
	}
	archive.events <- Finished{Idx: archive.idx}
}

func parseMeta(content []byte) FileMetas {
	metas := make(FileMetas, 0)
	records, err := csv.NewReader(bytes.NewReader(content)).ReadAll()
	if err != nil || len(records) == 0 {
		return metas
	}

	for _, record := range records[1:] {
		if len(record) == 5 {
			iNode, er1 := strconv.ParseUint(record[0], 10, 64)
			path := record[1]
			size, er2 := strconv.ParseInt(record[2], 10, 64)
			modTime, er3 := time.Parse(time.RFC3339, record[3])
			modTime = modTime.UTC().Round(time.Second)
			hash := record[4]
			if hash == "" || er1 != nil || er2 != nil || er3 != nil {
				continue
			}

			metas = append(metas, FileMeta{
				Inode:   iNode,
				Path:    path,
				Size:    size,
				ModTime: modTime,
				Hash:    hash,
			})
		}
	}
	return metas
}

func serializeMeta(metas FileMetas) []byte {
	result := make([][]string, 1, len(metas)+1)
	result[0] = []string{"INode", "Name", "Size", "ModTime", "Hash"}

	for _, meta := range metas {
		if meta.Hash == "" {
			continue
		}
		result = append(result, []string{
			fmt.Sprint(meta.Inode),
			norm.NFC.String(meta.Path),
			fmt.Sprint(meta.Size),
			meta.ModTime.UTC().Format(time.RFC3339Nano),
			meta.Hash,
		})
	}

	buffer := bytes.Buffer{}
	csv.NewWriter(&buffer).WriteAll(result)
	return buffer.Bytes()
}
