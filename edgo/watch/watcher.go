package watch

import (
	"bytes"
	"log"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Op describes a file operation on a watched object.
// In addition to the fsnotify events, an op mask of
// DirChild indicates that the event occurred in a watched
// directory rather than for a watched file itself.
type Op uint32

const (
	Create Op = 1 << iota
	Write
	Remove
	Rename
	Chmod
	DirChild
)

func (op Op) String() string {
	var buffer bytes.Buffer
	if op&Create == Create {
		buffer.WriteString("|CREATE")
	}
	if op&Remove == Remove {
		buffer.WriteString("|REMOVE")
	}
	if op&Write == Write {
		buffer.WriteString("|WRITE")
	}
	if op&Rename == Rename {
		buffer.WriteString("|RENAME")
	}
	if op&Chmod == Chmod {
		buffer.WriteString("|CHMOD")
	}
	if op&DirChild == DirChild {
		buffer.WriteString("|DIR_CHILD")
	}
	if buffer.Len() == 0 {
		return ""
	}
	return buffer.String()[1:] // Strip leading pipe
}

// Event describes a file change that affects `Name`
type Event struct {
	Name string
	Op   Op
}

// Watcher watches a directory or file and guards against
// multiple watching events by tracking the registered set,
// which fsnotify does not do.
type Watcher struct {
	Events   chan Event
	mux      sync.Mutex
	watchset map[string]struct{} // guarded by mux
	add      chan string
	remove   chan string
	err      chan error
}

func MakeWatcher() *Watcher {
	return &Watcher{
		watchset: make(map[string]struct{}),
		add:      make(chan string, 1),
		remove:   make(chan string, 1),
		err:      make(chan error, 1),
		Events:   make(chan Event, 1),
	}
}

// Close watcher related resources.
func (w *Watcher) Close() {
	close(w.add)
	close(w.remove)
	close(w.err)
	close(w.Events)
}

// RunLoop is the main watcher run loop; typically this is
// used inside a go routine.
func (w *Watcher) RunLoop(shutdown Shutdown) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		shutdown.Kill(err)
		return
	}
	defer watcher.Close()

	for {
		var evt fsnotify.Event
		var ok bool
		var err error

		select {
		case fname := <-w.add:
			w.err <- watcher.Add(fname)
			continue

		case fname := <-w.remove:
			w.err <- watcher.Remove(fname)
			continue

		case err, ok = <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
			continue

		case evt, ok = <-watcher.Events:
			if !ok {
				return
			}

		case <-shutdown.Dying():
			return
		}

		absName, err := filepath.Abs(evt.Name)
		dname := filepath.Dir(absName)

		w.mux.Lock()
		_, ok = w.watchset[absName]
		_, dok := w.watchset[dname]
		w.mux.Unlock()

		newevent := Event{
			Name: absName,
			Op:   Op(evt.Op),
		}

		if ok {
			// This entry existed in the watcher.
		} else if dok {
			// This entry parent existed in the watcher
			newevent.Op |= DirChild
		}

		select {
		case w.Events <- newevent:
			/*noop*/
		case <-shutdown.Dying():
			return
		}
	}
	panic("unreachable")
}

// AddWatch adds a watch to the watcher, if it is not already present.
func (w *Watcher) AddWatch(fname string) error {
	fname, err := filepath.Abs(fname)
	if err != nil {
		return err
	}

	w.mux.Lock()
	_, ok := w.watchset[fname]
	if ok {
		return nil
	}
	w.watchset[fname] = struct{}{}
	w.mux.Unlock()

	w.add <- fname
	return <-w.err
}

// RemoveWatch removes a watch from the watcher, unless it is not already present.
func (w *Watcher) RemoveWatch(fname string) error {
	fname, err := filepath.Abs(fname)
	if err != nil {
		return err
	}

	w.mux.Lock()
	_, ok := w.watchset[fname]
	if !ok {
		return nil
	}
	delete(w.watchset, fname)
	w.mux.Unlock()

	w.remove <- fname
	return <-w.err
}
