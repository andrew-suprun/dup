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

type Files struct{}

func (f *Files) ScanFolder(path string) (backend.FolderMeta, error) {
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

func (f *Files) MoveFile(src, dest string) error {
	fmt.Printf("Moving %q to %q\n", src, dest)
	return nil
}
