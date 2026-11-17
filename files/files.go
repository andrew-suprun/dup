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

type FolderScanned struct {
	Idx            int
	FilesScanned   int
	FoldersScanned int
	FoldersToScan  int
}

type ArchiveScanned struct {
	Idx         int
	FolderMetas metas.FolderMetas
	FileMetas   metas.FileMetas
}

func ScanArchive(idx int, root string, events chan<- any, quit <-chan struct{}) {
	rootPrefixLen := len(root) + 1
	foldersToScan := []string{root}
	foldersScanned := 0
	result := ArchiveScanned{Idx: idx}

	for len(foldersToScan) > 0 {
		select {
		case <-quit:
			return
		default:
		}
		folderToScan := foldersToScan[0]
		foldersToScan = foldersToScan[1:]

		entries, err := os.ReadDir(folderToScan)
		if err != nil {
			events <- err
			return
		}
		for _, entry := range entries {
			childPath := filepath.Join(folderToScan, entry.Name())
			if entry.IsDir() {
				if !strings.HasPrefix(entry.Name(), "~~~") {
					data, err := xattr.Get(childPath, ".dup.meta.")
					record := strings.Split(string(data), ",")
					if len(record) == folderMetaFields && record[0] == folderMetaVersion && err == nil {
						result.FolderMetas = append(result.FolderMetas, metas.FolderMeta{
							Path: childPath[rootPrefixLen:],
							Id:   record[1],
						})
					} else {
						result.FolderMetas = append(result.FolderMetas, metas.FolderMeta{
							Path: childPath[rootPrefixLen:],
						})
					}
					foldersToScan = append(foldersToScan, childPath)
				}
			} else {
				if !entry.Type().IsRegular() || entry.Name() == ".DS_Store" || strings.HasPrefix(entry.Name(), "._") {
					continue
				}
				info, err := entry.Info()
				if err != nil {
					events <- err
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
					events <- err
					return
				}

				result.FileMetas = append(result.FileMetas, metas.FileMeta{
					Path:    norm.NFC.String(childPath),
					Size:    size,
					ModTime: modTime,
					Hash:    hash,
				})
			}
		}

		if err != nil {
			events <- err
			return
		}

		foldersScanned++
		events <- FolderScanned{
			Idx:            idx,
			FilesScanned:   len(result.FileMetas),
			FoldersScanned: foldersScanned,
			FoldersToScan:  len(foldersToScan),
		}
	}

	events <- result
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
