package files

import (
	"crypto/sha256"
	"dup/metas"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/xattr"
	"golang.org/x/text/unicode/norm"
)

func ScanArchive(idx int, base string, events chan<- any, quit <-chan struct{}) {
	root := metas.NewFolder(nil, base)
	scanner := &Scanner{idx: idx, events: events, quit: quit}
	scanner.scan(root)
	events <- root
}

type Scanner struct {
	idx            int
	events         chan<- any
	quit           <-chan struct{}
	filesScanned   int
	foldersScanned int
	err            error
}

func (scanner *Scanner) scan(folder *metas.Folder) {
	if scanner.err != nil {
		scanner.events <- scanner.err
		return
	}
	select {
	case <-scanner.quit:
		return
	default:
	}

	folderPath := filepath.Join(folder.Path()...)
	data, err := xattr.Get(folderPath, ".dup.meta.")
	record := strings.Split(string(data), ",")
	if len(record) == folderMetaFields && record[0] == folderMetaVersion && err == nil {
		folder.Id = record[1]
	}

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		scanner.err = err
		return
	}
	for _, entry := range entries {
		childPath := filepath.Join(folderPath, entry.Name())
		if entry.IsDir() {
			if !strings.HasPrefix(entry.Name(), "~~~") {
				fmt.Println("folder", entry)
				childFolder := metas.NewFolder(folder, entry.Name())
				folder.Subfolders[entry.Name()] = childFolder
				scanner.scan(childFolder)
			}
		} else {
			if !entry.Type().IsRegular() || entry.Name() == ".DS_Store" || strings.HasPrefix(entry.Name(), "._") {
				continue
			}
			fmt.Println("file", entry)
			info, err := entry.Info()
			if err != nil {
				scanner.err = err
				return
			}

			size := info.Size()
			if size == 0 {
				continue
			}

			modTime := info.ModTime()
			modTime = modTime.UTC().Round(time.Second)

			hash, err := hashFile(childPath, info)
			if err != nil {
				scanner.err = err
				return
			}

			folder.Files[entry.Name()] = metas.NewFile(folder, norm.NFC.String(entry.Name()), size, modTime, hash)
			scanner.filesScanned++
		}
	}

	scanner.foldersScanned++
	scanner.events <- metas.FolderScanned{
		Idx:            scanner.idx,
		FilesScanned:   scanner.filesScanned,
		FoldersScanned: scanner.foldersScanned,
	}
}

const bufSize int64 = 256 * 1024
const folderMetaFields = 2
const folderMetaVersion = "v1"
const fileMetaFields = 4
const fileMetaVersion = "v1"

func hashFile(path string, info fs.FileInfo) (string, error) {
	data, err := xattr.Get(path, ".dup.meta.")
	if err == nil {
		record := strings.Split(string(data), ",")
		if len(record) == fileMetaFields {
			version := record[0]
			size, _ := strconv.ParseInt(record[1], 10, 64)
			modTime, _ := time.Parse(time.RFC3339, record[2])
			modTime = modTime.UTC().Round(time.Second)
			hash := record[3]
			if version == fileMetaVersion && size == info.Size() && modTime.Equal(info.ModTime()) && hash != "" {
				return hash, nil
			}
		}
	}

	hash := sha256.New()
	buf := make([]byte, bufSize)
	binary.LittleEndian.PutUint64(buf, uint64(info.Size()))
	hash.Write(buf[:8])

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	offset := bufSize
	if info.Size() > 2*bufSize {
		offset = info.Size() - bufSize
	}
	nr, er := file.Read(buf)
	if er != nil && er != io.EOF {
		return "", err
	}
	hash.Write(buf[:nr])
	if info.Size() > bufSize {
		nr, er := file.ReadAt(buf, offset)
		if er != nil && er != io.EOF {
			return "", er
		}
		hash.Write(buf[:nr])
	}

	strHash := base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
	meta := fmt.Sprintf("%s,%d,%s,%s", fileMetaVersion, info.Size(), info.ModTime().UTC().Format(time.RFC3339Nano), strHash)
	xattr.Set(path, ".dup.meta.", []byte(meta))
	return strHash, nil
}
