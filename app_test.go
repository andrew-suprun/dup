package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("len", len(entries))
	events := make(chan any, 1)
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".test") {
			continue
		}
		fmt.Println(entry)
		content, err := os.ReadFile(filepath.Join("testdata", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		scripts := strings.Split(string(content), "===\n")
		fmt.Println("scripts", len(scripts))
		fss := []FS{}
		for i, scriptStr := range scripts {
			script := strings.Split(scriptStr, "\n")
			fmt.Println("script", script)
			fs := New(i, t, script, events)
			fss = append(fss, fs)
		}
		Run(fss, "backup")
	}
}
