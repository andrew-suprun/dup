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
	switch c := cmd.(type) {
	case LoadMetas:
		r.loadMetas()
	case StoreMetas:
		r.storeMetas(c)
	case ScanArchive:
		r.scanArchive()
	case HashFile:
		r.hashFile(c)
	case MoveFile:
		r.moveFile(c)
	}
}

func (r *fsRunner) loadMetas() {
	metas := r.backend.LoadMetas()
	r.events <- MetasLoaded{Idx: r.idx, Metas: metas}
}

func (r *fsRunner) storeMetas(metas StoreMetas) {
	r.backend.StoreMetas(FileMetas(metas))
	r.events <- MetasStored{Idx: r.idx}
}

func (r *fsRunner) scanArchive() {
	metas, err := r.backend.ScanArchive()
	if err != nil {
		r.events <- err
	}
	r.events <- MetasScanned{Idx: r.idx, Metas: metas}
}

func (r *fsRunner) hashFile(cmd HashFile) {
	hash, err := r.backend.HashFile(cmd.Path)
	if err != nil {
		r.events <- err
	}
	r.events <- FileHashed{Idx: r.idx, Path: cmd.Path, Hash: hash}
}

func (r *fsRunner) moveFile(cmd MoveFile) {
	err := r.backend.MoveFile(cmd.Src, cmd.Dest)
	if err != nil {
		r.events <- err
	}
	r.events <- FileMoved{Idx: r.idx, Src: cmd.Src, Dest: cmd.Dest}
}
