package main

import "time"

func (app *app) ui() {
	for {
		if app.lc.ShoudStop() {
			return
		}
		// Draw UI
		time.Sleep(16 * time.Millisecond)
	}
}
