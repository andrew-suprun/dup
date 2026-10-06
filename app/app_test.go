package app

import (
	"dup/fs"
	"dup/fs/testfs"
	"dup/lifecycle"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("len", len(entries))
	events := make(chan any, 1)
	lc := lifecycle.New()
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".test") {
			continue
		}
		fmt.Println(entry)
		content, err := os.ReadFile(filepath.Join("testdata", entry.Name()))
		if err != nil {
			t.Error(err)
		}
		scripts := strings.Split(string(content), "===\n")
		fmt.Println("scripts", len(scripts))
		fss := []fs.FS{}
		for i, scriptStr := range scripts {
			script := strings.Split(scriptStr, "\n")
			fmt.Println("script", script)
			fs := testfs.New(i, t, script, events)
			fss = append(fss, fs)
		}
		Run(fss, "backup", lc, events)
	}
}
