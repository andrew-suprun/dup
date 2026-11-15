package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	suffexes := []string{", flac", ", m4a", ", m4b", ", mp3", ", mp4", ", opus", ",  flac", ",  m4a", ",  m4b", ",  mp3", ",  mp4", ",  opus"}
	paths := []string{os.Args[1]}

	i := 1
	for len(paths) > 0 {
		path := paths[0]
		paths = paths[1:]

		suffix := ""
		for _, s := range suffexes {
			if strings.HasSuffix(path, s) {
				suffix = s
				break
			}
		}

		if suffix != "" {
			fmt.Println(i, path)
			i++
			os.RemoveAll(path)
			continue
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			panic(err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				paths = append(paths, filepath.Join(path, entry.Name()))
			}
		}
	}
}
