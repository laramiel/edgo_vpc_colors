package edgo

import (
	"errors"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"./watch"
)

var (
	journalRE     = regexp.MustCompile(`^journal[.][0-9]+[.][0-9]+.log$`)
	ErrEWShutdown = errors.New("elitewatcher: shutdown")
)

// EliteWatcher watches the Elite Dangerous journal directory
// for updates to journal entries, parses the json entries, and
// sends events of the parsed structures to the Journals channel.
//
// Under the covers, EliteWatcher runs several goroutines. One
// goroutine is dedicated to parsing the journal files and sending
// those events.
type EliteWatcher struct {
	DataDirectory string
	Journals      chan interface{}
	EventFilter   map[string]struct{}
	watcher       *watch.Watcher
	update        chan bool   // the existing journal has been updated
	newjournal    chan string // a new journal file is sent,
	statuswrite   chan string // the named status file has been updated
	tail          *watch.Tail
	shutdown      watch.Shutdown
}

func NewEliteWatcher(dirname string, shutdown watch.Shutdown) *EliteWatcher {
	return &EliteWatcher{
		DataDirectory: dirname,
		Journals:      make(chan interface{}, 10),
		watcher:       watch.MakeWatcher(),
		update:        make(chan bool, 1),
		newjournal:    make(chan string, 1),
		statuswrite:   make(chan string, 1),
		shutdown:      shutdown,
	}
}

func (ew *EliteWatcher) Close() {
	close(ew.Journals)
	close(ew.update)
	close(ew.newjournal)
	close(ew.statuswrite)
	ew.watcher.Close()
}

// setupInitialJournalFile scans the journal files in
// the existing directory and finds the most recent
// based on the journal filename.
func (ew *EliteWatcher) setupInitialJournalFile() error {
	files, err := ioutil.ReadDir(ew.DataDirectory)
	if err != nil {
		return err
	}

	var journalFiles []string
	for _, file := range files {
		base := filepath.Base(file.Name())
		if journalRE.MatchString(strings.ToLower(base)) {
			journalFiles = append(journalFiles, file.Name())
		}
	}
	if len(journalFiles) > 0 {
		sort.Strings(journalFiles)
		// name == basename
		journal, err := filepath.Abs(filepath.Join(ew.DataDirectory, journalFiles[len(journalFiles)-1]))
		if err != nil {
			return err
		}
		ew.maybeSetJournalFile(journal)
	}
	return nil
}

// fileTailer is the goroutine that is in charge of watching the
// journal file, reading, and parsing those files.
func (ew *EliteWatcher) fileTailer() {

	// Catch up with whatever status and journal file is set.
	for _, fname := range []string{"cargo.json", "market.json", "modulesinfo.json", "navroute.json", "outfitting.json", "shipyard.json", "status.json"} {
		if statusfile, err := filepath.Abs(filepath.Join(ew.DataDirectory, fname)); err == nil {
			ew.readAndParseStatusFile(statusfile)
		}
	}
	ew.tailJournalFile()

	// And loop forever receiving events.
	var fname string
	var ok bool
	for {
		select {
		case fname, ok = <-ew.newjournal:
			if !ok {
				return
			}
			ew.maybeSetJournalFile(fname)
			ew.tailJournalFile()

		case fname, ok = <-ew.statuswrite:
			if !ok {
				return
			}
			ew.readAndParseStatusFile(fname)

		case <-ew.update:
			ew.tailJournalFile()

		case <-ew.shutdown.Dying():
			return
		}
	}
	panic("unreachable")
}

func (ew *EliteWatcher) tailJournalFile() {
	if ew.tail == nil {
		// no tail file set, nothing to do
		return
	}
	ew.tail.ProcessLines(func(l string) error {
		b := []byte(l)

		// If ew.EventFilter is not empty, then check
		// whether the existing event is one we're interested in
		name := ""
		if len(ew.EventFilter) > 0 {
			name := GetEventNameByte(b)
			if name != "" {
				if _, ok := ew.EventFilter[name]; !ok {
					return nil
				}
			}
		}
		content, err := ParseJournalLine(b)
		if err == nil {
			if len(ew.EventFilter) > 0 && name == "" {
				// Apparently the byte-based event name filtering failed, so
				// filter based on the parsed representation.
				name := GetEventName(content)
				if name != "" {
					if _, ok := ew.EventFilter[name]; !ok {
						return nil
					}
				}
			}
			// Emit the event.
			select {
			case ew.Journals <- content:
				/*noop*/
			case <-ew.shutdown.Dying():
				return ErrEWShutdown
			}
		}
		return nil
	})
}

func (ew *EliteWatcher) readAndParseStatusFile(filename string) {
	// A status file is parsed directly by the watcher goroutine.
	// though we could change that.
	// TODO: Add event filter
	log.Println("status:", filename)
	content, err := ParseStatusData(filename)
	if err == nil {
		select {
		case <-ew.shutdown.Dying():
			return
		case ew.Journals <- content:
		}
	}
}

func (ew *EliteWatcher) maybeSetJournalFile(filename string) {
	if ew.tail != nil {
		current := strings.ToLower(filepath.Base(ew.tail.Filename))
		newfile := strings.ToLower(filepath.Base(filename))
		if strings.Compare(current, newfile) >= 0 {
			// same or newer file.
			return
		}
		ew.tail.Close()
	}

	tail, err := watch.TailFile(filename)
	if err != nil {
		log.Println("set journal: ", err, filename)
	} else {
		ew.tail = tail
		log.Println("set journal: ", ew.tail.Filename)
	}
}

// handleLoop is the goroutine that watches for changes to the
// journal directory and handles those change events.
func (ew *EliteWatcher) handleLoop() {
	ew.watcher.AddWatch(ew.DataDirectory)

	var event watch.Event
	var ok bool
	for {
		select {
		case <-ew.shutdown.Dying():
			return

		case event, ok = <-ew.watcher.Events:
			if !ok {
				return
			}
		}

		base := strings.ToLower(filepath.Base(event.Name))

		switch {
		case event.Op&watch.Write == watch.Write:
			if IsStatusFile(base) {
				// A status file was written. Block until the event is received.
				select {
				case <-ew.shutdown.Dying():
					return
				case ew.statuswrite <- event.Name:
					/*noop*/
				}
			} else if journalRE.MatchString(base) {
				// This is a journal file. Send an update if unblocked.
				select {
				case ew.update <- true:
				default:
				}
			} else {
				log.Println("file unknown", base)
			}
		case event.Op&watch.Create == watch.Create:
			if journalRE.MatchString(base) {
				// A new journal file was created. Block until the event is received.
				select {
				case <-ew.shutdown.Dying():
					return
				case ew.newjournal <- event.Name:
					/*noop*/
				}
			}
		}
	}
	panic("unreachable")
}

func (ew *EliteWatcher) Main() {
	ew.setupInitialJournalFile()

	go ew.fileTailer()
	go ew.handleLoop()

	ew.watcher.RunLoop(ew.shutdown)
}
