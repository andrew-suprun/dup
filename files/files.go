package files

import (
	"crypto/sha256"
	"dup/backend"
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

type ScannedFiles struct {
	Idx            int
	FilesScanned   int
	FoldersScanned int
	FoldersToScan  int
}

type ScannedArchive struct {
	Idx   int
	Metas backend.FileMetas
}

func ScanArchive(idx int, root string, events chan<- any, quit <-chan struct{}) {
	foldersToScan := []string{root}
	foldersScanned := 0
	metas := backend.FileMetas{}
	for len(foldersToScan) > 0 {
		select {
		case <-quit:
			return
		default:
		}
		folderToScan := foldersToScan[0]
		foldersToScan = foldersToScan[1:]
		folder, err := scanFolder(folderToScan)
		if err != nil {
			events <- err
			return
		}

		metas = append(metas, folder.Metas...)

		foldersScanned++
		foldersToScan = append(foldersToScan, folder.Folders...)
		events <- ScannedFiles{
			Idx:            idx,
			FilesScanned:   len(metas),
			FoldersScanned: foldersScanned,
			FoldersToScan:  len(foldersToScan),
		}
	}

	events <- ScannedArchive{
		Idx:   idx,
		Metas: metas,
	}
}

func scanFolder(path string) (backend.FolderMeta, error) {
	result := backend.FolderMeta{}
	entries, err := os.ReadDir(path)
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			if !strings.HasPrefix(entry.Name(), "~~~") {
				result.Folders = append(result.Folders, childPath)
			}
		} else {
			if !entry.Type().IsRegular() || entry.Name() == ".DS_Store" {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				return result, nil
			}

			size := info.Size()
			if size == 0 {
				return result, nil
			}

			modTime := info.ModTime()
			modTime = modTime.UTC().Round(time.Second)

			hash, err := hashFile(childPath, info)
			if err != nil {
				return result, nil
			}

			result.Metas = append(result.Metas, backend.FileMeta{
				Path:    norm.NFC.String(childPath),
				Size:    size,
				ModTime: modTime,
				Hash:    hash,
			})
		}
	}
	return result, nil
}

const bufSize int64 = 256 * 1024
const version int64 = 0

func hashFile(path string, info fs.FileInfo) (string, error) {
	data, err := xattr.Get(path, ".dup.meta.")
	if err == nil {
		record := strings.Split(string(data), ",")
		if len(record) == 4 {
			v, _ := strconv.ParseInt(record[0], 10, 64)
			size, _ := strconv.ParseInt(record[1], 10, 64)
			modTime, _ := time.Parse(time.RFC3339, record[2])
			modTime = modTime.UTC().Round(time.Second)
			hash := record[3]
			if v == version && size == info.Size() && modTime.Equal(info.ModTime()) && hash != "" {
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
	meta := fmt.Sprintf("%d,%d,%s,%s", version, info.Size(), info.ModTime().UTC().Format(time.RFC3339Nano), strHash)
	xattr.Set(path, ".dup.meta.", []byte(meta))
	return strHash, nil
}
