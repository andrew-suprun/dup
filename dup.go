package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"dup/app"
	"dup/fs"
	"dup/fs/realfs"
	"dup/lifecycle"
)

func main() {
	lc := lifecycle.New()
	go handleInterrupts(lc)

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

	events := make(chan any, len(os.Args)-1)
	fss := []fs.FS{}
	fss = make([]fs.FS, 0, len(os.Args)-1)
	for idx, path := range os.Args[1:] {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			log.Printf("Failed to scan archives: %W\n", err)
			panic(err)
		}
		path, err := realfs.AbsPath(path)
		if err != nil {
			log.Printf("Failed to scan archives: %W\n", err)
			panic(err)
		}
		fss = append(fss, realfs.New(path, idx, lc, events))
	}

	backup := time.Now().Format("~~~060102-150405~~~")
	app.Run(fss, backup, lc, events)
}

func handleInterrupts(lc *lifecycle.Lifecycle) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	signal.Reset()
	lc.Stop()
}
