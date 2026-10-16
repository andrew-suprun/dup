package main

type fsRunner struct {
	path     string
	idx      int
	commands <-chan any
	events   chan<- any
	quit     <-chan struct{}
	backend  Backend
}

func runFs(path string, idx int, lc *Lifecycle, commands <-chan any, events chan<- any, quit <-chan struct{}, backend Backend) {
	runner := &fsRunner{
		path:     path,
		idx:      idx,
		commands: commands,
		events:   events,
		quit:     quit,
		backend:  backend,
	}
	go runner.run()
}

func (r *fsRunner) run() {
	for {
		select {
		case cmd := <-r.commands:
			r.handleCommand(cmd)
		case <-r.quit:
			r.events <- Quit{}
		}
	}
}

func (r *fsRunner) handleCommand(cmd any) {
	switch cmd := cmd.(type) {
	case LoadMetas:
		metas := r.backend.LoadMetas()
		r.events <- MetasLoaded{Idx: r.idx, Metas: metas}
	case StoreMetas:
		r.backend.StoreMetas(FileMetas(cmd))
		r.events <- MetasStored{Idx: r.idx}
	case ScanArchive:
		metas, err := r.backend.ScanArchive()
		if err != nil {
			r.events <- err
		}
		r.events <- MetasScanned{Idx: r.idx, Metas: metas}
	case HashFile:
		hash, err := r.backend.HashFile(cmd.Path)
		if err != nil {
			r.events <- err
		}
		r.events <- FileHashed{Idx: r.idx, Path: cmd.Path, Hash: hash}
	case MoveFile:
		err := r.backend.MoveFile(cmd.Src, cmd.Dest)
		if err != nil {
			r.events <- err
		}
		r.events <- FileMoved{Idx: r.idx, Src: cmd.Src, Dest: cmd.Dest}
	case OpenFile:
		err := r.backend.OpenFile(cmd.Path)
		if err != nil {
			r.events <- err
		}
		r.events <- FileOpened{Idx: r.idx, Path: cmd.Path}
	case ReadFile:
		size, err := r.backend.ReadFile(cmd.Buf)
		if err != nil {
			r.events <- err
		}
		r.events <- FileRead{Idx: r.idx, Size: size}
	case CreateFile:
		err := r.backend.CreateFile(cmd.Path)
		if err != nil {
			r.events <- err
		}
		r.events <- FileCreated{Idx: r.idx, Path: cmd.Path}
	case WriteFile:
		err := r.backend.WriteFile(cmd.Buf)
		if err != nil {
			r.events <- err
		}
		r.events <- FileWritten{Idx: r.idx}
	case CloseFile:
		r.backend.CloseFile()
		r.events <- FileClosed{Idx: r.idx}
	case RemoveFile:
		r.backend.RemoveFile(cmd.Path)
		r.events <- FileRemoved{Idx: r.idx, Path: cmd.Path}
	}
}
