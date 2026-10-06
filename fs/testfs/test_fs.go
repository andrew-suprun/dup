package testfs

import (
	"dup/fs"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

type TestFs struct {
	idx    int
	script []string
	events chan any
	t      *testing.T
	line   int
}

func New(idx int, t *testing.T, script []string, events chan any) *TestFs {
	return &TestFs{
		idx:    idx,
		script: script,
		events: events,
		t:      t,
	}
}

func (fsys *TestFs) Scan() {
	metas := []fs.FileMeta{}
	for _, line := range fsys.script {
		fmt.Printf("line [%d]: %q\n", fsys.idx, line)
		if line == "---" {
			break
		}
		parts := strings.Split(line, "/")
		meta := fs.FileMeta{
			Path: parts[0],
			Hash: parts[1],
		}
		if len(parts) == 3 {
			meta.Size, _ = strconv.Atoi(parts[2])
		}
		metas = append(metas, meta)
	}
	fsys.events <- fs.FileMetas{
		Idx:   fsys.idx,
		Metas: metas,
	}
	fsys.events <- fs.ArchiveHashed{Idx: fsys.idx}
}

func (fs *TestFs) Rename(dest, src string) error {
	fmt.Println("rename", src, "->", dest)
	expected := fs.script[fs.line]

	if expected == "ERROR" {
		return Error("ERROR")
	}

	received := fmt.Sprintf("mv: %s -> %s", src, dest)
	if received != expected {
		fs.t.Errorf("expected %q | received %q\n", expected, received)
	}
	fs.line++
	return nil
}

func (fs *TestFs) Copy(dest []string, src string) {

}

func (fs *TestFs) StoreMetas(fs.FileMetas) {}

type Error string

func (e Error) Error() string {
	return string(e)
}
