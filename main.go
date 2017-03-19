package main

import (
	"log"
	"path/filepath"

	"github.com/rjeczalik/notify"
    "os"
)

var F, _ = os.OpenFile("debug.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)

func main() {
	done := make(chan struct{})
	monitorCh := make(chan notify.EventInfo, 1)
	notifyCh := make(chan notify.EventInfo, 1)

    log.SetOutput(F)

	watchingPath := os.Args[1]

	if err := notify.Watch(filepath.Dir(watchingPath), monitorCh, notify.Create, notify.Remove); err != nil {
		log.Fatal(err)
	}
    log.Println("Watching " + watchingPath)
	defer notify.Stop(monitorCh)

	go CreateUI(done, notifyCh, watchingPath)

	for {
		select {
		case <-done:
			return
		case ei := <-monitorCh:
            log.Println(ei)
			notifyCh <- ei
		}
	}
}
