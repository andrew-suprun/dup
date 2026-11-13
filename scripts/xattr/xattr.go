package main

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/pkg/xattr"
)

func getXattrs(filePath string) {
	filepath.WalkDir(filePath, func(path string, d fs.DirEntry, err error) error {
		data, err := xattr.Get(path, ".dup.meta.")
		if err == nil {
			fmt.Printf("%q: %q\n", path, data)
		}

		return nil
	})

	// xattr.Set(filePath, ".dup.meta.", []byte("META"))

	// metas, err := xattr.List(filePath)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("----", filePath)
	// for _, meta := range metas {
	// 	data, err := xattr.Get(filePath, meta)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Printf("%q: %q\n", meta, data)
	// }
}

func main() {
	// getXattrs("./qwerty/abc.txt")
	// getXattrs("./qwerty/abc copy.txt")
	// getXattrs("./qwerty/abc copy 2.txt")
	getXattrs("..")
}
