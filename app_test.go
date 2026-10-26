package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRun(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("len", len(entries))
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".test") {
			continue
		}
		fmt.Println(entry)
		content, err := os.ReadFile(filepath.Join("testdata", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}

		scripts := make([]any, 0)
		err = yaml.Unmarshal(content, &scripts)
		if err != nil {
			panic(err)
		}

		commands := make([]chan any, len(scripts))
		events := make(chan any, len(scripts))

		for i, script := range scripts {
			go runTestFs(i, t, script.([]any), commands[i], events)
		}
		run(commands, events, uiEvents, quit)
	}
}

// func run(commands []chan<- any, events <-chan any, uiEvents chan<- any, quit <-chan struct{}) {
