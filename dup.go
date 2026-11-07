package main

import (
	"dup/app"
	"dup/archive"
	"dup/files"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"golang.org/x/text/unicode/norm"
)

func main() {
	events := make(chan any, 2*len(os.Args)-1)
	quit := make(chan struct{})
	go handleInterrupts(quit)

	logName := os.Getenv("DUP_LOG")
	if logName != "" {
		logFile, err := os.Create(logName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer logFile.Close()

		log.SetOutput(logFile)
		log.SetFlags(log.Lmicroseconds)
	} else {
		log.SetOutput(io.Discard)
	}

	purge := time.Now().Format("~~~060102-150405~~~")
	commands := make([]chan<- any, 0, len(os.Args)-1)
	for idx, path := range os.Args[1:] {
		fmt.Println("Idx", idx, "path", path)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			log.Printf("Failed to scan archives: %W\n", err)
			panic(err)
		}
		path, err := AbsPath(path)
		if err != nil {
			log.Printf("Failed to scan archives: %W\n", err)
			panic(err)
		}
		commandChan := archive.RunArchive[*files.Files](idx, path, purge, events, quit)
		commands = append(commands, commandChan)
	}

	app.Run(commands, events, quit)
}

func AbsPath(path string) (string, error) {
	var err error
	path, err = filepath.Abs(path)
	path = norm.NFC.String(path)
	if err != nil {
		return "", err
	}

	_, err = os.Stat(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func handleInterrupts(quit chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	signal.Reset()
	close(quit)
}
