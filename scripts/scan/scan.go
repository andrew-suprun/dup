package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipFolder compresses sourceDir into targetZipPath
func ZipFolder(sourceDir, targetZipPath string) error {
	zipFile, err := os.Create(targetZipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	baseDir := filepath.Clean(sourceDir)

	return filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path so the zip doesn't store absolute local paths
		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Ensure cross-platform compatibility by converting backslashes to forward slashes
		header.Name = filepath.ToSlash(relPath)

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate // Apply standard DEFLATE compression
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

func main() {
	suffexes := []string{", flac", ", m4a", ", m4b", ", mp3", ", mp4", ", opus"}

	paths := []string{"/Volumes/SanDisk/Fiction/"}

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
			out := path[:len(path)-len(suffix)] + "." + suffix[2:] + ".zip"
			fmt.Println(out)
			ZipFolder(path, out)
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
