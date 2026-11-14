package app

import (
	"dup/analyzer"
	"dup/dump"
	"dup/files"
	"dup/metas"

	"golang.org/x/text/unicode/norm"

	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

type App struct {
	roots       []string
	purgeFolder string
	events      chan any
	quit        chan struct{}
	finished    []bool
}

func Run() {
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

	app := &App{
		purgeFolder: time.Now().Format("~~~060102-150405~~~"),
		events:      make(chan any, 2*len(os.Args)-1),
		quit:        make(chan struct{}),
		finished:    make([]bool, len(os.Args)-1),
	}

	for idx, path := range os.Args[1:] {
		fmt.Println("Idx", idx, "path", path)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			log.Printf("Failed to scan archives: %W\n", err)
			panic(err)
		}
		root, err := AbsPath(path)
		if err != nil {
			log.Printf("Failed to read root: $q: %W\n", path, err)
			panic(err)
		}
		app.roots = append(app.roots, root)
	}

	app.run()
}

func (app *App) run() {
	for idx, root := range app.roots {
		go files.ScanArchive(idx, root, app.events, app.quit)
	}
	scanned := make([]bool, len(app.roots))
	metas := make([]metas.FileMetas, len(app.roots))
	analyses := make([]analyzer.Analysis, len(app.roots))
scanLoop:
	for {
		event := <-app.events
		fmt.Printf("Event: %T %s\n", event, dump.Dump(event))
		switch event := event.(type) {
		case files.FolderScanned:
			// TODO UI
		case files.ArchiveScanned:
			metas[event.Idx] = event.Metas
			scanned[event.Idx] = true
			if event.Idx == 0 {
				for i := range app.roots[1:] {
					idx := i + 1
					if scanned[idx] {
						analyses[idx] = analyzer.Analyze(metas[0], metas[idx])
					}
				}
			} else {
				if scanned[0] {
					analyses[event.Idx] = analyzer.Analyze(metas[0], metas[event.Idx])
				}
			}
			allScanned := true
			for _, s := range scanned {
				if !s {
					allScanned = false
					break
				}
			}
			if allScanned {
				break scanLoop
			}

		case error:
			fmt.Println(event)
			close(app.quit)

		default:
			log.Panicf("### Unexpected Event: %T", event)
		}
	}

	fmt.Println(dump.Dump(analyses))
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
