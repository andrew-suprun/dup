package metas

import (
	"dup/dump"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestMetas(t *testing.T) {
	now := time.Now()
	root := &Folder{
		Name:       "/root",
		Id:         "1",
		Subfolders: Folders{},
		Files:      Files{},
	}
	level1 := root.AddSubfolder("level1")
	file := level1.AddFile("file1", 123, now, "hash")
	fmt.Println("path1", file.Path())
	if filepath.Join(file.Path()...) != "/root/level1/file1" {
		t.FailNow()
	}
	file.Move([]string{"a", "b", "c", "d"})
	fmt.Println("path2", file.Path())
	if filepath.Join(file.Path()...) != "/root/a/b/c/d" {
		t.FailNow()
	}
	file.Move([]string{"new", "file2"})
	fmt.Println("path3", file.Path())
	if filepath.Join(file.Path()...) != "/root/new/file2" {
		t.FailNow()
	}
	file.Move([]string{"level1", "file3"})
	fmt.Println("path4", file.Path())
	if filepath.Join(file.Path()...) != "/root/level1/file3" {
		t.FailNow()
	}
	recovered := assertPanics(t, func() {
		file.Move([]string{"level1"})
	})
	if recovered != "Cannot move File. A Folder with the same name exists." {
		t.FailNow()
	}
	fmt.Println(dump.Dump(recovered))
}

func assertPanics(t *testing.T, fn func()) (recovered any) {
	t.Helper()
	defer func() {
		recovered = recover()
		if recovered == nil {
			t.Errorf("expected function to panic, but it did not")
		}
	}()
	fn()
	return nil
}
