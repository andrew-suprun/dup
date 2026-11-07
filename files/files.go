package files

import (
	"crypto/sha256"
	"dup/backend"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/text/unicode/norm"
)

const hashFileName = ".meta.csv"
const bufSize int64 = 256 * 1024

type Files struct{}

func (f *Files) LoadMetas(base string) []byte {
	content, _ := os.ReadFile(filepath.Join(base, hashFileName))
	return content
}

func (f *Files) StoreMetas(base string, content []byte) {
	os.WriteFile(filepath.Join(base, hashFileName), content, 0644)
}

func (f *Files) ScanFolder(path string) (backend.FolderMeta, error) {
	result := backend.FolderMeta{}
	entries, err := os.ReadDir(path)
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		name := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			if !strings.HasPrefix(entry.Name(), "~~~") {
				result.Folders = append(result.Folders, name)
			}
		} else {
			if !entry.Type().IsRegular() || entry.Name() == hashFileName || entry.Name() == ".DS_Store" {
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

			sys := info.Sys().(*syscall.Stat_t)

			result.Files = append(result.Files, backend.FileMeta{
				Inode:   sys.Ino,
				Path:    norm.NFC.String(path),
				Size:    size,
				ModTime: modTime,
			})
		}
	}
	return result, nil
}

func (f *Files) HashFile(path string, size int64) (string, error) {
	hash := sha256.New()
	buf := make([]byte, bufSize)

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	offset := bufSize
	if size > 2*bufSize {
		offset = size - bufSize
	}
	nr, er := file.Read(buf)
	if er != nil && er != io.EOF {
		return "", err
	}
	hash.Write(buf[0:nr])
	if size > bufSize {
		nr, er := file.ReadAt(buf, offset)
		if er != nil && er != io.EOF {
			return "", er
		}
		hash.Write(buf[0:nr])
	}

	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil)), nil
}
