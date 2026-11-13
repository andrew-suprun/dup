package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/xattr"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Provide folder path.")
		os.Exit(1)
	}
	pending := []string{os.Args[1]}
	scan(&pending)
}

func scan(paths *[]string) {
	for len(*paths) > 0 {
		path := (*paths)[0]
		*paths = (*paths)[1:]

		entries, err := os.ReadDir(path)
		if err != nil {
			panic(err)
		}
		for _, entry := range entries {
			childPath := filepath.Join(path, entry.Name())
			if entry.IsDir() {
				(*paths) = append((*paths), childPath)
			} else {
				info, err := entry.Info()
				if err != nil {
					panic(err)
				}
				hash, err := hash(childPath, info)
				if err != nil {
					panic(err)
				}
				fmt.Println(childPath, hash)
			}
		}
	}
}

const bufSize int64 = 256 * 1024

func hash(path string, info fs.FileInfo) (string, error) {
	data, err := xattr.Get(path, ".dup.meta.")
	if err == nil {
		record := strings.Split(string(data), ",")
		if len(record) == 3 {
			size, _ := strconv.ParseInt(record[0], 10, 64)
			modTime, _ := time.Parse(time.RFC3339, record[1])
			modTime = modTime.UTC().Round(time.Second)
			hash := record[2]
			if size == info.Size() && modTime.Equal(info.ModTime()) && hash != "" {
				return hash, nil
			}
		}
	}

	hash := sha256.New()
	buf := make([]byte, bufSize)

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
	hash.Write(buf[0:nr])
	if info.Size() > bufSize {
		nr, er := file.ReadAt(buf, offset)
		if er != nil && er != io.EOF {
			return "", er
		}
		hash.Write(buf[0:nr])
	}

	strHash := base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
	meta := fmt.Sprintf("%d,%s,%s", info.Size(), info.ModTime().UTC().Format(time.RFC3339Nano), strHash)
	xattr.Set(path, ".dup.meta.", []byte(meta))
	return strHash, nil
}
