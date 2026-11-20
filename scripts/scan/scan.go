package main

import (
	"dup/dump"
	"dup/files"
	"dup/metas"
	"fmt"
	"os"
)

func main() {
	events := make(chan any, 1)
	quit := make(chan struct{})
	go func() {
		for {
			event := <-events
			fmt.Println(dump.Dump(event))
			if _, ok := event.(*metas.Folder); ok {
				quit <- struct{}{}
			}
		}
	}()
	files.ScanArchive(0, os.Args[1], events, quit)
	<-quit
}
