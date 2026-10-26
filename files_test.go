package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

type TestFs struct {
	idx      int
	t        *testing.T
	script   []any
	commands <-chan any
	events   chan<- any
	lineIdx  int
}

func runTestFs(idx int, t *testing.T, script []any, commands <-chan any, events chan<- any) {
	test := &TestFs{
		idx:      idx,
		script:   script,
		commands: commands,
		events:   events,
		t:        t,
	}
	test.run()
}

func (fsys *TestFs) run() {
}

func (fsys *TestFs) LoadMetas() FileMetas {
	return FileMetas{}
}

func (fsys *TestFs) StoreMetas(FileMetas) {

}

func (fsys *TestFs) ScanArchive() (FileMetas, error) {
	metas := []FileMeta{}
	for idx, line := range fsys.script {
		fmt.Printf("line [%d]: %q\n", fsys.idx, line)
		if line == "---" {
			fsys.lineIdx = idx + 1
			break
		}
		parts := strings.Split(line, "/")
		meta := FileMeta{
			Path: parts[0],
			Hash: parts[1],
		}
		if len(parts) == 3 {
			meta.Size, _ = strconv.Atoi(parts[2])
		}
		metas = append(metas, meta)
	}
	return metas, nil
}

func (fsys *TestFs) HashFile(path string) (string, error) {
	return "", nil
}

func (fsys *TestFs) MoveFile(dest, src string) error {
	received := fmt.Sprintf("mv: %s -> %s", src, dest)
	if fsys.lineIdx >= len(fsys.script) {
		fsys.t.Fatalf("Unexpected rename: %s", received)
	}
	expected := fsys.script[fsys.lineIdx]

	if expected == "ERROR" {
		return Error("ERROR")
	}

	if received != expected {
		fsys.t.Fatalf("expected %s | received %s\n", expected, received)
	}
	fsys.lineIdx++
	return nil
}

func (fsys *TestFs) OpenFile(path string) error {
	return nil
}

func (fsys *TestFs) ReadFile([]byte) error {
	return nil
}

func (fsys *TestFs) CreateFile(path string) error {
	return nil
}

func (fsys *TestFs) WriteFile([]byte) error {
	return nil
}

func (fsys *TestFs) CloseFile() error {
	return nil
}

func (fsys *TestFs) RemoveFile(path string) {

}

type Error string

func (e Error) Error() string {
	return string(e)
}
