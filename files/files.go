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

func ScanArchive(rootPath string, events chan<- any, quit <-chan struct{}) {
	root := &metas.Folder{Name: rootPath}
	scanner := &Scanner{rootPath: rootPath, events: events, quit: quit}
	scanner.scan(root)
	events <- root
}

type Scanner struct {
	rootPath       string
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

	folderPath := folder.Path()
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
				childFolder := &metas.Folder{
					Parent: folder,
					Name:   entry.Name(),
				}
				folder.Subfolders = append(folder.Subfolders, childFolder)
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

			folder.Files = append(folder.Files, &metas.File{
				Parent:  folder,
				Name:    norm.NFC.String(entry.Name()),
				Size:    size,
				ModTime: modTime,
				Hash:    hash,
			})
			scanner.filesScanned++
		}
	}

	scanner.foldersScanned++
	scanner.events <- metas.FolderScanned{
		Root:           scanner.rootPath,
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
	meta := fmt.Sprintf("%d,%d,%s,%s", fileMetaVersion, info.Size(), info.ModTime().UTC().Format(time.RFC3339Nano), strHash)
	xattr.Set(path, ".dup.meta.", []byte(meta))
	return strHash, nil
}
