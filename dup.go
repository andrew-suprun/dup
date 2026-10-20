package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	lc := NewLifecycle()
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

	fss := []FS{}
	fss = make([]FS, 0, len(os.Args)-1)
	for idx, path := range os.Args[1:] {
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
		fss = append(fss, NewFS(path, idx, lc, events))
	}

	backup := time.Now().Format("~~~060102-150405~~~")
	AppState.Run(fss, backup, lc, events)
}

func handleInterrupts(lc *Lifecycle) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	signal.Reset()
	lc.Stop()
}
